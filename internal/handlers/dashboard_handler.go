package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"
	"tu-proyecto/internal/models"

	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	db *sql.DB
}

func NewDashboardHandler(db *sql.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// ShowDashboard - Muestra el panel principal después del login
func (h *DashboardHandler) ShowDashboard(c echo.Context) error {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		log.Println("[ERROR] Dashboard: No se pudo obtener user_id del contexto")
		return c.Redirect(http.StatusSeeOther, "/login?error=Sesión inválida")
	}

	userName := c.Get("user_name")
	if userName == nil {
		userName = "Usuario"
	}

	userEmail := c.Get("user_email")
	if userEmail == nil {
		userEmail = "correo@no.disponible"
	}

	userPerfil := c.Get("user_perfil")
	if userPerfil == nil {
		userPerfil = "usuario"
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	// ✅ Obtener permisos REALES del contexto (cargados por el middleware de autenticación)
	permisos, ok := c.Get("permisos").(map[string]models.Permiso)
	if !ok {
		permisos = make(map[string]models.Permiso)
	}

	// Obtener estadísticas reales de la BD
	stats := map[string]interface{}{
		"TotalUsuarios":   h.getTotalUsers(),
		"TotalPerfiles":   h.getTotalPerfiles(),
		"TotalModulos":    h.getTotalModulos(),
		"UsuariosActivos": h.getActiveUsers(),
		"UltimoAcceso":    time.Now().Format("02/01/2006 15:04:05"),
	}

	recentActivity := []map[string]interface{}{
		{"usuario": userName, "accion": "Inició sesión", "tiempo": "Hace unos momentos"},
	}

	// ✅ Obtener módulos filtrados por permisos REALES
	modulos := h.getModulosConPermisos(permisos)

	log.Printf("[INFO] Dashboard cargado para usuario %d (%s) con perfil %s", userID, userName, userPerfil)

	return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
		"Title":          "Dashboard Principal",
		"UserID":         userID,
		"UserName":       userName,
		"UserEmail":      userEmail,
		"UserPerfil":     userPerfil,
		"Stats":          stats,
		"RecentActivity": recentActivity,
		"Modulos":        modulos,
		"Permisos":       permisos, // ✅ Pasamos los permisos al template
		"CurrentDate":    time.Now().Format("02/01/2006"),
		"CurrentTime":    time.Now().Format("15:04:05"),
		"CSRFToken":      csrfToken,
		"breadcrumbs": []map[string]string{
			{"name": "Inicio", "url": "/"},
			{"name": "Dashboard", "url": "/dashboard"},
		},
	})
}

// ✅ NUEVO: Obtener módulos según permisos reales
func (h *DashboardHandler) getModulosConPermisos(permisos map[string]models.Permiso) []map[string]interface{} {
	modulos := []map[string]interface{}{}

	// Dashboard y Mi Perfil siempre visibles
	modulos = append(modulos, map[string]interface{}{
		"nombre": "Dashboard",
		"ruta":   "/dashboard",
		"icono":  "bi-speedometer2",
		"color":  "primary",
	})
	modulos = append(modulos, map[string]interface{}{
		"nombre": "Mi Perfil",
		"ruta":   "/perfil",
		"icono":  "bi-person-circle",
		"color":  "info",
	})

	// Módulos de Seguridad - solo si tiene permiso de ver
	modulosSeguridad := []struct {
		nombre string
		ruta   string
		icono  string
		color  string
	}{
		{"Perfiles", "/seguridad/perfiles", "bi-person-badge", "danger"},
		{"Módulos", "/seguridad/modulos", "bi-grid-3x3", "warning"},
		{"Permisos", "/seguridad/permisos-perfil", "bi-shield-lock", "dark"},
		{"Usuarios", "/seguridad/usuarios", "bi-people", "success"},
	}

	for _, m := range modulosSeguridad {
		if p, ok := permisos[m.nombre]; ok && p.PuedeVer {
			modulos = append(modulos, map[string]interface{}{
				"nombre": m.nombre,
				"ruta":   m.ruta,
				"icono":  m.icono,
				"color":  m.color,
			})
		}
	}

	// Módulos principales - solo si tiene permiso de ver
	modulosPrincipales := []struct {
		nombre string
		ruta   string
		icono  string
		color  string
	}{
		{"Principal 1.1", "/principal/clientes", "bi-building", "secondary"},
		{"Principal 1.2", "/principal/productos", "bi-box", "secondary"},
		{"Principal 2.1", "/principal/facturas", "bi-receipt", "secondary"},
		{"Principal 2.2", "/principal/proveedores", "bi-truck", "secondary"},
	}

	for _, m := range modulosPrincipales {
		if p, ok := permisos[m.nombre]; ok && p.PuedeVer {
			modulos = append(modulos, map[string]interface{}{
				"nombre": m.nombre,
				"ruta":   m.ruta,
				"icono":  m.icono,
				"color":  m.color,
			})
		}
	}

	return modulos
}

func (h *DashboardHandler) getTotalUsers() int {
	if h.db == nil {
		return 0
	}
	var count int
	h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count
}

func (h *DashboardHandler) getTotalPerfiles() int {
	if h.db == nil {
		return 3
	}
	var count int
	h.db.QueryRow("SELECT COUNT(*) FROM perfiles").Scan(&count)
	return count
}

func (h *DashboardHandler) getTotalModulos() int {
	if h.db == nil {
		return 8
	}
	var count int
	h.db.QueryRow("SELECT COUNT(*) FROM modulos").Scan(&count)
	return count
}

func (h *DashboardHandler) getActiveUsers() int {
	if h.db == nil {
		return 0
	}
	var count int
	h.db.QueryRow("SELECT COUNT(*) FROM users WHERE activo = true").Scan(&count)
	return count
}
