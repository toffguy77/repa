package admin

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/handler"
)

type Handler struct {
	queries  db.Querier
	username string
	password string
}

func NewHandler(queries db.Querier, username, password string) *Handler {
	return &Handler{queries: queries, username: username, password: password}
}

func (h *Handler) BasicAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if h.password == "" {
			return handler.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Admin not configured")
		}
		user, pass, ok := c.Request().BasicAuth()
		if !ok || user != h.username || pass != h.password {
			c.Response().Header().Set("WWW-Authenticate", `Basic realm="admin"`)
			return handler.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid credentials")
		}
		return next(c)
	}
}

func (h *Handler) ListReports(c echo.Context) error {
	ctx := c.Request().Context()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	reports, err := h.queries.ListReports(ctx, db.ListReportsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to list reports")
	}

	total, err := h.queries.CountReports(ctx)
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to count reports")
	}

	items := make([]map[string]any, len(reports))
	for i, r := range reports {
		var reason *string
		if r.Reason.Valid {
			reason = &r.Reason.String
		}
		items[i] = map[string]any{
			"id":                r.ID,
			"question_id":      r.QuestionID,
			"question_text":    r.QuestionText,
			"question_category": r.QuestionCategory,
			"question_status":  r.QuestionStatus,
			"reporter_id":      r.ReporterID,
			"reporter_username": r.ReporterUsername,
			"reason":           reason,
			"created_at":       r.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

type resolveReportRequest struct {
	Action string `json:"action" validate:"required,oneof=approve reject"`
}

func (h *Handler) ResolveReport(c echo.Context) error {
	ctx := c.Request().Context()
	reportID := c.Param("id")

	var req resolveReportRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	report, err := h.queries.GetReportByID(ctx, reportID)
	if err != nil {
		return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Report not found")
	}

	var newStatus db.QuestionStatus
	if req.Action == "approve" {
		newStatus = db.QuestionStatusACTIVE
	} else {
		newStatus = db.QuestionStatusREJECTED
	}

	err = h.queries.UpdateQuestionStatus(ctx, db.UpdateQuestionStatusParams{
		ID:     report.QuestionID,
		Status: newStatus,
	})
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to update question")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"report_id":   reportID,
			"action":      req.Action,
			"question_id": report.QuestionID,
			"new_status":  string(newStatus),
		},
	})
}

func (h *Handler) GetStats(c echo.Context) error {
	ctx := c.Request().Context()

	dau, _ := h.queries.CountActiveUsers7Days(ctx)
	mau, _ := h.queries.CountActiveUsers30Days(ctx)
	groups, _ := h.queries.CountGroups(ctx)
	revenue, _ := h.queries.SumRevenue7Days(ctx)

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"dau_7d":          dau,
			"mau_30d":        mau,
			"groups_count":   groups,
			"revenue_7d_rub": revenue,
		},
	})
}
