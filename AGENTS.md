# AGENTS.md

## Commands

Run `mise t` to list all available commands.

## Never read `.env`

The root `.env` is gitignored and may contain secrets. Never open, cat, or
grep it. Use `test/sqlite.env` / `test/pg.env` as reference env files instead.

