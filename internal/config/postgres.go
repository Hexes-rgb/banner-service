package config

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresConfig struct {
	ConnStr         string
	MaxConns        int32
	MaxConnIdleTime time.Duration
}

func NewPostgresPool(cfg PostgresConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.ConnStr)
	if err != nil {
		return nil, err
	}

	config.MaxConns = cfg.MaxConns
	config.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
