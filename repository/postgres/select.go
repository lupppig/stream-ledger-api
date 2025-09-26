package postgres

import (
	"context"
	"time"
)

func (p *PostgresDB) SelectSingleEntity(query string, receiver interface{}, id ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return p.DB.NewSelect().Model(receiver).Where(query, id...).Scan(ctx)
}

func (p *PostgresDB) SelectOneWithRelation(query string, relation string, receiver interface{}, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return p.DB.NewSelect().
		Model(receiver).
		Relation(relation).Where(query, args...).Scan(ctx)
}
