package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthorizationMiddleware_Unauthorized(t *testing.T) {
	t.Parallel()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	h := AuthorizationMiddleware("secret", next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401", rr.Code)
	}
	if called {
		t.Fatal("next should not be called")
	}
}

func TestAuthorizationMiddleware_Authorized(t *testing.T) {
	t.Parallel()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	h := AuthorizationMiddleware("secret", next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/y", nil)
	req.Header.Set("Authorization", "secret")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rr.Code)
	}
	if !called {
		t.Fatal("next should be called")
	}
}

func TestRequestLoggingMiddleware_PushesLog(t *testing.T) {
	t.Parallel()

	logCh := make(chan RequestLog, 1)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := RequestLoggingMiddleware(logCh, next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/z", nil)
	h.ServeHTTP(rr, req)

	select {
	case entry := <-logCh:
		if entry.Path != "/z" {
			t.Fatalf("unexpected path: %s", entry.Path)
		}
		if time.Since(entry.Timestamp) > 5*time.Second {
			t.Fatalf("unexpected timestamp: %v", entry.Timestamp)
		}
	default:
		t.Fatal("expected a log entry")
	}
}
