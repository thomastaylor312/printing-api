-- name: GetPictures :many
SELECT * FROM pictures;

-- name: GetPicturesByUser :many
SELECT * FROM pictures WHERE user_id = $user_id;

-- name: AddPicture :one
INSERT INTO pictures (name, user_id) VALUES ($name, $user_id) RETURNING *;

-- name: DeletePicture :exec
DELETE FROM pictures WHERE id = $id;
