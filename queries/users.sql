-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY id;

-- name: CreateUser :one
INSERT INTO users (username, password, email, comission_per_sale_in_percent, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET username                      = $2,
    password                      = $3,
    email                         = $4,
    comission_per_sale_in_percent = $5,
    role                          = $6
WHERE id = $1
RETURNING *;

-- name: DeleteUser :execrows
DELETE FROM users WHERE id = $1;

-- name: SeedAdmin :exec
INSERT INTO users (username, password, email, comission_per_sale_in_percent, role)
VALUES ($1, $2, $3, $4, 'admin')
ON CONFLICT (username) DO NOTHING;
