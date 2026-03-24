package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func JWTAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]string{
						"code":    "UNAUTHORIZED",
						"message": "Missing authorization header",
					},
				})
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]string{
						"code":    "UNAUTHORIZED",
						"message": "Invalid authorization format",
					},
				})
			}

			token, err := jwt.ParseWithClaims(parts[1], &JWTClaims{}, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]string{
						"code":    "UNAUTHORIZED",
						"message": "Invalid or expired token",
					},
				})
			}

			claims, ok := token.Claims.(*JWTClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]string{
						"code":    "UNAUTHORIZED",
						"message": "Invalid token claims",
					},
				})
			}

			c.Set("user", claims)
			return next(c)
		}
	}
}

func GetCurrentUser(c echo.Context) *JWTClaims {
	v, _ := c.Get("user").(*JWTClaims)
	return v
}
