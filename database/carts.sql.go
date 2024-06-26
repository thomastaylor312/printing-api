// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: carts.sql

package database

import (
	"context"
)

const getCarts = `-- name: GetCarts :many
SELECT id, user_id FROM carts
`

func (q *Queries) GetCarts(ctx context.Context) ([]Cart, error) {
	rows, err := q.query(ctx, q.getCartsStmt, getCarts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Cart
	for rows.Next() {
		var i Cart
		if err := rows.Scan(&i.ID, &i.UserID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserCart = `-- name: GetUserCart :one
SELECT carts.id, carts.user_id, prints.id, prints.picture_id, prints.paper_id, prints.order_id, prints.cart_id, prints.width, prints.height, prints.border_size, prints.crop_x, prints.crop_y, prints.cost, prints.quantity FROM carts JOIN prints ON carts.id = prints.cart_id WHERE carts.user_id = ?1
`

type GetUserCartRow struct {
	Cart  Cart  `json:"cart"`
	Print Print `json:"print"`
}

func (q *Queries) GetUserCart(ctx context.Context, userID int64) (GetUserCartRow, error) {
	row := q.queryRow(ctx, q.getUserCartStmt, getUserCart, userID)
	var i GetUserCartRow
	err := row.Scan(
		&i.Cart.ID,
		&i.Cart.UserID,
		&i.Print.ID,
		&i.Print.PictureID,
		&i.Print.PaperID,
		&i.Print.OrderID,
		&i.Print.CartID,
		&i.Print.Width,
		&i.Print.Height,
		&i.Print.BorderSize,
		&i.Print.CropX,
		&i.Print.CropY,
		&i.Print.Cost,
		&i.Print.Quantity,
	)
	return i, err
}

const upsertCart = `-- name: UpsertCart :exec
;

INSERT INTO carts (user_id) VALUES (?1) ON CONFLICT(user_id) DO NOTHING
`

func (q *Queries) UpsertCart(ctx context.Context, userID int64) error {
	_, err := q.exec(ctx, q.upsertCartStmt, upsertCart, userID)
	return err
}
