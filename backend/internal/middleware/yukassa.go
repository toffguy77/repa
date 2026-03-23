package middleware

import (
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
)

// YuKassa webhook IP ranges from https://yookassa.ru/developers/using-api/webhooks
var yukassaCIDRs = []string{
	"185.71.76.0/27",
	"185.71.77.0/27",
	"77.75.153.0/25",
	"77.75.156.11/32",
	"77.75.156.35/32",
	"77.75.154.128/25",
	"2a02:5180::/32",
}

func YukassaIPAllowlist() echo.MiddlewareFunc {
	nets := make([]*net.IPNet, 0, len(yukassaCIDRs))
	for _, cidr := range yukassaCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		nets = append(nets, ipNet)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := net.ParseIP(c.RealIP())
			if ip == nil {
				return c.JSON(http.StatusForbidden, map[string]any{
					"error": map[string]string{
						"code":    "FORBIDDEN",
						"message": "Access denied",
					},
				})
			}

			for _, ipNet := range nets {
				if ipNet.Contains(ip) {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]any{
				"error": map[string]string{
					"code":    "FORBIDDEN",
					"message": "Access denied",
				},
			})
		}
	}
}
