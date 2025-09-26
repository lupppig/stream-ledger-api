package migrations

import (
	"context"

	"github.com/lupppig/stream-ledger-api/model"
	"github.com/uptrace/bun"
)

func init() {
	migrates.MustRegister(
		func(ctx context.Context, db *bun.DB) error {
			return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
				if _, err := tx.NewCreateTable().
					Model((*model.Transaction)(nil)).
					IfNotExists().
					Exec(ctx); err != nil {
					return err
				}
				return nil
			})
		},
		func(ctx context.Context, db *bun.DB) error {
			return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
				if _, err := tx.NewDropTable().
					Model((*model.Transaction)(nil)).
					IfExists().
					Exec(ctx); err != nil {
					return err
				}
				return nil
			})
		},
	)
}
