# Domain rules

## Commission

- Per order: at most ONE `vendor_comissions` row. Enforced by `UNIQUE (order_id)` on the table.
- Formula: `comissionAmount = (user.comissionPerSaleInPercent / 100) * order.total`.
- A vendor with `NULL` or `<= 0` percent yields a zero commission, with the row still inserted (vendor + total are recorded for reporting).
- Commission lives in a single transaction with the order create/update. If the vendor doesn't exist, the order rolls back too.
- C# would INSERT a new row on every PUT. Go uses `UpsertComissionByOrder` (ON CONFLICT (order_id) DO UPDATE). See `../bug-fixes-report/001-duplicate-commission-on-order-update.md`.

## Admin seeding

- `server/main.go` calls `users.SeedAdmin` after migrations.
- Implementation is `INSERT ... ON CONFLICT (username) DO NOTHING` - idempotent and race-free across concurrent boots.
- Seeded admin defaults to `comission_per_sale_in_percent = 3` for parity with the C# bootstrap, even though admin doesn't earn commissions in practice.

## Car images

- `images` is part of the Car DTO on POST/PUT. Each image carries an optional `id`.
- On PUT:
  - image with `id` matching an existing image of this car -> URL is updated
  - image with no `id` -> inserted as new
  - image with `id` not belonging to this car (or unknown) -> inserted as new (matches the surprising C# behaviour)
- **Images NOT present in the request are NOT deleted.** Same as C#. There is currently no delete-image endpoint. Document this when adding one.
- All image work happens in a single transaction with the car update.
- Updates use `BulkUpdateCarImages` via `unnest(bigint[], text[])` - constant number of round trips regardless of image count.

## Order items

- Same insert/update/no-delete semantics as car images.
- Bulk update uses `unnest` with 4 arrays (id, car_id, price, discount).

## Misspellings preserved on purpose

The C# domain uses "Comission" (single 'm') instead of "Commission". Everywhere on the wire we keep that misspelling: JSON field names (`comissionPerSaleInPercent`, `comissionAmount`, `comissionPercentage`), the endpoint path (`/api/v1/comission`), DB table names (`vendor_comissions`). Internal Go variable names sometimes use the correct spelling in prose/comments, but never on the wire.

If a future task is "fix the comission/commission typo", that is a coordinated breaking change across DB, API, and clients - not a quick rename.
