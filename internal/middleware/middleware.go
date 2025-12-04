package middleware

import (
	"net/http"
	"time"
)

func AuthorizationMiddleware(expectedAPIKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("Authorization")
		if apiKey != expectedAPIKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequestLoggingMiddleware(logCh chan<- RequestLog, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		logCh <- RequestLog{
			Timestamp: time.Now(),
			Path:      r.URL.Path,
		}
	})
}

type RequestLog struct {
	Timestamp time.Time
	Path      string
}
