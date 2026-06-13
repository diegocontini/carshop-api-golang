# Architecture

## Layers

```
HTTP request
   |
   v
controller/   -- gin handlers, DTOs, JWT middleware
   |
   v
service/      -- business rules, transactions, orchestration
   |
   v
domain/       -- pure structs, invariants, errors
   |
   v
infra/sqlc/   -- generated query code (do NOT hand-edit)
infra/db/     -- pgxpool + goose runner
```

## Dependency direction

- `controller` -> `service` -> `domain`
- `service` -> `infra/sqlc` (for queries)
- `domain` imports nothing internal. Pure types + small methods only.
- `controller/dto` imports `domain` (and `decimal`) but not `sqlc`.

If you find yourself importing `sqlc` from a controller, stop. Put the call behind a service method.

## Files per aggregate

For each aggregate (User, Car, Order, Comission) there is exactly one file in:
- `src/infra/queries/<name>.sql` (input to sqlc)
- `src/infra/sqlc/<name>.sql.go` (generated)
- `src/service/<name>_service.go` (business rules)
- `src/controller/<name>_controller.go` (HTTP)
- `src/controller/dto/<name>.go` (request/response DTOs + conversions)

Adding a 7th aggregate is the same pattern repeated.

## Why no interfaces by default

The C# original used DI containers and abstract base classes. We don't. A service is a concrete struct that takes `*pgxpool.Pool` and constructs its own `*sqlc.Queries`. We only introduce an interface when a test genuinely needs a fake - which integration tests don't (they hit a real Postgres via testcontainers).

If a future feature needs a mock (e.g. a service calls an external HTTP API and you want to stub it), define the interface where it's USED, not where it's implemented.

## Transactions

Writes that span multiple tables (order + items + commission, car + images) wrap the work in `pgx.BeginFunc`. Inside the callback we call `s.q.WithTx(tx)` to get a `*sqlc.Queries` scoped to the transaction. The pool is the only thing the service holds onto beyond construction.

## Embedded migrations

`migrations/` lives at the module root because Go `//go:embed` cannot reference paths above the source file. The embed.FS lives in `embed.go` at the module root (package `carshop`) and is passed into `db.Migrate` from `server/main.go`. The Makefile's `migrate-up` target uses the same `migrations/` folder via the `goose` CLI, so dev workflow and prod runtime agree.

## sqlc input vs output

- Input SQL lives in `src/infra/queries/<aggregate>.sql` (hand-written).
- Generated Go lives in `src/infra/sqlc/` (do not hand-edit).

Both are inside `src/infra/` because both are infrastructure concerns. They are kept in sibling folders rather than nested so the path of generated code stays short and the input/output split is visible at a glance.
