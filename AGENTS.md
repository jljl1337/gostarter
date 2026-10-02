# AGENTS.md

Go library (`pkg/`) for building monolith/sidecar backends, plus a full reference app (`examples/full/`).
Single module `github.com/jljl1337/gostarter`, Go 1.25. Dual-dialect: **SQLite and PostgreSQL**.

## Commands

```bash
go build ./...            # compile
golangci-lint run         # lint (v2 config; `default: standard`)
bash script/test.sh       # THE ONLY test suite (see below)
```

- **There are no Go unit tests.** `go test ./...` reports "no test files" for every package. Do not
  treat it as verification.
- `.golangci.yml` uses `default: standard` = errcheck, govet, ineffassign, staticcheck, unused.
  No formatter is enforced; match surrounding style (gofumpt-like).
- errcheck excludes `Tx.Rollback`, `DB.Close`, `Conn.Close` — so `defer tx.Rollback()` without
  error handling is the expected style.

### E2E tests (Playwright, `test/`)

`bash script/test.sh` **must be run from the repo root** (relative paths). It: builds
`full.out` from `examples/full/cmd/main.go`, starts Postgres via `test/pg.compose.yml`, launches the
app twice (`test/sqlite.env` → :3000, `test/pg.env` → :3001), runs `pnpm exec playwright test` in
`test/`, then tears down processes/compose and `rm -rf data/test`.

Prereqs: `docker` and pnpm (`pnpm install --frozen-lockfile` in `test/`, as CI does).

Focused run — start the server yourself first:

```bash
go build -o full.out examples/full/cmd/main.go
./full.out -env=test/sqlite.env &
cd test && pnpm exec playwright test --project=sqlite   # or --project=postgres
```

Gotcha: `reporter: 'html'` defaults to `open: 'on-failure'`, so a **failing** local run makes the
script appear to hang while it serves the report. Use `PLAYWRIGHT_HTML_OPEN=never`.

`test/tests/api.spec.ts` is one long stateful scenario (sign-up → owner → role changes → notes →
sign-out), so it only makes sense against a **fresh database**.

## Migrations (hand-rolled, not goose)

```bash
go run ./cmd/new-migration -name <name> [-dir internal/sql/migration]   # creates empty up/down pair
```

- Two sets: `pkg/shared/sql/migration` (core, `gs_`-prefixed tables) and the app's
  `internal/sql/migration`. Both are `//go:embed migration` and **must** be rooted at a directory
  literally named `migration` (the runner does `fs.ReadDir("migration")`).
- Core migrations run only with `WithGostarterMigration()`; app migrations run with `WithAppMigrations()`.
- **Both `.up.sql` and `.down.sql` must exist** for an ID. `LoadMigrations` only records a migration
  once it has seen the second file, so an orphan up file is silently skipped.
- Rollback = delete the files and restart. The runner compares applied-vs-embedded counts and
  applies/rolls back in one startup transaction.
- **Never edit an already-applied migration.** Applied statements are stored and compared; a mismatch
  aborts startup with "applied migration does not match embedded migration". Always add a new one.
- SQL must run on both engines: `TEXT`/`INTEGER` only (no `SERIAL`/`AUTOINCREMENT` — IDs are ULIDs
  generated in Go), timestamps as ISO8601 `TEXT`, and **sqlx named params (`:name`) only** — never
  `?` or `$1`.

## Architecture

Layering is strict: `transport` (HTTP) → `service` (logic, returns `*ServiceError` with `ErrorCode`)
→ `repository` (raw SQL consts + typed param structs). `examples/full/internal/*` mirrors these and
embeds/wraps core types (`Queries` embeds `repository.Queries`, `Account = repository.Account`) so
app queries reuse core ones.

Wiring is entirely via `server.Option` funcs (`pkg/core/server/option.go`) in
`examples/full/internal/server/server.go`. **Order matters**: `WithDB` must precede
`WithDefaultScheduler` (it reads `s.db`), `WithPort`/`WithUnixSocket` must precede `WithHttpServer`,
and `WithCustomRoleManager` must precede the handler options. `WithHttpServer()` is mandatory or
`Start()` errors. `Start()` also runs migrations, then scheduler, then queue, then serves.

Listening: `Start()` serves TCP on `PORT` unless `SOCKET_PATH` is set, in which case it serves
**only** on that unix socket (parent dir auto-created, stale file removed if nothing answers, live
socket refused). The socket file is unlinked on shutdown and chmod to `SOCKET_PERM` (octal string,
default `0666`) on listen. Neither `examples/full` nor the test env files touch these — TCP stays
the default. `test/playwright.config.ts` hardcodes `baseURL`, so socket mode has no e2e coverage.

HTTP routing:
- Handlers register on an inner mux mounted at `/api` via `http.StripPrefix`. The auth middleware's
  `publicRoutes` map and role gating therefore see paths **without** `/api`.
- Go 1.22 patterns are the convention: `mux.HandleFunc("PUT /notes/{id}", ...)`, read via `r.PathValue("id")`.
- **Role gating is by first path segment** (`/api/owner/...`, `/api/moderator/...`). Role order comes
  from `role.NewRoleManager(roles...)` — first arg is highest. Under-privileged requests get **404,
  not 403** (deliberate, hides route existence). Gating is skipped entirely when only one role exists.
- First account on a fresh DB gets the top role; all later accounts get the bottom role.
- Manual API cookbook: `examples/full/test.http`.

## Configuration

- Env is loaded once into package-level globals via `env.MustSetConstants*` (panics on bad values)
  **before** anything else — `log.SetCustomLoggerFromEnv` fails if `env.ConstantsSet` is false.
- Two naming modes: `MustSetConstantsWithPrefix` reads `GOSTARTER_PORT`; `MustSetConstantsWithoutPrefix`
  (used by `examples/full`) reads bare `PORT`. All env files here (`test/*.env`, root `.env`) are unprefixed.
- Every key also supports `<KEY>_FILE` to read the value from a file path (Docker secrets pattern).
- `LOG_LEVEL` is a raw `log/slog` level int (`-10` = debug; the test env files set `-10` to stay quiet).
- Root `.env` is gitignored; defaults must work without it (`DATA_DIR=data`, driver sqlite, port 3000).

## Conventions

- Commits are Conventional Commits: `feat:`, `fix:`, `refactor:`, `test:`, `ops:`, `docs:`.
  History has `(#NN)` PR suffixes from squash merges — don't hand-write those.
- Leave `examples/full/internal/transport/endopint_note.go` alone; the filename typo is pre-existing.
- Never commit local artifacts: `full.out`, `data/`, `test/node_modules/`, `test/playwright-report/`,
  `test/test-results/`.
