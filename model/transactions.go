package model

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/lupppig/stream-ledger-api/repository/postgres"
	"github.com/lupppig/stream-ledger-api/utils"
	"github.com/uptrace/bun"
)

var ErrorInsuffcientBalance = errors.New("insufficient balance")
var ErrorDuplicateTransaction = errors.New("duplicate transaction")

type Transaction struct {
	ID        int64     `bun:",pk,autoincrement" json:"transaction_id"`
	WalletID  int64     `bun:"column:wallet_id,notnull" json:"wallet_id"`
	Entry     string    `bun:"type:transaction_entry,notnull" json:"entry"` // credit or debit
	Amount    int64     `bun:",notnull" json:"amount"`                      // in kobo
	TransID   string    `bun:",unique" json:"trans_id"`
	CreatedAt time.Time `bun:",nullzero,default:current_timestamp" json:"created_at"`
	Wallet    *Wallet   `bun:"rel:belongs-to,join:wallet_id=id" json:"-"`
}

func (t *Transaction) CreateTransaction(db *postgres.PostgresDB, userId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return db.DB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// check if transaction exist to prevent duplicate transaction
		existing := new(Transaction)
		err := tx.NewSelect().
			Model(existing).
			Where("trans_id = ?", t.TransID).
			Scan(ctx)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		} else if err == nil {
			return ErrorDuplicateTransaction
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

func (t *Transaction) GetUserTransaction(db *postgres.PostgresDB, userId int64, pagination utils.Pagination) ([]*Transaction, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var total int
	var transactions []*Transaction
	var err error

	total, err = db.DB.NewSelect().
		Model(t).
		Join(`JOIN wallets w ON "w".id = transaction.wallet_id`).
		Where(`w.user_id = ?`, userId).
		Count(ctx)

	if err != nil {
		log.Println(err.Error())
		return nil, 0, err
	}

	err = db.DB.NewSelect().
		Model(&transactions).
		Relation("Wallet"). // Eager load wallet
		Where("wallet.user_id = ?", userId).
		Order("transaction.created_at DESC").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Scan(ctx)
	if err != nil {
		log.Println(err.Error())
		return nil, 0, err
	}
	return transactions, total, nil
}
