package handlers

import (
	"log"
	"net/http"
	"tu-proyecto/internal/service"

	"github.com/labstack/echo/v4"
)

type PrincipalHandler struct {
	permisoService service.PermisoService
}

func NewPrincipalHandler(ps service.PermisoService) *PrincipalHandler {
	return &PrincipalHandler{
		permisoService: ps,
	}
}

// getPermisosSeguro - Obtiene permisos de forma segura
func (h *PrincipalHandler) getPermisosSeguro(c echo.Context, moduloNombre string) map[string]bool {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return map[string]bool{
			"puede_ver":      true,
			"puede_crear":    false,
			"puede_editar":   false,
			"puede_eliminar": false,
			"puede_detalle":  false,
		}
	}

	permisos := map[string]bool{
		"puede_ver":      true,
		"puede_crear":    false,
		"puede_editar":   false,
		"puede_eliminar": false,
		"puede_detalle":  false,
	}

	if h.permisoService != nil {
		puedeCrear, _ := h.permisoService.UserHasPermission(userID, moduloNombre, "crear")
		puedeEditar, _ := h.permisoService.UserHasPermission(userID, moduloNombre, "editar")
		puedeEliminar, _ := h.permisoService.UserHasPermission(userID, moduloNombre, "eliminar")
		puedeDetalle, _ := h.permisoService.UserHasPermission(userID, moduloNombre, "detalle")

		permisos["puede_crear"] = puedeCrear
		permisos["puede_editar"] = puedeEditar
		permisos["puede_eliminar"] = puedeEliminar
		permisos["puede_detalle"] = puedeDetalle
	}

	return permisos
}

// renderPrincipalTemplate - Función auxiliar para renderizar templates
func (h *PrincipalHandler) renderPrincipalTemplate(c echo.Context, title, moduloNombre, descripcion, templateName string) error {
	userID := c.Get("user_id").(int)
	userName := c.Get("user_name").(string)
	userRole := c.Get("user_role").(string)

	permisos := h.getPermisosSeguro(c, moduloNombre)

	log.Printf("[AUDIT] Usuario %d (%s) accedió al módulo %s", userID, userName, moduloNombre)

	return c.Render(http.StatusOK, templateName, map[string]interface{}{
		"Title":             title,
		"ModuloNombre":      moduloNombre,
		"ModuloDescripcion": descripcion,
		"UserName":          userName,
		"UserRole":          userRole,
		"Permisos":          permisos,
		"breadcrumbs": []map[string]string{
			{"name": "Inicio", "url": "/"},
			{"name": "Dashboard", "url": "/dashboard"},
			{"name": moduloNombre, "url": ""},
		},
	})
}

// Principal11 - Módulo Principal 1.1
func (h *PrincipalHandler) Principal11(c echo.Context) error {
	return h.renderPrincipalTemplate(c,
		"Principal 1.1 - Clientes",
		"principal11",
		"Pantalla estática con acciones visibles según permisos.",
		"principal/principal11.html",
	)
}

// Principal12 - Módulo Principal 1.2
func (h *PrincipalHandler) Principal12(c echo.Context) error {
	return h.renderPrincipalTemplate(c,
		"Principal 1.2 - Productos",
		"principal12",
		"Pantalla estática con acciones visibles según permisos.",
		"principal/principal12.html",
	)
}

// Principal21 - Módulo Principal 2.1
func (h *PrincipalHandler) Principal21(c echo.Context) error {
	return h.renderPrincipalTemplate(c,
		"Principal 2.1 - Facturas",
		"principal21",
		"Pantalla estática con acciones visibles según permisos.",
		"principal/principal21.html",
	)
}

// Principal22 - Módulo Principal 2.2
func (h *PrincipalHandler) Principal22(c echo.Context) error {
	return h.renderPrincipalTemplate(c,
		"Principal 2.2 - Proveedores",
		"principal22",
		"Pantalla estática con acciones visibles según permisos.",
		"principal/principal22.html",
	)
}
