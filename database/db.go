// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package database

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.addPaperStmt, err = db.PrepareContext(ctx, addPaper); err != nil {
		return nil, fmt.Errorf("error preparing query AddPaper: %w", err)
	}
	if q.addPictureStmt, err = db.PrepareContext(ctx, addPicture); err != nil {
		return nil, fmt.Errorf("error preparing query AddPicture: %w", err)
	}
	if q.addPrintStmt, err = db.PrepareContext(ctx, addPrint); err != nil {
		return nil, fmt.Errorf("error preparing query AddPrint: %w", err)
	}
	if q.addUserStmt, err = db.PrepareContext(ctx, addUser); err != nil {
		return nil, fmt.Errorf("error preparing query AddUser: %w", err)
	}
	if q.createOrderStmt, err = db.PrepareContext(ctx, createOrder); err != nil {
		return nil, fmt.Errorf("error preparing query CreateOrder: %w", err)
	}
	if q.deleteOrderStmt, err = db.PrepareContext(ctx, deleteOrder); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteOrder: %w", err)
	}
	if q.deletePaperStmt, err = db.PrepareContext(ctx, deletePaper); err != nil {
		return nil, fmt.Errorf("error preparing query DeletePaper: %w", err)
	}
	if q.deletePictureStmt, err = db.PrepareContext(ctx, deletePicture); err != nil {
		return nil, fmt.Errorf("error preparing query DeletePicture: %w", err)
	}
	if q.deletePrintStmt, err = db.PrepareContext(ctx, deletePrint); err != nil {
		return nil, fmt.Errorf("error preparing query DeletePrint: %w", err)
	}
	if q.deleteUserStmt, err = db.PrepareContext(ctx, deleteUser); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteUser: %w", err)
	}
	if q.getCartsStmt, err = db.PrepareContext(ctx, getCarts); err != nil {
		return nil, fmt.Errorf("error preparing query GetCarts: %w", err)
	}
	if q.getOrderForUserStmt, err = db.PrepareContext(ctx, getOrderForUser); err != nil {
		return nil, fmt.Errorf("error preparing query GetOrderForUser: %w", err)
	}
	if q.getOrdersForUserStmt, err = db.PrepareContext(ctx, getOrdersForUser); err != nil {
		return nil, fmt.Errorf("error preparing query GetOrdersForUser: %w", err)
	}
	if q.getPapersStmt, err = db.PrepareContext(ctx, getPapers); err != nil {
		return nil, fmt.Errorf("error preparing query GetPapers: %w", err)
	}
	if q.getPicturesStmt, err = db.PrepareContext(ctx, getPictures); err != nil {
		return nil, fmt.Errorf("error preparing query GetPictures: %w", err)
	}
	if q.getPicturesByUserStmt, err = db.PrepareContext(ctx, getPicturesByUser); err != nil {
		return nil, fmt.Errorf("error preparing query GetPicturesByUser: %w", err)
	}
	if q.getUserStmt, err = db.PrepareContext(ctx, getUser); err != nil {
		return nil, fmt.Errorf("error preparing query GetUser: %w", err)
	}
	if q.getUserCartStmt, err = db.PrepareContext(ctx, getUserCart); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserCart: %w", err)
	}
	if q.getUsersStmt, err = db.PrepareContext(ctx, getUsers); err != nil {
		return nil, fmt.Errorf("error preparing query GetUsers: %w", err)
	}
	if q.updateOrderStmt, err = db.PrepareContext(ctx, updateOrder); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateOrder: %w", err)
	}
	if q.updateOrderStatusStmt, err = db.PrepareContext(ctx, updateOrderStatus); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateOrderStatus: %w", err)
	}
	if q.updatePaperStmt, err = db.PrepareContext(ctx, updatePaper); err != nil {
		return nil, fmt.Errorf("error preparing query UpdatePaper: %w", err)
	}
	if q.updatePrintQuantityStmt, err = db.PrepareContext(ctx, updatePrintQuantity); err != nil {
		return nil, fmt.Errorf("error preparing query UpdatePrintQuantity: %w", err)
	}
	if q.updateUserStmt, err = db.PrepareContext(ctx, updateUser); err != nil {
		return nil, fmt.Errorf("error preparing query UpdateUser: %w", err)
	}
	if q.upsertCartStmt, err = db.PrepareContext(ctx, upsertCart); err != nil {
		return nil, fmt.Errorf("error preparing query UpsertCart: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.addPaperStmt != nil {
		if cerr := q.addPaperStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addPaperStmt: %w", cerr)
		}
	}
	if q.addPictureStmt != nil {
		if cerr := q.addPictureStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addPictureStmt: %w", cerr)
		}
	}
	if q.addPrintStmt != nil {
		if cerr := q.addPrintStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addPrintStmt: %w", cerr)
		}
	}
	if q.addUserStmt != nil {
		if cerr := q.addUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addUserStmt: %w", cerr)
		}
	}
	if q.createOrderStmt != nil {
		if cerr := q.createOrderStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createOrderStmt: %w", cerr)
		}
	}
	if q.deleteOrderStmt != nil {
		if cerr := q.deleteOrderStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteOrderStmt: %w", cerr)
		}
	}
	if q.deletePaperStmt != nil {
		if cerr := q.deletePaperStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deletePaperStmt: %w", cerr)
		}
	}
	if q.deletePictureStmt != nil {
		if cerr := q.deletePictureStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deletePictureStmt: %w", cerr)
		}
	}
	if q.deletePrintStmt != nil {
		if cerr := q.deletePrintStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deletePrintStmt: %w", cerr)
		}
	}
	if q.deleteUserStmt != nil {
		if cerr := q.deleteUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteUserStmt: %w", cerr)
		}
	}
	if q.getCartsStmt != nil {
		if cerr := q.getCartsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getCartsStmt: %w", cerr)
		}
	}
	if q.getOrderForUserStmt != nil {
		if cerr := q.getOrderForUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getOrderForUserStmt: %w", cerr)
		}
	}
	if q.getOrdersForUserStmt != nil {
		if cerr := q.getOrdersForUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getOrdersForUserStmt: %w", cerr)
		}
	}
	if q.getPapersStmt != nil {
		if cerr := q.getPapersStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPapersStmt: %w", cerr)
		}
	}
	if q.getPicturesStmt != nil {
		if cerr := q.getPicturesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPicturesStmt: %w", cerr)
		}
	}
	if q.getPicturesByUserStmt != nil {
		if cerr := q.getPicturesByUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getPicturesByUserStmt: %w", cerr)
		}
	}
	if q.getUserStmt != nil {
		if cerr := q.getUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserStmt: %w", cerr)
		}
	}
	if q.getUserCartStmt != nil {
		if cerr := q.getUserCartStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserCartStmt: %w", cerr)
		}
	}
	if q.getUsersStmt != nil {
		if cerr := q.getUsersStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUsersStmt: %w", cerr)
		}
	}
	if q.updateOrderStmt != nil {
		if cerr := q.updateOrderStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateOrderStmt: %w", cerr)
		}
	}
	if q.updateOrderStatusStmt != nil {
		if cerr := q.updateOrderStatusStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateOrderStatusStmt: %w", cerr)
		}
	}
	if q.updatePaperStmt != nil {
		if cerr := q.updatePaperStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updatePaperStmt: %w", cerr)
		}
	}
	if q.updatePrintQuantityStmt != nil {
		if cerr := q.updatePrintQuantityStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updatePrintQuantityStmt: %w", cerr)
		}
	}
	if q.updateUserStmt != nil {
		if cerr := q.updateUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing updateUserStmt: %w", cerr)
		}
	}
	if q.upsertCartStmt != nil {
		if cerr := q.upsertCartStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing upsertCartStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db                      DBTX
	tx                      *sql.Tx
	addPaperStmt            *sql.Stmt
	addPictureStmt          *sql.Stmt
	addPrintStmt            *sql.Stmt
	addUserStmt             *sql.Stmt
	createOrderStmt         *sql.Stmt
	deleteOrderStmt         *sql.Stmt
	deletePaperStmt         *sql.Stmt
	deletePictureStmt       *sql.Stmt
	deletePrintStmt         *sql.Stmt
	deleteUserStmt          *sql.Stmt
	getCartsStmt            *sql.Stmt
	getOrderForUserStmt     *sql.Stmt
	getOrdersForUserStmt    *sql.Stmt
	getPapersStmt           *sql.Stmt
	getPicturesStmt         *sql.Stmt
	getPicturesByUserStmt   *sql.Stmt
	getUserStmt             *sql.Stmt
	getUserCartStmt         *sql.Stmt
	getUsersStmt            *sql.Stmt
	updateOrderStmt         *sql.Stmt
	updateOrderStatusStmt   *sql.Stmt
	updatePaperStmt         *sql.Stmt
	updatePrintQuantityStmt *sql.Stmt
	updateUserStmt          *sql.Stmt
	upsertCartStmt          *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                      tx,
		tx:                      tx,
		addPaperStmt:            q.addPaperStmt,
		addPictureStmt:          q.addPictureStmt,
		addPrintStmt:            q.addPrintStmt,
		addUserStmt:             q.addUserStmt,
		createOrderStmt:         q.createOrderStmt,
		deleteOrderStmt:         q.deleteOrderStmt,
		deletePaperStmt:         q.deletePaperStmt,
		deletePictureStmt:       q.deletePictureStmt,
		deletePrintStmt:         q.deletePrintStmt,
		deleteUserStmt:          q.deleteUserStmt,
		getCartsStmt:            q.getCartsStmt,
		getOrderForUserStmt:     q.getOrderForUserStmt,
		getOrdersForUserStmt:    q.getOrdersForUserStmt,
		getPapersStmt:           q.getPapersStmt,
		getPicturesStmt:         q.getPicturesStmt,
		getPicturesByUserStmt:   q.getPicturesByUserStmt,
		getUserStmt:             q.getUserStmt,
		getUserCartStmt:         q.getUserCartStmt,
		getUsersStmt:            q.getUsersStmt,
		updateOrderStmt:         q.updateOrderStmt,
		updateOrderStatusStmt:   q.updateOrderStatusStmt,
		updatePaperStmt:         q.updatePaperStmt,
		updatePrintQuantityStmt: q.updatePrintQuantityStmt,
		updateUserStmt:          q.updateUserStmt,
		upsertCartStmt:          q.upsertCartStmt,
	}
}
