-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users WHERE id = $id;

-- name: AddUser :one
INSERT INTO users (username, email, is_admin, created) VALUES ($username, $email, $is_admin, $created) RETURNING *;

-- name: UpdateUser :exec
UPDATE users SET username = $username, email = $email, is_admin = $is_admin, created = $created WHERE id = $id;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $id;
