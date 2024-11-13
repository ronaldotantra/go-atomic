package atomic

import (
	"context"
	"database/sql"
)

type Executor interface {
	Run(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context) error) error
}

type executor struct {
	db *sql.DB
}

func New(db *sql.DB) Executor {
	return &executor{
		db: db,
	}
}

// Run will begin a transaction if there is no tx in the context
// If the fn return an error, the transaction will be rollback
// If the fn success without any error, the transaction will be commited
func (e *executor) Run(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context) error) error {
	tx := GetTransactionClient(ctx)
	if tx != nil {
		return fn(ctx)
	}

	tx, err := e.db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	ctxWithTx := setTransactionClient(ctx, tx)

	err = fn(ctxWithTx)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
