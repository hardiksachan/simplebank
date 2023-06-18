package postgres

import (
	"context"
	"database/sql"
	"fmt"
	db "github.com/hardiksachan/simplebank/db/generated"
)

type Store struct {
	*db.Queries
	db *sql.DB
}

func NewStore(d *sql.DB) *Store {
	return &Store{
		Queries: db.New(d),
		db:      d,
	}
}

func (s *Store) execTx(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := db.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rbErr: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    db.Transfer `json:"transfer"`
	FromAccount db.Account  `json:"from_account"`
	ToAccount   db.Account  `json:"to_account"`
	FromEntry   db.Entry    `json:"from_entry"`
	ToEntry     db.Entry    `json:"to_entry"`
}

func (s *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := s.execTx(ctx, func(q *db.Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, db.CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, db.CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, db.CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, arg.ToAccountID, -arg.Amount, arg.Amount)
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.FromAccountID, arg.Amount, -arg.Amount)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *db.Queries,
	accID1 int64,
	accID2 int64, amt1 int64, amt2 int64,
) (acc1 db.Account, acc2 db.Account, err error) {
	acc1, err = q.AddAccountBalance(ctx, db.AddAccountBalanceParams{
		ID:     accID1,
		Amount: amt1,
	})
	if err != nil {
		return
	}

	acc2, err = q.AddAccountBalance(ctx, db.AddAccountBalanceParams{
		ID:     accID2,
		Amount: amt2,
	})
	return
}
