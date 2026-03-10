package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"tu-proyecto/model"
	"tu-proyecto/repository"
	"tu-proyecto/utils"

	"github.com/labstack/echo/v4"
)

// ErrorHandler maneja errores globales
func ErrorHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("🔥 PANIC RECUPERADO: %v", r)
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
		log.Printf("🍞 Breadcrumb - Path: %s", currentPath)

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
			breadcrumbs = append(breadcrumbs,
				map[string]string{"name": "Dashboard", "url": "/dashboard"},
				map[string]string{"name": "🔧 Mantenimiento", "url": "/maintenance"},
			)
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

		case "/perfil":
			breadcrumbs = append(breadcrumbs,
				map[string]string{"name": "Dashboard", "url": "/dashboard"},
				map[string]string{"name": "Mi Perfil", "url": "/perfil"},
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

// getRolNombre - Función auxiliar para obtener nombre del rol
func getRolNombre(roleID int) string {
	switch roleID {
	case 1:
		return "admin"
	case 2:
		return "user"
	case 3:
		return "editor"
	default:
		return "user"
	}
}

// dumpSession - Función auxiliar para mostrar contenido de sesión
func dumpSession(session *model.Session, tag string) {
	if session == nil {
		log.Printf("📭 [%s] SESIÓN: nil", tag)
		return
	}

	sessionJSON, _ := json.Marshal(session)
	log.Printf("📦 [%s] SESIÓN: %s", tag, string(sessionJSON))
}

// AuthMiddleware maneja la autenticación de usuarios y verifica rol actualizado
func AuthMiddleware(sm *utils.SessionManager, authRepo repository.AuthRepository) echo.MiddlewareFunc {
	log.Println("🚀 INICIALIZANDO AuthMiddleware V2 con verificación de BD")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			currentPath := c.Path()
			log.Printf("🔐 AuthMiddleware procesando: %s", currentPath)

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
			if publicPaths[currentPath] {
				log.Printf("✅ Ruta pública: %s - acceso permitido", currentPath)
				return next(c)
			}

			log.Printf("🔒 Ruta protegida: %s - verificando autenticación", currentPath)

			// Verificar si el usuario tiene sesión activa
			session, err := sm.GetSession(c)
			if err != nil {
				log.Printf("❌ Error obteniendo sesión: %v", err)
				return c.Redirect(http.StatusSeeOther, "/login?error=Por favor inicie sesión para continuar")
			}

			log.Printf("✅ Sesión encontrada para usuario ID: %d", session.UserID)
			dumpSession(session, "ANTES_VERIFICACION")

			// --- VERIFICACIÓN DE ROL ACTUALIZADO EN BD ---
			log.Printf("🔍 Verificando rol en BD para usuario %d (rol actual en sesión: %d)",
				session.UserID, session.RoleID)

			userActualizado, err := authRepo.FindByID(session.UserID)
			if err != nil {
				log.Printf("⚠️ Error al consultar BD para usuario %d: %v", session.UserID, err)
			} else {
				log.Printf("📊 Usuario en BD: ID=%d, Email=%s, RoleID=%d",
					userActualizado.ID, userActualizado.Email, userActualizado.RoleID)

				if userActualizado.RoleID != session.RoleID {
					log.Printf("🔄 ¡CAMBIO DETECTADO! Rol en sesión: %d, Rol en BD: %d",
						session.RoleID, userActualizado.RoleID)

					session.RoleID = userActualizado.RoleID
					session.RoleNombre = getRolNombre(userActualizado.RoleID)

					log.Printf("📝 Sesión actualizada en memoria: RoleID=%d, RoleNombre=%s",
						session.RoleID, session.RoleNombre)
					dumpSession(session, "DESPUES_ACTUALIZACION")

					// ¡IMPORTANTE! Actualizar la cookie también
					log.Printf("🍪 Actualizando cookie con nuevo rol...")
					if err := sm.UpdateSession(c, session); err != nil {
						log.Printf("❌ Error al actualizar cookie: %v", err)
					} else {
						log.Printf("✅ Cookie actualizada exitosamente con rol %d", session.RoleID)
					}
				} else {
					log.Printf("✅ Roles coinciden (sesión=%d, BD=%d) - no hay cambios",
						session.RoleID, userActualizado.RoleID)
				}
			}
			// ---------------------------------------------

			// Actualizar última actividad de la sesión
			session.LastActivity = time.Now()

			// Guardar información del usuario en el contexto para los handlers
			c.Set("user_id", session.UserID)
			c.Set("user_name", session.Name)
			c.Set("user_email", session.Email)
			c.Set("user_role_id", session.RoleID)
			c.Set("user_role", session.RoleNombre)

			log.Printf("🎯 Contexto actualizado - RoleID: %d, RoleNombre: %s",
				session.RoleID, session.RoleNombre)

			// Continuar con la siguiente función
			return next(c)
		}
	}
}

// AdminMiddleware - Solo permite acceso a usuarios con role_id = 1 (admin)
func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Printf("👑 AdminMiddleware verificando acceso")

		// Obtener role_id del contexto (establecido por AuthMiddleware)
		roleID, ok := c.Get("user_role_id").(int)
		if !ok {
			log.Printf("❌ No se pudo obtener role_id del contexto")
			return c.Redirect(http.StatusSeeOther, "/dashboard?error=Acceso no autorizado")
		}

		roleNombre, _ := c.Get("user_role").(string)
		log.Printf("👤 Usuario con rol: %d (%s) intenta acceder a ruta admin", roleID, roleNombre)

		// Verificar si es admin (role_id = 1)
		if roleID != 1 {
			log.Printf("⛔ Acceso denegado - rol %d no es admin", roleID)
			return c.Redirect(http.StatusSeeOther, "/dashboard?error=Área restringida para administradores")
		}

		log.Printf("✅ Acceso permitido para admin")
		return next(c)
	}
}

// PermissionMiddleware - Verifica permisos específicos (versión simple)
func PermissionMiddleware(allowedRoles ...int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roleID, ok := c.Get("user_role_id").(int)
			if !ok {
				log.Printf("❌ PermissionMiddleware: No se pudo obtener role_id")
				return c.Redirect(http.StatusSeeOther, "/dashboard?error=Acceso no autorizado")
			}

			log.Printf("🔑 PermissionMiddleware - Usuario rol: %d, Roles permitidos: %v", roleID, allowedRoles)

			// Verificar si el rol del usuario está en los roles permitidos
			for _, allowed := range allowedRoles {
				if roleID == allowed {
					log.Printf("✅ Acceso permitido - rol %d está en lista permitida", roleID)
					return next(c)
				}
			}

			log.Printf("⛔ Acceso denegado - rol %d no está en lista permitida", roleID)
			return c.Redirect(http.StatusSeeOther, "/dashboard?error=No tiene permisos para esta acción")
		}
	}
}

// LoadUserDataMiddleware - Carga datos del usuario para todas las rutas (incluso públicas)
func LoadUserDataMiddleware(sm *utils.SessionManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Printf("📊 LoadUserDataMiddleware - Path: %s", c.Path())

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
				log.Printf("✅ Usuario autenticado: %s (rol: %s)", session.Email, session.RoleNombre)
			} else {
				// No hay sesión
				c.Set("is_authenticated", false)
				log.Printf("ℹ️ Usuario no autenticado")
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
			log.Printf("🔄 Redirigiendo HTTP a HTTPS")
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
