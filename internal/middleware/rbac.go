package middleware

import (
	"database/sql"
	"log"
	"net/http"
	"tu-proyecto/internal/repository"

	"github.com/labstack/echo/v4"
)

type RBACMiddleware struct {
	permisoRepo repository.PermisoRepository
}

func NewRBACMiddleware(permisoRepo repository.PermisoRepository) *RBACMiddleware {
	return &RBACMiddleware{
		permisoRepo: permisoRepo,
	}
}

func (m *RBACMiddleware) CheckPermission(moduloNombre string, permiso string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get("user_id").(int)
			if !ok {
				log.Printf("[RBAC] ERROR: No se pudo obtener user_id")
				return c.Redirect(http.StatusSeeOther, "/login?error=Sesión no válida")
			}

			// ============================================
			// LOGS DE DEPURACIÓN
			// ============================================
			log.Printf("[RBAC] ========== VERIFICANDO PERMISO ==========")
			log.Printf("[RBAC] UserID: %d", userID)
			log.Printf("[RBAC] Módulo: %s", moduloNombre)
			log.Printf("[RBAC] Permiso: %s", permiso)

			tienePermiso, err := m.permisoRepo.UserHasPermission(userID, moduloNombre, permiso)

			log.Printf("[RBAC] Resultado - tienePermiso: %v", tienePermiso)
			if err != nil {
				log.Printf("[RBAC] Error: %v", err)
			}

			if err != nil {
				if err == sql.ErrNoRows {
					log.Printf("[RBAC] ❌ ACCESO DENEGADO - No hay registro de permisos")
					return c.Redirect(http.StatusSeeOther, "/dashboard?error=No tienes permiso para acceder")
				}
				log.Printf("[RBAC] ❌ ERROR EN BD")
				return c.Redirect(http.StatusSeeOther, "/maintenance?error=Error verificando permisos")
			}

			if !tienePermiso {
				log.Printf("[RBAC] ❌ ACCESO DENEGADO - Usuario %d no tiene permiso %s en módulo %s", userID, permiso, moduloNombre)
				return c.Redirect(http.StatusSeeOther, "/dashboard?error=No tienes permiso para esta acción")
			}

			log.Printf("[RBAC] ✅ ACCESO PERMITIDO - Usuario %d tiene permiso %s en módulo %s", userID, permiso, moduloNombre)
			return next(c)
		}
	}
}

func (m *RBACMiddleware) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := c.Get("user_id").(int)
		if !ok {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		isAdmin, err := m.permisoRepo.IsAdmin(userID)
		if err != nil || !isAdmin {
			return c.Redirect(http.StatusSeeOther, "/dashboard?error=Área restringida a administradores")
		}

		return next(c)
	}
}

// RequireModuleAccess - Verifica acceso al módulo (permiso "ver")
func (m *RBACMiddleware) RequireModuleAccess(moduloNombre string) echo.MiddlewareFunc {
	return m.CheckPermission(moduloNombre, "ver")
}

// GetUserPermissions - Endpoint para obtener permisos del usuario (para frontend)
func (m *RBACMiddleware) GetUserPermissions(c echo.Context) error {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No autenticado"})
	}

	permisos, err := m.permisoRepo.GetUserPermissions(userID)
	if err != nil {
		log.Printf("[RBAC] Error obteniendo permisos para usuario %d: %v", userID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error cargando permisos"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":  true,
		"permisos": permisos,
	})
}
