package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrorDuplicateEmail = errors.New("email exists")

func (p *PostgresDB) CreateModel(model interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.NewInsert().Model(model).Returning("*").Scan(ctx, model)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return ErrorDuplicateEmail
			}
		}
		return err
	}

	return nil
}
