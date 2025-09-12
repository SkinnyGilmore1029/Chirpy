-- name: CreateChirp :one
INSERT INTO chirps (id, body, user_id)
 VALUES ($1, $2, $3)
 RETURNING *;

-- name: GetAllChirps :many
SELECT id, created_at, updated_at, body, user_id
 FROM chirps
 ORDER BY created_at ASC, id ASC;

-- name: GetChirp :one
SELECT id, created_at, updated_at, body, user_id
 FROM chirps
 WHERE id = $1;

-- name: RemoveChirp :exec
DELETE FROM chirps
 WHERE id = $1;