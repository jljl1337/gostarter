#!/usr/bin/env bash
set -euo pipefail

app_pids=()
start_pwd=$(pwd)

cleanup() {
    cd "$start_pwd"

    echo "Cleaning up application processes..."
	local exit_code=$?

    echo "Shutting down the database compose..."
	docker compose -f "test/pg.compose.yml" down

	set +e

    if [ ${#app_pids[@]} -eq 0 ]; then
        echo "No application processes to clean up."
        exit "$exit_code"
    fi

    echo "Stopping the application processes..."
	for pid in "${app_pids[@]}"; do
        echo "Killing application process with PID $pid..."
		kill "$pid" 2>/dev/null || true
	done

    echo "Waiting for application processes to exit..."
	for pid in "${app_pids[@]}"; do
        echo "Waiting for application process with PID $pid to exit..."
		wait "$pid" 2>/dev/null || true
	done

    echo "Cleanup complete. Exiting with code $exit_code."
	exit "$exit_code"
}

trap cleanup EXIT INT TERM

echo "Building the binary..."
go build -o full.out examples/full/cmd/main.go

echo "Running the database compose..."
docker compose -f test/pg.compose.yml up -d --wait

echo "Running the application..."

./full.out -env=test/sqlite.env > /dev/null 2>&1 &
sqlite_pid=$!
app_pids+=($sqlite_pid)
echo "Started the application with SQLite PID: $sqlite_pid"

./full.out -env=test/pg.env > /dev/null 2>&1 &
pg_pid=$!
app_pids+=($pg_pid)
echo "Started the application with PostgreSQL PID: $pg_pid"

cd test
echo "Running the tests..."
set +e
pnpm exec playwright test
set -e
cd ..