-- name: GetCar :one
SELECT * FROM cars WHERE id = $1;

-- name: ListCars :many
SELECT * FROM cars ORDER BY id;

-- name: CreateCar :one
INSERT INTO cars (new, brand, model, year, price, color, km, description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateCar :one
UPDATE cars
SET new         = $2,
    brand       = $3,
    model       = $4,
    year        = $5,
    price       = $6,
    color       = $7,
    km          = $8,
    description = $9
WHERE id = $1
RETURNING *;

-- name: DeleteCar :execrows
DELETE FROM cars WHERE id = $1;

-- name: ListCarImagesByCarIDs :many
SELECT * FROM car_images WHERE car_id = ANY($1::bigint[]) ORDER BY id;

-- name: BulkInsertCarImages :copyfrom
INSERT INTO car_images (url, car_id) VALUES ($1, $2);

-- name: UpdateCarImage :exec
UPDATE car_images SET url = $2 WHERE id = $1 AND car_id = $3;

-- name: DeleteCarImagesByCarID :exec
DELETE FROM car_images WHERE car_id = $1;
