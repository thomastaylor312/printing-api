-- name: GetOrdersForUser :many
SELECT * FROM orders WHERE user_id = $user_id;

-- name: GetOrderForUser :one
SELECT * FROM orders WHERE user_id = $user_id AND id = $id;

-- name: CreateOrder :one
INSERT INTO orders (user_id, shipping_detail_id, created, external_order_id, payment_link, is_paid, order_status) VALUES ($user_id, $shipping_detail_id, $created, $external_order_id, $payment_link, $is_paid, $order_status) RETURNING *;

-- name: UpdateOrder :exec
UPDATE orders SET shipping_detail_id = $shipping_detail_id, created = $created, external_order_id = $external_order_id, payment_link = $payment_link, is_paid = $is_paid, order_status = $order_status WHERE id = $id;

-- name: UpdateOrderStatus :exec
UPDATE orders SET order_status = $order_status WHERE id = $id;

-- name: DeleteOrder :exec
DELETE FROM orders WHERE id = $id;
