package crystals

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/repa-app/repa/internal/handler"
	appmw "github.com/repa-app/repa/internal/middleware"
	crystalssvc "github.com/repa-app/repa/internal/service/crystals"
)

type Handler struct {
	svc *crystalssvc.Service
}

func NewHandler(svc *crystalssvc.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetBalance(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)

	balance, err := h.svc.GetBalance(c.Request().Context(), claims.UserID)
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Something went wrong")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"balance": balance,
		},
	})
}

func (h *Handler) GetPackages(c echo.Context) error {
	packages := h.svc.GetPackages()
	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"packages": packages,
		},
	})
}

type initPurchaseRequest struct {
	PackageID string `json:"package_id" validate:"required"`
}

func (h *Handler) InitPurchase(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)

	var req initPurchaseRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	result, err := h.svc.InitPurchase(c.Request().Context(), claims.UserID, req.PackageID)
	if err != nil {
		if errors.Is(err, crystalssvc.ErrPackageNotFound) {
			return handler.ErrorResponse(c, http.StatusBadRequest, "PACKAGE_NOT_FOUND", "Invalid package ID")
		}
		if errors.Is(err, crystalssvc.ErrPaymentsUnavailable) {
			return handler.ErrorResponse(c, http.StatusServiceUnavailable, "PAYMENTS_UNAVAILABLE", "Payments are not configured")
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to create payment")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": result,
	})
}

func (h *Handler) Webhook(c echo.Context) error {
	var event crystalssvc.WebhookEvent
	if err := c.Bind(&event); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := h.svc.ProcessWebhook(c.Request().Context(), event); err != nil {
		if errors.Is(err, crystalssvc.ErrPaymentNotFound) {
			return c.NoContent(http.StatusOK)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) VerifyPurchase(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)
	paymentID := c.Param("paymentId")

	result, err := h.svc.VerifyPurchase(c.Request().Context(), claims.UserID, paymentID)
	if err != nil {
		if errors.Is(err, crystalssvc.ErrPaymentNotFound) {
			return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Payment not found")
		}
		if errors.Is(err, crystalssvc.ErrPaymentsUnavailable) {
			return handler.ErrorResponse(c, http.StatusServiceUnavailable, "PAYMENTS_UNAVAILABLE", "Payments are not configured")
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to verify payment")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": result,
	})
}
