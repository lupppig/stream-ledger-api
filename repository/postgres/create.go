package postgres

import (
	"context"
	"fmt"
	"time"
)

func (p *PostgresDB) CreateModel(model interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := p.DB.NewInsert().Model(model).Returning("*").Exec(ctx)
	if err != nil {
		return err
	}

	if aff, err := res.RowsAffected(); err != nil || aff == 0 {
		return fmt.Errorf("no data insert: %w", err)
	}
	return nil
}
