CREATE TABLE gs_queue_task (
    id TEXT NOT NULL,
    lane TEXT NOT NULL,
    payload TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,

    PRIMARY KEY (id)
);