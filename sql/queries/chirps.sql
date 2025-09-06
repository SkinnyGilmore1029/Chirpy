-- name: CreateChirp :one
INSERT INTO chirps (id, body, user_id)
 VALUES ($1, $2, $3)
 RETURNING *;

-- name: GetAllChirps :many
SELECT id, created_at, updated_at, body, user_id
 FROM chirps
 ORDER BY created_at ASC, id ASC;