package repository

import (
	"context"
	"mm/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func CreateDBConnection(c *config.Config) (*bun.DB, error) {
	conf, err := pgxpool.ParseConfig(c.DB.GetConnectString())
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return nil, err
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	db := bun.NewDB(sqlDB, pgdialect.New())

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
