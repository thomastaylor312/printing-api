-- name: AddPrint :one
INSERT INTO prints (picture_id, paper_id, order_id, cart_id, width, height, border_size, crop_x, crop_y, cost, quantity) VALUES ($picture_id, $paper_id, $order_id, $cart_id, $width, $height, $border_size, $crop_x, $crop_y, $cost, $quantity) RETURNING *;

-- name: UpdatePrintQuantity :exec
UPDATE prints SET quantity = $quantity WHERE id = $id;

-- name: DeletePrint :exec
DELETE FROM prints WHERE id = $id;
