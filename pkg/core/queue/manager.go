package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/shared/generator"
	"github.com/jljl1337/gostarter/pkg/shared/log"
)

// QueueManager owns one worker goroutine per active lane, started lazily.
type QueueManager struct {
	repo     TaskRepository
	newID    func() string
	handlers map[string]TaskHandler // lane -> handler

	mu    sync.Mutex
	lanes map[string]*laneState // lane -> running worker state

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup // tracks all lane workers together
}

type laneState struct {
	wake chan struct{}
}

func NewDefaultQueueManager(db *sqlx.DB, laneList ...Lane) *QueueManager {
	return NewQueueManagerWithSQLRepo(db, generator.NewULID, laneList...)
}

func NewQueueManagerWithSQLRepo(db *sqlx.DB, newID func() string, laneList ...Lane) *QueueManager {
	repo := NewSQLTaskRepository(db)
	return NewQueueManager(repo, newID, laneList...)
}

func NewQueueManager(repo TaskRepository, newID func() string, laneList ...Lane) *QueueManager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &QueueManager{
		repo:     repo,
		newID:    newID,
		handlers: make(map[string]TaskHandler),
		lanes:    make(map[string]*laneState),
		ctx:      ctx,
		cancel:   cancel,
	}

	for _, lh := range laneList {
		m.handlers[lh.Name] = lh.TaskHandler
	}
	return m
}

// Resume starts workers for any lane that already has pending/running tasks
// left over from before a restart. Call once at startup.
func (m *QueueManager) Resume() error {
	err := m.repo.ResetRunningTasks()
	if err != nil {
		return fmt.Errorf("failed to reset running tasks: %w", err)
	}

	lanes, err := m.repo.PendingLanes()
	if err != nil {
		return fmt.Errorf("failed to get pending lanes: %w", err)
	}

	for _, ln := range lanes {
		m.ensureLaneRunning(ln)
	}

	return nil
}

// Enqueue inserts a task into the given lane and activates that lane's
// worker if it isn't already running.
func (m *QueueManager) Enqueue(lane, payload string) error {
	if _, ok := m.handlers[lane]; !ok {
		return fmt.Errorf("no handler registered for lane %q", lane)
	}

	task := Task{
		ID:      m.newID(),
		Lane:    lane,
		Payload: payload,
	}
	if err := m.repo.InsertIfNotInQueue(task); err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	l := m.ensureLaneRunning(lane)
	select {
	case l.wake <- struct{}{}:
	default:
	}

	return nil
}

// ensureLaneRunning lazily creates + starts a worker goroutine for laneName
// the first time it's needed, and is a no-op if it's already running.
func (m *QueueManager) ensureLaneRunning(lane string) *laneState {
	m.mu.Lock()
	defer m.mu.Unlock()

	if l, ok := m.lanes[lane]; ok {
		return l
	}

	l := &laneState{wake: make(chan struct{}, 1)}
	m.lanes[lane] = l

	m.wg.Add(1)
	go m.runLane(lane, l)

	return l
}

func (m *QueueManager) runLane(lane string, l *laneState) {
	defer m.wg.Done()
	handler := m.handlers[lane]

	for {
		if m.ctx.Err() != nil {
			return
		}

		task, ok := m.repo.Dequeue(lane)
		if ok {
			err := handler(task.Payload)
			if err != nil {
				log.Errorf("queue task %s failed: %v", task.ID, err)
			}

			err = m.repo.UpdateTaskStatus(task.ID, err == nil)
			if err != nil {
				log.Errorf("failed to update task status: %v", err)
			}
			continue
		}

		select {
		case <-l.wake:
			continue
		case <-m.ctx.Done():
			return
		}
	}
}

func (m *QueueManager) Shutdown(ctx context.Context) error {
	m.cancel()

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
