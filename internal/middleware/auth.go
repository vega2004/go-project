package middleware

import (
	"log"
	"net/http"
	"time"
	"tu-proyecto/internal/repository"
	"tu-proyecto/internal/utils"

	"github.com/labstack/echo/v4"
)

// ============================================
// ESTRUCTURA PRINCIPAL (ACTUALIZADA)
// ============================================

type AuthMiddleware struct {
	sessionManager *utils.SessionManager
	jwtManager     *utils.JWTManager // ← NUEVO: para JWT
	authRepo       repository.AuthRepository
	permisoRepo    repository.PermisoRepository
	useJWT         bool // ← NUEVO: habilitar/deshabilitar JWT
}

// NewAuthMiddleware - Constructor ACTUALIZADO
func NewAuthMiddleware(
	sm *utils.SessionManager,
	jwtManager *utils.JWTManager,
	authRepo repository.AuthRepository,
	permisoRepo repository.PermisoRepository,
	useJWT bool,
) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sm,
		jwtManager:     jwtManager,
		authRepo:       authRepo,
		permisoRepo:    permisoRepo,
		useJWT:         useJWT,
	}
}

// ============================================
// MIDDLEWARE PRINCIPAL (ACTUALIZADO con JWT)
// ============================================

// RequireAuth - Verifica autenticación (soporta Sesión + JWT)
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		// ============================================
		// 1. INTENTAR AUTENTICACIÓN POR JWT (si está habilitado)
		// ============================================
		if m.useJWT && m.jwtManager != nil {
			tokenString, err := m.jwtManager.ExtractFromRequest(c)
			if err == nil {
				claims, err := m.jwtManager.Validate(tokenString)
				if err == nil {
					// Verificar que el usuario existe y está activo en BD
					user, err := m.authRepo.FindByID(claims.UserID)
					if err == nil && user.Activo {
						// Guardar en contexto
						c.Set("user_id", claims.UserID)
						c.Set("user_name", claims.Name)
						c.Set("user_email", claims.Email)
						c.Set("user_role_id", claims.RoleID)
						c.Set("user_role", claims.RoleName)
						c.Set("is_authenticated", true)
						c.Set("auth_method", "jwt")

						log.Printf("[AUTH] Usuario %d (%s) autenticado por JWT - Duración: %v",
							claims.UserID, claims.Email, time.Since(start))
						return next(c)
					}
				}
			}
		}

		// ============================================
		// 2. FALLBACK: AUTENTICACIÓN POR SESIÓN
		// ============================================
		session, err := m.sessionManager.GetSession(c)
		if err != nil {
			log.Printf("[AUTH] Acceso denegado - Sin sesión: %v", err)
			return c.Redirect(http.StatusSeeOther, "/login?error=Debe iniciar sesión")
		}

		// Verificar usuario en BD
		user, err := m.authRepo.FindByID(session.UserID)
		if err != nil {
			log.Printf("[AUTH] Usuario %d no encontrado en BD", session.UserID)
			m.sessionManager.ClearSession(c)
			return c.Redirect(http.StatusSeeOther, "/login?error=Usuario no encontrado")
		}

		// Verificar si está activo
		if !user.Activo {
			log.Printf("[AUTH] Usuario %d está desactivado", session.UserID)
			m.sessionManager.ClearSession(c)
			return c.Redirect(http.StatusSeeOther, "/login?error=Usuario desactivado")
		}

		// Actualizar rol si cambió en BD
		if session.RoleID != user.RoleID {
			log.Printf("[AUTH] Actualizando rol usuario %d: %d -> %d", user.ID, session.RoleID, user.RoleID)
			session.RoleID = user.RoleID
			session.RoleNombre = m.getRolNombre(user.RoleID)
			m.sessionManager.UpdateSession(c, session)
		}

		// Actualizar última actividad
		session.LastActivity = time.Now()

		// Guardar en contexto
		c.Set("user_id", session.UserID)
		c.Set("user_name", session.Name)
		c.Set("user_email", session.Email)
		c.Set("user_role_id", session.RoleID)
		c.Set("user_role", session.RoleNombre)
		c.Set("is_authenticated", true)
		c.Set("auth_method", "session")

		log.Printf("[AUTH] Usuario %d (%s) autenticado por sesión como %s - Duración: %v",
			session.UserID, session.Email, session.RoleNombre, time.Since(start))

		return next(c)
	}
}

// ============================================
// MIDDLEWARE DE ROLES
// ============================================

