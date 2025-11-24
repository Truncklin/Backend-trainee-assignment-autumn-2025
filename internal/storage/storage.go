package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(storagePath string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dbConf, err := pgxpool.ParseConfig(storagePath)
	if err != nil {
		return nil, err
	}

	dbConf.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, dbConf)
	if err != nil {
		return nil, err
	}

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelPing()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func RunMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	migrationsPath := "/migrations/init.sql"

	data, err := os.ReadFile(filepath.Clean(migrationsPath))
	if err != nil {
		return err
	}

	queries := strings.Split(string(data), ";")

	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		_, err := pool.Exec(ctx, q)
		if err != nil {
			return err
		}
	}

	return nil
}
