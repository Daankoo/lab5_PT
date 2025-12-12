-- name: CreateUser :one
INSERT INTO users (name, email, age)
VALUES ($1, $2, $3)
RETURNING id, name, email, age;

-- name: GetUser :one
SELECT id, name, email, age
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, name, email, age
FROM users
ORDER BY id;

-- name: UpdateUser :one
UPDATE users
SET name = $2,
    email = $3,
    age = $4
WHERE id = $1
RETURNING id, name, email, age;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;