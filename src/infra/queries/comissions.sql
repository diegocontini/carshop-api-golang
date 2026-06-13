-- name: GetComission :one
SELECT * FROM vendor_comissions WHERE id = $1;

-- name: ListComissions :many
SELECT * FROM vendor_comissions ORDER BY id;

-- name: ListComissionsByVendor :many
SELECT * FROM vendor_comissions WHERE vendor_id = $1 ORDER BY id;

-- name: CreateComission :one
INSERT INTO vendor_comissions (
    vendor_id, vendor_name, comission_percentage, comission_amount, order_id, order_total
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateComission :one
UPDATE vendor_comissions
SET vendor_id            = $2,
    vendor_name          = $3,
    comission_percentage = $4,
    comission_amount     = $5,
    order_id             = $6,
    order_total          = $7
WHERE id = $1
RETURNING *;

-- name: DeleteComission :execrows
DELETE FROM vendor_comissions WHERE id = $1;

-- name: UpsertComissionByOrder :one
INSERT INTO vendor_comissions (
    vendor_id, vendor_name, comission_percentage, comission_amount, order_id, order_total
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (order_id) DO UPDATE
SET vendor_id            = EXCLUDED.vendor_id,
    vendor_name          = EXCLUDED.vendor_name,
    comission_percentage = EXCLUDED.comission_percentage,
    comission_amount     = EXCLUDED.comission_amount,
    order_total          = EXCLUDED.order_total
RETURNING *;
