-- name: CreateUser :one
INSERT INTO users (
    email,
    password,
    role
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetUserByID :one
SELECT id, email, password, role, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users
SET
  email = COALESCE(NULLIF($2, ''), email),
  role = COALESCE(NULLIF($3, ''), role),
  updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, password, role, created_at, updated_at, deleted_at;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL LIMIT 1;

-- name: GetAllUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = now()
WHERE id = $1 AND deleted_at IS NULL;
