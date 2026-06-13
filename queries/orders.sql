-- name: GetOrder :one
SELECT * FROM orders WHERE id = $1;

-- name: ListOrders :many
SELECT * FROM orders ORDER BY id;

-- name: ListOrdersByVendor :many
SELECT * FROM orders WHERE vendor_id = $1 ORDER BY id;

-- name: CreateOrder :one
INSERT INTO orders (customer_name, order_date, total, vendor_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateOrder :one
UPDATE orders
SET customer_name = $2,
    order_date    = $3,
    total         = $4,
    vendor_id     = $5
WHERE id = $1
RETURNING *;

-- name: DeleteOrder :execrows
DELETE FROM orders WHERE id = $1;

-- name: ListOrderItemsByOrderIDs :many
SELECT * FROM order_items WHERE order_id = ANY($1::bigint[]) ORDER BY id;

-- name: BulkInsertOrderItems :copyfrom
INSERT INTO order_items (order_id, car_id, price, discount) VALUES ($1, $2, $3, $4);

-- name: UpdateOrderItem :exec
UPDATE order_items
SET car_id = $2, price = $3, discount = $4
WHERE id = $1 AND order_id = $5;

-- name: DeleteOrderItemsByOrderID :exec
DELETE FROM order_items WHERE order_id = $1;
