package middleware

import (
	"log"
	"net/http"
	"time"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/repository"
	"tu-proyecto/internal/utils"

	"github.com/labstack/echo/v4"
)

// ============================================
// ESTRUCTURA PRINCIPAL
// ============================================

type AuthMiddleware struct {
	sessionManager *utils.SessionManager
	jwtManager     *utils.JWTManager
	authRepo       repository.AuthRepository
	permisoRepo    repository.PermisoRepository
	useJWT         bool
}

// NewAuthMiddleware - Constructor
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
// MIDDLEWARE PRINCIPAL
// ============================================

// RequireAuth - Verifica autenticación (soporta Sesión + JWT)
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// ============================================
		// 1. AUTENTICACIÓN POR JWT
		// ============================================
		if m.useJWT && m.jwtManager != nil {
			tokenString, err := m.jwtManager.ExtractFromRequest(c)
			if err == nil {
				claims, err := m.jwtManager.Validate(tokenString)
				if err == nil {
					user, err := m.authRepo.FindByID(claims.UserID)
					if err == nil && user.Activo {
						c.Set("user_id", claims.UserID)
						c.Set("user_name", claims.Name)
						c.Set("user_email", claims.Email)
						c.Set("user_perfil_id", claims.PerfilID)
						c.Set("user_perfil", claims.PerfilNombre)
						c.Set("is_authenticated", true)

						permisos, _ := m.permisoRepo.GetPermisosByPerfil(claims.PerfilID)
						c.Set("permisos", permisos)

						log.Printf("[AUTH] Usuario %d autenticado por JWT", claims.UserID)
						return next(c)
					}
				}
			}
		}

		// ============================================
		// 2. AUTENTICACIÓN POR SESIÓN
		// ============================================
		session, err := m.sessionManager.GetSession(c)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login?error=Debe iniciar sesión")
		}

		user, err := m.authRepo.FindByID(session.UserID)
		if err != nil {
			m.sessionManager.ClearSession(c)
			return c.Redirect(http.StatusSeeOther, "/login?error=Usuario no encontrado")
		}

		if !user.Activo {
			m.sessionManager.ClearSession(c)
			return c.Redirect(http.StatusSeeOther, "/login?error=Usuario desactivado")
		}

		if session.PerfilID != user.PerfilID {
			session.PerfilID = user.PerfilID
			session.PerfilNombre = m.getPerfilNombre(user.PerfilID)
			m.sessionManager.UpdateSession(c, session)
		}

		session.LastActivity = time.Now()

		c.Set("user_id", session.UserID)
		c.Set("user_name", session.Name)
		c.Set("user_email", session.Email)
		c.Set("user_perfil_id", session.PerfilID)
		c.Set("user_perfil", session.PerfilNombre)
		c.Set("is_authenticated", true)

		permisos, err := m.permisoRepo.GetPermisosByPerfil(session.PerfilID)
		if err != nil {
			permisos = make(map[string]models.Permiso)
		}
		c.Set("permisos", permisos)

		log.Printf("[AUTH] Usuario %d autenticado por sesión", session.UserID)

		return next(c)
	}
}

// ============================================
// FUNCIONES AUXILIARES
// ============================================

func (m *AuthMiddleware) getPerfilNombre(perfilID int) string {
	switch perfilID {
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
