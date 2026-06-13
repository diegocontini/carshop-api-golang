# carshop-api-golang

Go port of `../CarShopApi` (C# / .NET 8). Same HTTP contract, better quality.

## Stack
- HTTP: `gin`
- DB: `pgx/v5` + `sqlc` (queries in `/queries`, generated code in `src/infra/sqlc`)
- Migrations: `goose` (SQL in `/migrations`, applied on boot from `main`)
- Auth: JWT HS256 + bcrypt passwords
- Money: `shopspring/decimal` (no float for money)

## Layout
```
server/           entrypoint
src/config/       Settings struct from env
src/controller/   gin handlers, DTOs, middleware
src/service/      business rules (one per aggregate)
src/domain/       pure entities + invariants
src/infra/db/     pgx pool + goose runner
src/infra/sqlc/   generated queries (do not hand-edit)
migrations/       goose .sql files
queries/          sqlc input .sql files
tests/            integration (testcontainers) + unit
docs/             HTTP contract (routes.md)
claude-memory/    cross-cutting context (read first for any non-trivial task)
bug-fixes-report/ one file per C# bug fixed during the port
```

Dependency direction: `controller → service → domain` and `service → infra/sqlc`. `domain` imports nothing internal.

## Commands
- `make run` — local dev (needs `.env` + a running Postgres)
- `make test` — unit + integration
- `make sqlc` — regenerate query code (after editing `/queries`)
- `make docker-up` / `make docker-down` — full stack via compose

## Read before working
- `claude-memory/README.md` — index
- `docs/routes.md` — HTTP contract (source of truth)
- `bug-fixes-report/README.md` — what was broken in C# and how we fixed it
