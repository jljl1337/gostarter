package repository

import "context"

const createQueueTask = `
	INSERT INTO gs_queue_task (
		id,
		lane,
		payload,
		status,
		created_at,
		updated_at
	) VALUES (
		:id,
		:lane,
		:payload,
		:status,
		:created_at,
		:updated_at
	)
`

func (q *Queries) CreateQueueTask(ctx context.Context, arg QueueTask) error {
	return q.NamedExecOneRowContext(ctx, createQueueTask, arg)
}

const getLanesByStatus = `
	SELECT
		DISTINCT lane
	FROM
		gs_queue_task
	WHERE
		status = :status
`

type GetLanesByStatusParams struct {
	Status string `db:"status"`
}

func (q *Queries) GetLanesByStatus(ctx context.Context, status string) ([]string, error) {
	lanes := []string{}

	err := q.NamedSelectContext(ctx, &lanes, getLanesByStatus, GetLanesByStatusParams{Status: status})

	return lanes, err
}

const getFirstTaskInLane = `
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

type GetFirstTaskInLaneParams struct {
	Lane   string `db:"lane"`
	Status string `db:"status"`
}

func (q *Queries) GetFirstTaskInLane(ctx context.Context, arg GetFirstTaskInLaneParams) ([]QueueTask, error) {
	items := []QueueTask{}

	err := q.NamedSelectContext(ctx, &items, getFirstTaskInLane, arg)

	return items, err
}

const getLaneTaskByPayloadStatus = `
	SELECT
		*
	FROM
		gs_queue_task
	WHERE
		lane = :lane AND
		payload = :payload AND
		status = :status
`

type GetLaneTaskByPayloadStatusParams struct {
	Lane    string `db:"lane"`
	Payload string `db:"payload"`
	Status  string `db:"status"`
}

func (q *Queries) GetLaneTaskByPayloadStatus(ctx context.Context, arg GetLaneTaskByPayloadStatusParams) ([]QueueTask, error) {
	items := []QueueTask{}

	err := q.NamedSelectContext(ctx, &items, getLaneTaskByPayloadStatus, arg)

	return items, err
}

const updateQueueTaskStatusByStatus = `
	UPDATE
		gs_queue_task
	SET
		status = :new_status,
		updated_at = :updated_at
	WHERE
		lane = :lane AND
		status = :old_status
`

type UpdateQueueTaskStatusByStatusParams struct {
	Lane      string `db:"lane"`
	OldStatus string `db:"old_status"`
	NewStatus string `db:"new_status"`
	UpdatedAt string `db:"updated_at"`
}

func (q *Queries) UpdateQueueTaskStatusByStatus(ctx context.Context, arg UpdateQueueTaskStatusByStatusParams) (int64, error) {
	return q.NamedExecRowsAffectedContext(ctx, updateQueueTaskStatusByStatus, arg)
}

const updateQueueTaskStatusByID = `
	UPDATE
		gs_queue_task
	SET
		status = :status,
		updated_at = :updated_at
	WHERE
		id = :id
`

type UpdateQueueTaskStatusByIDParams struct {
	ID        string `db:"id"`
	Status    string `db:"status"`
	UpdatedAt string `db:"updated_at"`
}

func (q *Queries) UpdateQueueTaskStatusByID(ctx context.Context, arg UpdateQueueTaskStatusByIDParams) error {
	return q.NamedExecOneRowContext(ctx, updateQueueTaskStatusByID, arg)
}
