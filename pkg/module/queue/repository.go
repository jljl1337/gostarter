package queue

import (
	"context"
	"fmt"

	"github.com/jljl1337/gostarter/pkg/repository"
)

type Task struct {
	ID             string `json:"id" db:"id"`
	Lane           string `json:"lane" db:"lane"`
	Payload        string `json:"payload" db:"payload"`
	Status         string `json:"status" db:"status"`
	FailedAttempts int    `json:"failedAttempts" db:"failed_attempts"`
	FailedCause    string `json:"failedCause" db:"failed_cause"`
	CreatedAt      string `json:"createdAt" db:"created_at"`
	UpdatedAt      string `json:"updatedAt" db:"updated_at"`
}

type NewRepository func(*repository.Database) (Repository, error)

type Repository interface {
	// TaskExist checks if a task with the same lane, payload, and status
	// already exists.
	TaskExist(ctx context.Context, task Task) (bool, error)

	// CreateTask persists a new pending task.
	CreateTask(ctx context.Context, task Task) error

	// SetTaskStatusByStatus updates the status of tasks from one to another.
	SetTaskStatusByStatus(ctx context.Context, fromStatus, toStatus, updatedAt string) error

	// GetLanesByStatus returns the distinct lanes for the given status.
	GetLanesByStatus(ctx context.Context, status string) ([]string, error)

	// NextTask returns the task for the given lane and status.
	NextTask(ctx context.Context, lane, status string) (*Task, error)

	// UpdateTask updates the task by its ID.
	UpdateTask(ctx context.Context, task Task) error
}

type defaultRepository struct {
	queryer *repository.Queryer
}

func newDefaultRepository(db *repository.Database) (Repository, error) {
	queryer, err := repository.NewQueryer(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create queryer: %w", err)
	}

	return &defaultRepository{
		queryer: queryer,
	}, nil
}

func (r *defaultRepository) TaskExist(ctx context.Context, task Task) (bool, error) {
	const statement = `
		SELECT
			COUNT(1)
		FROM
			gs_queue_task
		WHERE
			lane = :lane AND
			payload = :payload AND
			status = :status
	`

	type params struct {
		Lane    string `db:"lane"`
		Payload string `db:"payload"`
		Status  string `db:"status"`
	}

	p := params{
		Lane:    task.Lane,
		Payload: task.Payload,
		Status:  task.Status,
	}

	var count int
	err := r.queryer.NamedGetContext(ctx, &count, statement, p)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *defaultRepository) CreateTask(ctx context.Context, task Task) error {
	const statement = `
		INSERT INTO
			gs_queue_task (
				id,
				lane,
				payload,
				status,
				failed_attempts,
				failed_cause,
				created_at,
				updated_at
			) VALUES (
				:id,
				:lane,
				:payload,
				:status,
				:failed_attempts,
				:failed_cause,
				:created_at,
				:updated_at
			)
	`

	return r.queryer.NamedExecOneRowContext(ctx, statement, task)
}

func (r *defaultRepository) SetTaskStatusByStatus(ctx context.Context, fromStatus, toStatus, updatedAt string) error {
	const statement = `
		UPDATE
			gs_queue_task
		SET
			status = :to_status,
			updated_at = :updated_at
		WHERE
			status = :from_status
	`

	type params struct {
		FromStatus string `db:"from_status"`
		ToStatus   string `db:"to_status"`
		UpdatedAt  string `db:"updated_at"`
	}

	p := params{
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		UpdatedAt:  updatedAt,
	}

	_, err := r.queryer.NamedExecRowsAffectedContext(ctx, statement, p)
	return err
}

func (r *defaultRepository) GetLanesByStatus(ctx context.Context, status string) ([]string, error) {
	const statement = `
		SELECT
			DISTINCT lane
		FROM
			gs_queue_task
		WHERE
			status = :status
	`

	type params struct {
		Status string `db:"status"`
	}

	p := params{
		Status: status,
	}

	var lanes []string
	err := r.queryer.NamedSelectContext(ctx, &lanes, statement, p)

	return lanes, err
}

func (r *defaultRepository) NextTask(ctx context.Context, lane, status string) (*Task, error) {
	const statement = `
		SELECT
			*
		FROM
			gs_queue_task
		WHERE
			lane = :lane AND
			status = :status
		ORDER BY
			created_at ASC,
			id ASC
		LIMIT 1
	`

	type params struct {
		Lane   string `db:"lane"`
		Status string `db:"status"`
	}

	p := params{
		Lane:   lane,
		Status: status,
	}

	items := []Task{}
	err := r.queryer.NamedSelectContext(ctx, &items, statement, p)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil
	}

	return &items[0], nil
}

func (r *defaultRepository) UpdateTask(ctx context.Context, task Task) error {
	const statement = `
		UPDATE
			gs_queue_task
		SET
			status = :status,
			failed_attempts = :failed_attempts,
			failed_cause = :failed_cause,
			updated_at = :updated_at
		WHERE
			id = :id
	`

	return r.queryer.NamedExecOneRowContext(ctx, statement, task)
}
