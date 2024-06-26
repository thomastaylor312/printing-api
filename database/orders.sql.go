// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: orders.sql

package database

import (
	"context"
	"database/sql"
	"time"
)

const createOrder = `-- name: CreateOrder :one
INSERT INTO orders (user_id, shipping_detail_id, created, external_order_id, payment_link, is_paid, order_status) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7) RETURNING id, user_id, shipping_detail_id, created, external_order_id, payment_link, is_paid, order_status
`

type CreateOrderParams struct {
	UserID           int64         `json:"userId"`
	ShippingDetailID sql.NullInt64 `json:"shippingDetailId"`
	Created          time.Time     `json:"created"`
	ExternalOrderID  string        `json:"externalOrderId"`
	PaymentLink      string        `json:"paymentLink"`
	IsPaid           bool          `json:"isPaid"`
	OrderStatus      string        `json:"orderStatus"`
}

func (q *Queries) CreateOrder(ctx context.Context, arg CreateOrderParams) (Order, error) {
	row := q.queryRow(ctx, q.createOrderStmt, createOrder,
		arg.UserID,
		arg.ShippingDetailID,
		arg.Created,
		arg.ExternalOrderID,
		arg.PaymentLink,
		arg.IsPaid,
		arg.OrderStatus,
	)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ShippingDetailID,
		&i.Created,
		&i.ExternalOrderID,
		&i.PaymentLink,
		&i.IsPaid,
		&i.OrderStatus,
	)
	return i, err
}

const deleteOrder = `-- name: DeleteOrder :exec
DELETE FROM orders WHERE id = ?1
`

func (q *Queries) DeleteOrder(ctx context.Context, id int64) error {
	_, err := q.exec(ctx, q.deleteOrderStmt, deleteOrder, id)
	return err
}

const getOrderForUser = `-- name: GetOrderForUser :one
SELECT id, user_id, shipping_detail_id, created, external_order_id, payment_link, is_paid, order_status FROM orders WHERE user_id = ?1 AND id = ?2
`

type GetOrderForUserParams struct {
	UserID int64 `json:"userId"`
	ID     int64 `json:"id"`
}

func (q *Queries) GetOrderForUser(ctx context.Context, arg GetOrderForUserParams) (Order, error) {
	row := q.queryRow(ctx, q.getOrderForUserStmt, getOrderForUser, arg.UserID, arg.ID)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ShippingDetailID,
		&i.Created,
		&i.ExternalOrderID,
		&i.PaymentLink,
		&i.IsPaid,
		&i.OrderStatus,
	)
	return i, err
}

const getOrdersForUser = `-- name: GetOrdersForUser :many
SELECT id, user_id, shipping_detail_id, created, external_order_id, payment_link, is_paid, order_status FROM orders WHERE user_id = ?1
`

func (q *Queries) GetOrdersForUser(ctx context.Context, userID int64) ([]Order, error) {
	rows, err := q.query(ctx, q.getOrdersForUserStmt, getOrdersForUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Order
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.ShippingDetailID,
			&i.Created,
			&i.ExternalOrderID,
			&i.PaymentLink,
			&i.IsPaid,
			&i.OrderStatus,
		); err != nil {
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

const updateOrder = `-- name: UpdateOrder :exec
UPDATE orders SET shipping_detail_id = ?1, created = ?2, external_order_id = ?3, payment_link = ?4, is_paid = ?5, order_status = ?6 WHERE id = ?7
`

type UpdateOrderParams struct {
	ShippingDetailID sql.NullInt64 `json:"shippingDetailId"`
	Created          time.Time     `json:"created"`
	ExternalOrderID  string        `json:"externalOrderId"`
	PaymentLink      string        `json:"paymentLink"`
	IsPaid           bool          `json:"isPaid"`
	OrderStatus      string        `json:"orderStatus"`
	ID               int64         `json:"id"`
}

func (q *Queries) UpdateOrder(ctx context.Context, arg UpdateOrderParams) error {
	_, err := q.exec(ctx, q.updateOrderStmt, updateOrder,
		arg.ShippingDetailID,
		arg.Created,
		arg.ExternalOrderID,
		arg.PaymentLink,
		arg.IsPaid,
		arg.OrderStatus,
		arg.ID,
	)
	return err
}

const updateOrderStatus = `-- name: UpdateOrderStatus :exec
UPDATE orders SET order_status = ?1 WHERE id = ?2
`

type UpdateOrderStatusParams struct {
	OrderStatus string `json:"orderStatus"`
	ID          int64  `json:"id"`
}

func (q *Queries) UpdateOrderStatus(ctx context.Context, arg UpdateOrderStatusParams) error {
	_, err := q.exec(ctx, q.updateOrderStatusStmt, updateOrderStatus, arg.OrderStatus, arg.ID)
	return err
}