// RequireRole - Verifica que el usuario tenga un rol específico
func (m *AuthMiddleware) RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole, ok := c.Get("user_role").(string)
			if !ok {
				log.Printf("[AUTH] No se pudo obtener rol del usuario")
				return c.Redirect(http.StatusSeeOther, "/dashboard?error=Acceso no autorizado")
			}

			for _, role := range allowedRoles {
				if userRole == role {
					log.Printf("[AUTH] Usuario con rol %s autorizado", userRole)
					return next(c)
				}
			}

			log.Printf("[AUTH] Usuario con rol %s no autorizado (requiere: %v)", userRole, allowedRoles)
			return c.Redirect(http.StatusSeeOther, "/dashboard?error=No tiene permisos para esta página")
		}
	}
}

// RequireAdmin - Atajo para RequireRole("administrador")
func (m *AuthMiddleware) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return m.RequireRole("administrador")(next)
}

// ============================================
// MIDDLEWARE DE PERMISOS ESPECÍFICOS
// ============================================

// RequirePermission - Verifica un permiso específico en un módulo
func (m *AuthMiddleware) RequirePermission(moduloNombre, permiso string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get("user_id").(int)
			if !ok {
				return c.Redirect(http.StatusSeeOther, "/login?error=Debe iniciar sesión")
			}

			if m.permisoRepo == nil {
				log.Printf("[AUTH] permisoRepo no inicializado")
				return c.Redirect(http.StatusSeeOther, "/dashboard?error=Error de configuración")
			}

			hasPermission, err := m.permisoRepo.UserHasPermission(userID, moduloNombre, permiso)
			if err != nil {
				log.Printf("[AUTH] Error verificando permiso %s para módulo %s: %v", permiso, moduloNombre, err)
				return c.Redirect(http.StatusSeeOther, "/dashboard?error=Error verificando permisos")
			}

			if !hasPermission {
				log.Printf("[AUTH] Usuario %d no tiene permiso %s en módulo %s", userID, permiso, moduloNombre)
				return c.Redirect(http.StatusSeeOther, "/dashboard?error=No tiene permiso para esta acción")
			}

			log.Printf("[AUTH] Usuario %d tiene permiso %s en módulo %s", userID, permiso, moduloNombre)
			return next(c)
		}
	}
}

// ============================================
// MIDDLEWARE DE CSRF
// ============================================

// CSRFProtection - Protege contra ataques CSRF
func CSRFProtection(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Métodos que no modifican datos
		if c.Request().Method == "GET" || c.Request().Method == "HEAD" || c.Request().Method == "OPTIONS" {
			return next(c)
		}

		// Obtener token del header o formulario
		token := c.Request().Header.Get("X-CSRF-Token")
		if token == "" {
			token = c.FormValue("csrf_token")
		}

		// Obtener token de la cookie
		cookie, err := c.Cookie("csrf_token")
		if err != nil || cookie.Value == "" {
			log.Printf("[CSRF] Token no encontrado en cookie")
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Token CSRF no encontrado"})
		}

		// Verificar token
		if token != cookie.Value {
			log.Printf("[CSRF] Token inválido: esperado %s, recibido %s", cookie.Value, token)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Token CSRF inválido"})
		}

		return next(c)
	}
}

// ============================================
// MIDDLEWARE DE LOGGING
// ============================================

// RequestLogger - Loggea todas las peticiones
func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		duration := time.Since(start)

		userID := c.Get("user_id")
		userEmail := c.Get("user_email")

		log.Printf("[REQUEST] %s %s | Status: %d | Duration: %v | User: %v (%v)",
			c.Request().Method,
			c.Request().URL.Path,
			c.Response().Status,
			duration,
			userID,
			userEmail,
		)

		return err
	}
}

// ============================================
// MIDDLEWARE DE RUTAS PÚBLICAS
// ============================================

// PublicRoutes - Define rutas que no requieren autenticación
func PublicRoutes(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		publicPaths := map[string]bool{
			"/":            true,
			"/login":       true,
			"/do-login":    true,
			"/register":    true,
			"/do-register": true,
			"/health":      true,
			"/maintenance": true,
			"/success":     true,
		}

		if publicPaths[c.Path()] {
			return next(c)
		}

		return next(c)
	}
}

// ============================================
// FUNCIONES AUXILIARES
// ============================================

func (m *AuthMiddleware) getRolNombre(roleID int) string {
	switch roleID {
	case 1:
		return "administrador"
	case 2:
		return "usuario"
	case 3:
		return "editor"
	default:
		return "usuario"
	}
}
