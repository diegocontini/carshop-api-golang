# Decisions

ADR-lite log of choices that aren't obvious from the code.

## D-001: `shopspring/decimal` for money, not `float64`

Postgres columns for money use `NUMERIC`. Go-side we use `github.com/shopspring/decimal`. The C# version used `double` for `Car.Price` (footgun) and `decimal` elsewhere. We pick decimal everywhere money lives so rounding can't drift across calculations.

## D-002: `users.role` as TEXT + CHECK, not int enum

C# stored the role as an int via `JsonStringEnumConverter`. Postgres has no native enum without a CREATE TYPE, and the JSON shape on the wire is a string anyway. TEXT + `CHECK (role IN ('admin', 'vendor'))` keeps the migration single-file, makes SQL inspection trivial, and round-trips with JSON without conversion at the SQL boundary. Domain-level conversion happens in `dto.RoleToDTO` / `dto.RoleFromDTO`.

## D-003: Goose migrations run on boot from inside `main`

There is no sidecar migration container. `server/main.go` calls `db.Migrate(ctx, pool, carshop.MigrationsFS, "migrations")` before serving traffic. Concurrent boots are safe because goose locks the `goose_db_version` table. The trade-off: a broken migration aborts the API instead of leaving it running on an old schema. That's the right failure mode here.

## D-004: Module root owns the embed.FS

`migrations/` lives at root (per project layout requirement). Go `//go:embed` cannot use `..` paths, so the embed.FS lives in `embed.go` at the module root (package `carshop`) and is passed into `db.Migrate`. The `db` package therefore takes `fs.FS` as a parameter, not a hard-coded directive.

## D-005: Bulk image/item updates via `unnest`

Two options to update N rows in one round trip:
1. `pgx.Batch` with N statements - still 1 round trip, but N statements.
2. Single `UPDATE ... FROM (SELECT unnest($1::bigint[]) AS id, unnest($2::text[]) AS url) ...` - 1 statement.

We use (2). It's a single executable plan, friendlier to the query planner, and the sqlc-generated signature is simpler (takes parallel slices).

## D-006: No interfaces for repositories

We could have introduced `UserRepository` etc. for theoretical mockability. We didn't. Services hold a concrete `*sqlc.Queries`. Tests use real Postgres via testcontainers, which is what production hits. If a future service grows an external HTTP dependency that benefits from a fake, introduce the interface there.

## D-007: testcontainers-go over docker-compose'd test DB

Trade-off: testcontainers spins a fresh container per `go test` run, which adds ~12s of first-run cost (image pull + boot). It's the right call here because:
- Tests are self-contained: `go test ./...` works with no manual setup.
- CI doesn't need a sidecar.
- Truncate-between-tests + per-suite container is cheap relative to per-test container.

## D-008: Omit `password` from GET responses

Documented contract drift from the C# version, where `User.Password` was echoed back. The Go `UserResponse` DTO does not have the field at all - removing the leak entirely instead of relying on a serializer tag.
