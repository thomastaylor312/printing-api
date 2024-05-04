-- name: GetCarts :many
SELECT * FROM carts;

-- name: GetUserCart :one
SELECT sqlc.embed(carts), sqlc.embed(prints) FROM carts JOIN prints ON carts.id = prints.cart_id WHERE carts.user_id = $user_id ;

-- name: UpsertCart :exec
INSERT INTO carts (user_id) VALUES ($user_id) ON CONFLICT(user_id) DO NOTHING;
