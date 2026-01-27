package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorHandler maneja errores globales
func ErrorHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC RECUPERADO: %v", r)
				c.Redirect(http.StatusSeeOther, "/maintenance")
			}
		}()
		return next(c)
	}
}

// BreadcrumbMiddleware maneja las migajas de pan
func BreadcrumbMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		breadcrumbs := []map[string]string{
			{"name": "Inicio", "url": "/"},
		}

		currentPath := c.Path()

		if currentPath == "/form" {
			breadcrumbs = append(breadcrumbs, map[string]string{
				"name": "Registro",
				"url":  "/form",
			})
		} else if currentPath == "/success" {
			breadcrumbs = append(breadcrumbs,
				map[string]string{"name": "Registro", "url": "/form"},
				map[string]string{"name": "Éxito", "url": "/success"},
			)
		} else if currentPath == "/maintenance" {
			breadcrumbs = append(breadcrumbs, map[string]string{
				"name": "🔧 Mantenimiento",
				"url":  "/maintenance",
			})
		}

		c.Set("breadcrumbs", breadcrumbs)
		return next(c)
	}
}
