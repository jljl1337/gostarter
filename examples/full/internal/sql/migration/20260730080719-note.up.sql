CREATE TABLE note (
    id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (account_id) REFERENCES gs_account(id) ON DELETE CASCADE
);