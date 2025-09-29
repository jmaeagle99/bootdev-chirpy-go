-- name: RegisterRefreshToken :one
INSERT INTO refresh_tokens (
    token,
    user_id,
    expires_at
) VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token = $1;
