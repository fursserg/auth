package db

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	dbDSN         = "host=localhost port=54322 dbname=auth user=auth password=abcdef sslmode=disable"
	activeStatus  = 1
	deletedStatus = 10
)

var (
	connect *pgxpool.Pool
	once    sync.Once
)

type Statuses string

// Get Получает номер статуса по названию
func (s Statuses) Get() (int, error) {
	switch s {
	case "active":
		return activeStatus, nil
	case "deleted":
		return deletedStatus, nil
	default:
		return 0, fmt.Errorf("undefined status")
	}
}

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
