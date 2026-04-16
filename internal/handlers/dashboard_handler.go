package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

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

	userPerfil := c.Get("user_perfil") // ← Cambiado de user_role a user_perfil
	if userPerfil == nil {
		userPerfil = "usuario"
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
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

	modulos := h.getModulosByPerfil(userPerfil.(string)) // ← Cambiado de getModulosByRole a getModulosByPerfil

	log.Printf("[INFO] Dashboard cargado para usuario %d (%s) con perfil %s", userID, userName, userPerfil)

	return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
		"Title":          "Dashboard Principal",
		"UserID":         userID,
		"UserName":       userName,
		"UserEmail":      userEmail,
		"UserPerfil":     userPerfil, // ← Cambiado de UserRole a UserPerfil
		"Stats":          stats,
		"RecentActivity": recentActivity,
		"Modulos":        modulos,
		"CurrentDate":    time.Now().Format("02/01/2006"),
		"CurrentTime":    time.Now().Format("15:04:05"),
		"CSRFToken":      csrfToken,
		"breadcrumbs": []map[string]string{
			{"name": "Inicio", "url": "/"},
			{"name": "Dashboard", "url": "/dashboard"},
		},
	})
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

// getModulosByPerfil - Obtiene módulos según el perfil del usuario
func (h *DashboardHandler) getModulosByPerfil(perfil string) []map[string]interface{} {
	modulos := []map[string]interface{}{}

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

	// Si es administrador, mostrar módulos de seguridad
	if perfil == "administrador" {
		modulos = append(modulos, []map[string]interface{}{
			{"nombre": "Perfiles", "ruta": "/seguridad/perfiles", "icono": "bi-person-badge", "color": "danger"},
			{"nombre": "Módulos", "ruta": "/seguridad/modulos", "icono": "bi-grid-3x3", "color": "warning"},
			{"nombre": "Permisos", "ruta": "/seguridad/permisos-perfil", "icono": "bi-shield-lock", "color": "dark"},
			{"nombre": "Usuarios", "ruta": "/seguridad/usuarios", "icono": "bi-people", "color": "success"},
		}...)
	}

	// Módulos principales (siempre visibles)
	modulos = append(modulos, []map[string]interface{}{
		{"nombre": "Principal 1.1", "ruta": "/principal/clientes", "icono": "bi-building", "color": "secondary"},
		{"nombre": "Principal 1.2", "ruta": "/principal/productos", "icono": "bi-box", "color": "secondary"},
		{"nombre": "Principal 2.1", "ruta": "/principal/facturas", "icono": "bi-receipt", "color": "secondary"},
		{"nombre": "Principal 2.2", "ruta": "/principal/proveedores", "icono": "bi-truck", "color": "secondary"},
	}...)

	return modulos
}
