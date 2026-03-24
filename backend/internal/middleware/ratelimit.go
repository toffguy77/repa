package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

func RateLimit(rdb *redis.Client, key string, limit int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identifier := c.RealIP()
			if claims := c.Get("user"); claims != nil {
				if u, ok := claims.(*JWTClaims); ok {
					identifier = u.UserID
				}
			}

			redisKey := fmt.Sprintf("rl:%s:%s", key, identifier)
			ctx := c.Request().Context()

			pipe := rdb.TxPipeline()
			incrCmd := pipe.Incr(ctx, redisKey)
			pipe.Expire(ctx, redisKey, window)
			if _, err := pipe.Exec(ctx); err != nil {
				return next(c)
			}

			if incrCmd.Val() > int64(limit) {
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"error": map[string]string{
						"code":    "RATE_LIMIT",
						"message": "Слишком много запросов",
					},
				})
			}

			return next(c)
		}
	}
}
