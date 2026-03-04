package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	// servicios necesarios
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) ShowDashboard(c echo.Context) error {
	log.Println("🚦 INICIANDO ShowDashboard")

	// Obtener datos de la sesión
	userName := c.Get("user_name")
	userEmail := c.Get("user_email")
	userRole := c.Get("user_role")
	userID := c.Get("user_id")

	// LOGS PARA DIAGNÓSTICO
	log.Printf("📊 Datos de sesión recibidos:")
	log.Printf("   - UserID: %v", userID)
	log.Printf("   - UserName: %v", userName)
	log.Printf("   - UserEmail: %v", userEmail)
	log.Printf("   - UserRole: %v", userRole)

	// Verificar si los datos existen
	if userName == nil {
		log.Println("❌ ERROR: userName es nil - la sesión no tiene datos")
		return c.Redirect(http.StatusSeeOther, "/login?error=Sesión inválida")
	}

	if userEmail == nil {
		log.Println("⚠️ ADVERTENCIA: userEmail es nil - se usará valor por defecto")
		userEmail = "correo@no.disponible"
	}

	if userRole == nil {
		log.Println("⚠️ ADVERTENCIA: userRole es nil - se usará 'user' por defecto")
		userRole = "user"
	}

	// Datos de ejemplo para estadísticas (reemplazar con datos reales de BD)
	stats := map[string]interface{}{
		"TotalPersonas": 150,
		"TotalImagenes": 45,
		"TotalUsuarios": 230,
		"ActividadHoy":  18,
	}

	// Actividad reciente (ejemplo)
	recentActivity := []map[string]string{
		{"User": "Admin", "Action": "creó un nuevo usuario", "Time": "hace 5 min"},
		{"User": "Juan", "Action": "subió una imagen", "Time": "hace 15 min"},
	}

	// Fecha actual formateada
	currentDate := time.Now().Format("02/01/2006 15:04")
	lastAccess := time.Now().Format("02/01/2006 15:04")

	// Preparar datos para el template
	data := map[string]interface{}{
		"Title":          "Dashboard Principal",
		"UserName":       userName,
		"UserEmail":      userEmail,
		"UserRole":       userRole,
		"Stats":          stats,
		"RecentActivity": recentActivity,
		"CurrentDate":    currentDate,
		"LastAccess":     lastAccess,
		"breadcrumbs":    c.Get("breadcrumbs"),
	}

	log.Printf("📦 Datos enviados al template: %+v", data)

	// Renderizar el template
	err := c.Render(http.StatusOK, "dashboard.html", data)
	if err != nil {
		log.Printf("❌ Error al renderizar dashboard.html: %v", err)
		return c.Redirect(http.StatusSeeOther, "/maintenance")
	}

	log.Println("✅ Dashboard renderizado correctamente")
	return nil
}
