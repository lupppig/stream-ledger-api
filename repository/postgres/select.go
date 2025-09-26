package postgres

import (
	"context"
	"time"
)

func (p *PostgresDB) SelectSingleEntity(query string, receiver interface{}, id ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := p.DB.NewSelect().Model(receiver).Where(query, id...).Exec(ctx)

	if err != nil {
		return err
	}
	return nil
}
