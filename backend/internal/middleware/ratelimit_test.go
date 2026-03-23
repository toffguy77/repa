package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

func testRedis(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 15})
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping rate limit test")
	}
	rdb.FlushDB(ctx)
	t.Cleanup(func() {
		rdb.FlushDB(ctx)
		rdb.Close()
	})
	return rdb
}

func TestRateLimit_BlocksAfterLimit(t *testing.T) {
	rdb := testRedis(t)

	e := echo.New()
	e.POST("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, RateLimit(rdb, "test-block", 3, time.Minute))

	for i := range 3 {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set("X-Real-Ip", "10.0.0.1")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// 4th request should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("X-Real-Ip", "10.0.0.1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 after limit, got %d", rec.Code)
	}
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	rdb := testRedis(t)

	e := echo.New()
	e.POST("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, RateLimit(rdb, "test-diff-ip", 1, time.Minute))

	// First IP gets through
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("X-Real-Ip", "10.0.0.1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first IP: expected 200, got %d", rec.Code)
	}

	// Different IP should also get through
	req = httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("X-Real-Ip", "10.0.0.2")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("different IP: expected 200, got %d", rec.Code)
	}
}
