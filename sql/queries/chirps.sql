-- name: CreateChirp :one
INSERT INTO chirps (
    body,
    user_id
) VALUES (
    $1,
    $2
)
RETURNING *;

-- name: DeleteAllChirps :exec
DELETE FROM chirps;

-- name: GetAllChirps :many
SELECT *
FROM chirps
ORDER BY chirps.created_at;

-- name: GetChirp :one
SELECT *
FROM chirps
WHERE id = $1;