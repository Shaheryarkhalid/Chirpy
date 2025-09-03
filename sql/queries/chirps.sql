-- name: CreateChirp :one
INSERT INTO chirps(id, created_at, updated_at, body, user_id)VALUES( gen_random_uuid(), Now(), Now(), $1, $2) returning *;

-- name: GetOneChirp :one
select * from chirps where id = $1 LIMIT 1;

-- name: GetAllChirps :many
select * from chirps order by created_at asc;

-- name: DeleteChirp :one
DELETE  from chirps where id = $1 returning *;

-- name: GetChirpsByAuthor :many
select * from chirps where user_id= $1 order by created_at asc;
