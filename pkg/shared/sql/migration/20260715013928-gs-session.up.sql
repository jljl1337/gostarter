CREATE TABLE gs_session (
    id TEXT NOT NULL,
    account_id TEXT,
    token TEXT NOT NULL,
    csrf_token TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,

    PRIMARY KEY (id),
    UNIQUE (token),
    FOREIGN KEY (account_id) REFERENCES gs_account(id) ON DELETE CASCADE
);

CREATE INDEX idx_gs_session_account_id ON gs_session(account_id);