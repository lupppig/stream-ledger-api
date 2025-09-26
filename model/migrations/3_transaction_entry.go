package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	migrates.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`CREATE TYPE transaction_entry AS ENUM ('credit', 'debit')`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.Exec(`DROP TYPE transaction_entry`)
		return err
	})

}
