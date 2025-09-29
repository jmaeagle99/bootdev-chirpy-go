-- name: CreateUser :one
INSERT INTO users (
    email,
    hashed_password
)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserById :one
SELECT *
FROM users
WHERE users.id = $1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE users.email = $1;

-- name: GetUserByRefreshToken :one
SELECT users.*
FROM users
INNER JOIN refresh_tokens
ON users.id = refresh_tokens.user_id
WHERE
    refresh_tokens.token = $1 AND
    refresh_tokens.expires_at > now() AND
    refresh_tokens.revoked_at IS NULL
;

-- name: UpdateEmailAndPassword :one
UPDATE users
SET email = $2, hashed_password = $3
WHERE id = $1
RETURNING *;

-- name: UpgradeToRed :one
UPDATE users
SET is_chirpy_red = true
WHERE id = $1
RETURNING *;