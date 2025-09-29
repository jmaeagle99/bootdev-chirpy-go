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
