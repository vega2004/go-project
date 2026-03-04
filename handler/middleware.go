package handler

import (
	"log"
	"net/http"
	"time"
	"tu-proyecto/utils"

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

		switch currentPath {
		case "/form":
			breadcrumbs = append(breadcrumbs, map[string]string{
				"name": "Registro",
				"url":  "/form",
			})
		case "/success":
			breadcrumbs = append(breadcrumbs,
				map[string]string{"name": "Registro", "url": "/form"},
				map[string]string{"name": "Éxito", "url": "/success"},
			)
		case "/maintenance":
			breadcrumbs = append(breadcrumbs, map[string]string{
				"name": "🔧 Mantenimiento",
				"url":  "/maintenance",
			})
		case "/login":
			breadcrumbs = append(breadcrumbs, map[string]string{
				"name": "Iniciar Sesión",
				"url":  "/login",
			})
		case "/register":
			breadcrumbs = append(breadcrumbs, map[string]string{
				"name": "Registro",
				"url":  "/register",
			})
		case "/dashboard":
			breadcrumbs = append(breadcrumbs,
				map[string]string{"name": "Dashboard", "url": "/dashboard"},
			)
		case "/crud":
			breadcrumbs = append(breadcrumbs,
				map[string]string{"name": "Dashboard", "url": "/dashboard"},
				map[string]string{"name": "CRUD", "url": "/crud"},
			)
		case "/carrusel":
			breadcrumbs = append(breadcrumbs,
				map[string]string{"name": "Dashboard", "url": "/dashboard"},
				map[string]string{"name": "Carrusel", "url": "/carrusel"},
			)
		case "/admin/users":
			breadcrumbs = append(breadcrumbs,
				map[string]string{"name": "Dashboard", "url": "/dashboard"},
				map[string]string{"name": "Admin", "url": "/admin/users"},
			)
		case "/admin/users/create":
			breadcrumbs = append(breadcrumbs,
				map[string]string{"name": "Dashboard", "url": "/dashboard"},
				map[string]string{"name": "Admin", "url": "/admin/users"},
				map[string]string{"name": "Crear Usuario", "url": ""},
			)
		}

		c.Set("breadcrumbs", breadcrumbs)
		return next(c)
	}
}

// AuthMiddleware maneja la autenticación de usuarios
func AuthMiddleware(sm *utils.SessionManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Lista de rutas públicas (no requieren autenticación)
			publicPaths := map[string]bool{
				"/":            true,
				"/form":        true,
				"/submit":      true,
				"/login":       true,
				"/do-login":    true,
				"/register":    true,
				"/do-register": true,
				"/maintenance": true,
				"/debug":       true,
				"/health":      true,
				"/success":     true,
			}

			// Si es ruta pública, permitir acceso sin autenticación
			if publicPaths[c.Path()] {
				return next(c)
			}

			// Verificar si el usuario tiene sesión activa
			session, err := sm.GetSession(c)
			if err != nil {
				// No hay sesión, redirigir al login
				return c.Redirect(http.StatusSeeOther, "/login?error=Por favor inicie sesión para continuar")
			}

			// Actualizar última actividad de la sesión
			session.LastActivity = time.Now()

			// Guardar información del usuario en el contexto para los handlers
			c.Set("user_id", session.UserID)
			c.Set("user_name", session.Name)
			c.Set("user_email", session.Email)
			c.Set("user_role_id", session.RoleID)
			c.Set("user_role", session.RoleNombre)

			// Continuar con la siguiente función
			return next(c)
		}
	}
}

// AdminMiddleware - Solo permite acceso a usuarios con role_id = 1 (admin)
func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Obtener role_id del contexto (establecido por AuthMiddleware)
		roleID, ok := c.Get("user_role_id").(int)
		if !ok {
			// Si no hay role_id, redirigir al dashboard
			return c.Redirect(http.StatusSeeOther, "/dashboard?error=Acceso no autorizado")
		}

		// Verificar si es admin (role_id = 1)
		if roleID != 1 {
			// No es admin, redirigir con mensaje
			return c.Redirect(http.StatusSeeOther, "/dashboard?error=Área restringida para administradores")
		}

		// Es admin, continuar
		return next(c)
	}
}

// PermissionMiddleware - Verifica permisos específicos (versión simple)
func PermissionMiddleware(allowedRoles ...int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roleID, ok := c.Get("user_role_id").(int)
			if !ok {
				return c.Redirect(http.StatusSeeOther, "/dashboard?error=Acceso no autorizado")
			}

			// Verificar si el rol del usuario está en los roles permitidos
			for _, allowed := range allowedRoles {
				if roleID == allowed {
					return next(c)
				}
			}

			// No tiene permiso
			return c.Redirect(http.StatusSeeOther, "/dashboard?error=No tiene permisos para esta acción")
		}
	}
}

// LoadUserDataMiddleware - Carga datos del usuario para todas las rutas (incluso públicas)
func LoadUserDataMiddleware(sm *utils.SessionManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Intentar obtener sesión (si existe)
			session, err := sm.GetSession(c)
			if err == nil {
				// Hay sesión, cargar datos al contexto
				c.Set("user_id", session.UserID)
				c.Set("user_name", session.Name)
				c.Set("user_email", session.Email)
				c.Set("user_role_id", session.RoleID)
				c.Set("user_role", session.RoleNombre)
				c.Set("is_authenticated", true)
			} else {
				// No hay sesión
				c.Set("is_authenticated", false)
			}

			return next(c)
		}
	}
}

// RequireHTTPSMiddleware - Redirige HTTP a HTTPS en producción
func RequireHTTPSMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Solo en producción
		if c.Request().Header.Get("X-Forwarded-Proto") == "http" {
			return c.Redirect(http.StatusMovedPermanently, "https://"+c.Request().Host+c.Request().URL.String())
		}
		return next(c)
	}
}

// RateLimitMiddleware - Límite de peticiones (versión simple)
func RateLimitMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Aquí implementarías lógica de rate limiting
		// Por ahora, solo pasa
		return next(c)
	}
}
