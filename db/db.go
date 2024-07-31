package db

import (
	"context"
	"log"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	dbDSN         = "host=localhost port=54322 dbname=auth user=auth password=abcdef sslmode=disable"
	ActiveStatus  = 1
	DeletedStatus = 10
)

var (
	connect *pgxpool.Pool
	once    sync.Once
)

// GetConnect Получает коннект к БД.
// Если ранее коннект уже был установлен, то его и вернет,
// иначе установит новый
func GetConnect() *pgxpool.Pool {
	if connect != nil {
		return connect
	}

	once.Do(func() {
		ctx := context.Background()

		pool, err := pgxpool.Connect(ctx, dbDSN)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}

		connect = pool
	})

	return connect
}
