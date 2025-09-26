package migrations

import (
	"context"
	"log"
	"time"

	"github.com/lupppig/stream-ledger-api/model"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

var migrates = migrate.NewMigrations()

func init() {
	migrates.MustRegister(
		// UP
		func(ctx context.Context, db *bun.DB) error {
			return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
				if _, err := tx.NewCreateTable().
					Model((*model.User)(nil)).
					IfNotExists().
					Exec(ctx); err != nil {
					return err
				}
				return nil
			})
		},
		// DOWN
		func(ctx context.Context, db *bun.DB) error {
			return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
				if _, err := tx.NewDropTable().
					Model((*model.User)(nil)).
					IfExists().
					Exec(ctx); err != nil {
					return err
				}
				return nil
			})
		},
	)
}

func RunMigrations(db *bun.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	migrator := migrate.NewMigrator(db, migrates)

	if err := migrator.Init(ctx); err != nil {
		return err
	}
	group, err := migrator.Migrate(ctx)

	if err != nil {
		return err
	}
	if group.IsZero() {
		log.Println("No new migration to run")
	} else {
		log.Printf("migrated to %s\n", group)
	}
	return nil
}
