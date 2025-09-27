package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrorDuplicateEmail = errors.New("email exists")

func (p *PostgresDB) CreateModel(model interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := p.DB.NewInsert().Model(model).Returning("*").Exec(ctx)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrorDuplicateEmail
			}
		}
		return err
	}

	if aff, err := res.RowsAffected(); err != nil || aff == 0 {
		return fmt.Errorf("no data insert: %w", err)
	}
	return nil
}
