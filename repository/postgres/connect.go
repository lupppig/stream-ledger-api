package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type PostgresDB struct {
	DB *bun.DB
}

func Connect(connString string) (*PostgresDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(connString)

	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		return nil, err
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	db := bun.NewDB(sqlDB, pgdialect.New())

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return NewPostgresDB(db), nil
}

func NewPostgresDB(db *bun.DB) *PostgresDB {
	return &PostgresDB{DB: db}
}
