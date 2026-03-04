package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	// servicios necesarios
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) ShowDashboard(c echo.Context) error {
	userName := c.Get("user_name")

	data := map[string]interface{}{
		"Title":       "Dashboard Principal",
		"UserName":    userName,
		"breadcrumbs": c.Get("breadcrumbs"),
	}

	return c.Render(http.StatusOK, "dashboard.html", data)
}
