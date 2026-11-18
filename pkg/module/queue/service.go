package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jljl1337/gostarter/pkg/generator"
	"github.com/jljl1337/gostarter/pkg/log"
	"github.com/jljl1337/gostarter/pkg/repository"
)

const (
	taskStatusPending   = "pending"
	taskStatusRunning   = "running"
	taskStatusSucceeded = "succeeded"
	taskStatusFailed    = "failed"
)

const (
	DefaultMaxAttempts = 1
	DefaultBaseBackoff = time.Second
)

type Handler func(ctx context.Context, payload string) error

type Service struct {
	db            *repository.Database
	newRepository NewRepository
	newID         func() string
	mu            sync.Mutex
	lanes         map[string]*laneState
	running       bool
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

func NewService(db *repository.Database) *Service {
	return &Service{
		db:            db,
		newRepository: newDefaultRepository,
		lanes:         make(map[string]*laneState),
	}
}

// Subscribe registers the handler and retry policy of a lane.
// Call it once per lane, before Start.
func (s *Service) Subscribe(lane string, h Handler, opts ...LaneOption) error {
	if lane == "" {
		return fmt.Errorf("lane name cannot be empty")
	}

	if h == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	cfg := laneConfig{maxAttempts: DefaultMaxAttempts, baseBackoff: DefaultBaseBackoff}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.maxAttempts < 1 {
		cfg.maxAttempts = DefaultMaxAttempts
	}

	if cfg.baseBackoff < 0 {
		cfg.baseBackoff = DefaultBaseBackoff
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	l := s.laneLocked(lane)
	if l.handler != nil {
		return fmt.Errorf("queue: lane %s already has a subscriber", lane)
	}
	l.handler, l.cfg = h, cfg

	return nil
}

// Enqueue persists the task, then wakes (or starts) the lane's worker.
func (s *Service) Enqueue(ctx context.Context, lane, payload string) error {
	now := generator.NowISO8601()
	task := Task{
		ID:             s.newID(),
		Lane:           lane,
		Payload:        payload,
		Status:         taskStatusPending,
		FailedAttempts: 0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	repo, err := s.newRepository(s.db)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	exist, err := repo.TaskExist(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to check if task exists: %w", err)
	}

	if exist {
		return nil // task already exists, no need to enqueue again
	}

	if err := repo.CreateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	l := s.laneLocked(lane)
	l.hasWork = true
	s.startWorkerLocked(l)
	l.signal()
	return nil
}

// start starts workers for every lane that has leftover pending/running tasks.
// It only reads the lane names, tasks are fetched by the workers themselves.
func (s *Service) start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.running = true
	ctx := s.ctx
	s.mu.Unlock()

	repo, err := s.newRepository(s.db)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	now := generator.NowISO8601()

	err = repo.SetTaskStatusByStatus(ctx, taskStatusRunning, taskStatusPending, now)
	if err != nil {
		return fmt.Errorf("failed to reset running tasks to pending: %w", err)
	}

	names, err := repo.GetLanesByStatus(ctx, taskStatusPending)
	if err != nil {
		return fmt.Errorf("failed to get lanes with pending tasks: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, name := range names {
		l := s.laneLocked(name)
		l.hasWork = true
		if l.handler == nil {
			log.Warnf("Lane %s has pending tasks but no handler, skipping worker start", name)
			continue
		}
		s.startWorkerLocked(l)
	}

	return nil
}

// stop cancels all workers and waits for them to exit or for ctx to expire.
// Tasks interrupted by shutdown stay unfinished and resume on the next Start.
func (s *Service) stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.cancel()
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) laneLocked(name string) *laneState {
	l, ok := s.lanes[name]
	if !ok {
		l = &laneState{name: name, wake: make(chan struct{}, 1)}
		s.lanes[name] = l
	}
	return l
}

// startWorkerLocked starts the lane worker if the service is running, the lane
// has work and a handler, and no worker exists yet. s.mu must be held.
func (s *Service) startWorkerLocked(l *laneState) {
	if !s.running || l.started || !l.hasWork || l.handler == nil {
		return
	}
	l.started = true
	s.wg.Add(1)
	go s.work(s.ctx, l, l.handler, l.cfg)
}

// work processes one lane sequentially (FIFO). Storage is only read when the
// worker starts or when it is woken by Enqueue, never on a timer.
func (s *Service) work(ctx context.Context, l *laneState, h Handler, cfg laneConfig) {
	defer func() {
		s.mu.Lock()
		l.started = false
		s.mu.Unlock()
		s.wg.Done()
	}()

	for {
		if ctx.Err() != nil {
			return
		}

		repo, err := s.newRepository(s.db)
		if err != nil {
			log.Errorf("Failed to create repository for lane %s: %v", l.name, err)
			return
		}

		task, err := repo.NextTask(ctx, l.name, taskStatusPending)
		if err != nil {
			log.Errorf("Failed to fetch next task for lane %s: %v", l.name, err)
			task = nil // move on: wait for the next wake-up
		}

		if task == nil {
			select {
			case <-ctx.Done():
				return
			case <-l.wake:
			}
			continue
		}

		task.Status = taskStatusRunning
		task.UpdatedAt = generator.NowISO8601()
		if err := s.updateTask(ctx, task); err != nil {
			log.Errorf("Failed to mark task %s as running: %v", task.ID, err)
			// move on: the task will be retried on the next wake-up
			continue
		}

		s.process(ctx, task, h, cfg)
	}
}

// process runs one task until it succeeds, exhausts its attempts, or the
// service shuts down. Retries block the lane, which keeps ordering intact.
func (s *Service) process(ctx context.Context, t *Task, h Handler, cfg laneConfig) {
	// bookkeeping must still be persisted while shutting down
	persist := context.WithoutCancel(ctx)

	for {
		taskErr := safeCall(ctx, h, t.Payload)

		t.UpdatedAt = generator.NowISO8601()

		if taskErr == nil {
			t.Status = taskStatusSucceeded
			if err := s.updateTask(persist, t); err != nil {
				log.Errorf("Failed to mark task %s as succeeded: %v", t.ID, err)
			}
			return
		}

		if ctx.Err() != nil {
			// interrupted by shutdown: not a real failure, leave it unfinished
			return
		}

		t.FailedAttempts++
		if t.FailedAttempts >= cfg.maxAttempts {
			log.Errorf("Task %s failed after %d attempts: %v", t.ID, t.FailedAttempts, taskErr)

			// mark the task as failed, but don't retry anymore
			t.Status = taskStatusFailed
			t.FailedCause = taskErr.Error()
			if err := s.updateTask(persist, t); err != nil {
				log.Errorf("Failed to mark task %s as failed: %v", t.ID, err)
			}
			return
		}

		// retry: update the task and wait for the backoff delay
		if err := s.updateTask(persist, t); err != nil {
			log.Errorf("Failed to update task %s after failure: %v", t.ID, err)
		}

		delay := generator.Backoff(cfg.baseBackoff, t.FailedAttempts)
		log.Warnf("Task failed, retrying in %s (attempt %d/%d): %v", delay, t.FailedAttempts, cfg.maxAttempts, taskErr)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (s *Service) updateTask(ctx context.Context, t *Task) error {
	repo, error := s.newRepository(s.db)
	if error != nil {
		return fmt.Errorf("failed to create repository: %w", error)
	}

	if err := repo.UpdateTask(ctx, *t); err != nil {
		return fmt.Errorf("failed to update task %s: %w", t.ID, err)
	}

	return nil
}

func safeCall(ctx context.Context, h Handler, payload string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panic: %v", r)
		}
	}()
	return h(ctx, payload)
}
