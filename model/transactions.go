package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lupppig/stream-ledger-api/repository/postgres"
	"github.com/uptrace/bun"
)

var ErrorInsuffcientBalance = errors.New("insufficient balance")
var ErrorDuplicateTransaction = errors.New("duplicate transaction")

type Transaction struct {
	ID        int64     `bun:",pk,autoincrement"`
	WalletID  int64     `bun:"column:wallet_id,notnull"`
	Entry     string    `bun:"type:transaction_entry,notnull"` // credit or debit
	Amount    int64     `bun:",notnull"`                       // in kobo
	TransID   string    `bun:",unique"`
	CreatedAt time.Time `bun:",nullzero,default:current_timestamp"`
	Wallet    *Wallet   `bun:"rel:belongs-to,join:wallet_id=id"`
}

func (t *Transaction) CreateTransaction(db *postgres.PostgresDB, userId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return db.DB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// check if transaction exist to prevent duplicate transaction
		existing := new(Transaction)
		fmt.Println(t.TransID)
		err := tx.NewSelect().
			Model(existing).
			Where("trans_id = ?", t.TransID).
			Scan(ctx)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		} else if err == nil {
			return fmt.Errorf("duplicate transaction")
		}

		var wallet = &Wallet{}
		if err := wallet.getWallet(tx, userId); err != nil {
			return err
		}

		if t.Entry == "debit" && wallet.Balance < t.Amount {
			return ErrorInsuffcientBalance
		}

		if t.Entry == "credit" {
			_, err := tx.NewRaw(`
			UPDATE wallets 
			SET balance = balance + ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`, t.Amount, wallet.ID).Exec(ctx)
			if err != nil {
				return err
			}
		} else {
			res, err := tx.NewRaw(`
			UPDATE wallets 
			SET balance = balance - ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND balance >= ?`, t.Amount, wallet.ID, t.Amount).Exec(ctx)
			if err != nil {
				return err
			}
			if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
				return ErrorInsuffcientBalance
			}
		}

		t.WalletID = wallet.ID
		t.Wallet = wallet
		if _, err := tx.NewInsert().
			Model(t).
			Exec(ctx); err != nil {
			return err
		}
		return nil
	})

}
