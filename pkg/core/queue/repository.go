package queue

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/generator"
)

type TaskRepository interface {
	// PendingLanes returns a list of lanes that have tasks with the status
	// 'pending' or 'running'.
	PendingLanes() ([]string, error)
	// ResetRunningTasks resets all tasks with the status 'running' back to
	// 'pending'.
	ResetRunningTasks() error
	// InsertIfNotInQueue adds a new task to the queue with the status
	// 'pending', and if a task with the same lane and payload already exists in
	// the queue with the status 'pending', it does not insert a duplicate.
	InsertIfNotInQueue(t Task) error
	// Dequeue fetches the next pending task for the specified lane and marks it
	// as 'running'. It returns the task and a boolean indicating whether a task
	// was found.
	Dequeue(lane string) (Task, bool)
	// UpdateTaskStatus updates the status of a task by its ID.
	UpdateTaskStatus(taskID string, success bool) error
}

type SQLTaskRepository struct {
	db *sqlx.DB
}

func NewSQLTaskRepository(db *sqlx.DB) TaskRepository {
	return &SQLTaskRepository{db: db}
}

func (r *SQLTaskRepository) PendingLanes() ([]string, error) {
	ctx := context.Background()

	queries := repository.NewQueries(r.db)
	lanes, err := queries.GetLanesByStatus(ctx, env.QueueTaskStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending lanes: %w", err)
	}

	return lanes, nil
}

func (r *SQLTaskRepository) ResetRunningTasks() error {
	ctx := context.Background()

	queries := repository.NewQueries(r.db)
	_, err := queries.UpdateQueueTaskStatusByStatus(ctx, repository.UpdateQueueTaskStatusByStatusParams{
		OldStatus: env.QueueTaskStatusRunning,
		NewStatus: env.QueueTaskStatusPending,
		UpdatedAt: generator.NowISO8601(),
	})
	if err != nil {
		return fmt.Errorf("failed to reset running tasks: %w", err)
	}

	return nil
}
func (r *SQLTaskRepository) InsertIfNotInQueue(t Task) error {
	ctx := context.Background()

	now := generator.NowISO8601()
	queries := repository.NewQueries(r.db)

	existingTasks, err := queries.GetLaneTaskByPayloadStatus(ctx, repository.GetLaneTaskByPayloadStatusParams{
		Lane:    t.Lane,
		Payload: t.Payload,
		Status:  env.QueueTaskStatusPending,
	})
	if err != nil {
		return fmt.Errorf("failed to check for existing task: %w", err)
	}

	if len(existingTasks) > 0 {
		return nil // Task already exists, skip insertion
	}

	err = queries.CreateQueueTask(ctx, repository.QueueTask{
		ID:        t.ID,
		Lane:      t.Lane,
		Payload:   t.Payload,
		Status:    env.QueueTaskStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}
	return nil
}

func (r *SQLTaskRepository) Dequeue(lane string) (Task, bool) {
	ctx := context.Background()

	queries := repository.NewQueries(r.db)
	tasks, err := queries.GetFirstTaskInLane(ctx, repository.GetFirstTaskInLaneParams{
		Lane:   lane,
		Status: env.QueueTaskStatusPending,
	})
	if err != nil {
		fmt.Printf("failed to get queue task: %v\n", err)
		return Task{}, false
	}
	if len(tasks) < 1 {
		return Task{}, false
	}

	task := tasks[0]
	err = queries.UpdateQueueTaskStatusByID(ctx, repository.UpdateQueueTaskStatusByIDParams{
		ID:        task.ID,
		Status:    env.QueueTaskStatusRunning,
		UpdatedAt: generator.NowISO8601(),
	})
	if err != nil {
		fmt.Printf("failed to update queue task status: %v\n", err)
		return Task{}, false
	}

	return Task{
		ID:      task.ID,
		Lane:    task.Lane,
		Payload: task.Payload,
	}, true
}

func (r *SQLTaskRepository) UpdateTaskStatus(taskID string, success bool) error {
	ctx := context.Background()

	status := env.QueueTaskStatusFailed
	if success {
		status = env.QueueTaskStatusSucceeded
	}

	queries := repository.NewQueries(r.db)
	err := queries.UpdateQueueTaskStatusByID(ctx, repository.UpdateQueueTaskStatusByIDParams{
		ID:        taskID,
		Status:    status,
		UpdatedAt: generator.NowISO8601(),
	})
	if err != nil {
		return fmt.Errorf("failed to update queue task status: %w", err)
	}

	return nil
}
