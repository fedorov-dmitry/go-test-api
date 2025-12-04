package logging

import (
	"context"
	"log"

	"github.com/fedorov-dmitry/go-test-api/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func StartLoggingWorker(ctx context.Context, pgPool *pgxpool.Pool, logCh <-chan middleware.RequestLog) {
	sql := "INSERT INTO app.logs (timestamp, path) VALUES ($1, $2)"

	go func() {
		for entry := range logCh {
			_, err := pgPool.Exec(ctx, sql, entry.Timestamp, entry.Path)

			if err != nil {
				log.Println("failed to insert log entry:", err)
			}
		}
	}()
}
