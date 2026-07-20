CREATE TABLE gs_account (
    id TEXT NOT NULL,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    language_code TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,

    PRIMARY KEY (id),
    UNIQUE (username)
);

CREATE INDEX idx_gs_account_username ON gs_account(username);