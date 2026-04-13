package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/service"

	"github.com/labstack/echo/v4"
)

type PermisoHandler struct {
	permisoService service.PermisoService
	perfilService  service.PerfilService
}

func NewPermisoHandler(ps service.PermisoService, pfs service.PerfilService) *PermisoHandler {
	return &PermisoHandler{
		permisoService: ps,
		perfilService:  pfs,
	}
}

// Index - Muestra la página principal de gestión de permisos
func (h *PermisoHandler) Index(c echo.Context) error {
	// Obtener lista de perfiles para el selector
	perfiles, err := h.perfilService.GetAll(&models.PerfilFilter{Page: 1, PageSize: 100})
	if err != nil {
		log.Printf("[ERROR] PermisoHandler.Index: %v", err)
		return c.Render(http.StatusOK, "error.html", map[string]interface{}{
			"Title":   "Error",
			"Code":    500,
			"Message": "Error al cargar perfiles",
		})
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "seguridad/permisos.html", map[string]interface{}{
		"Title":      "Permisos por Perfil",
		"Perfiles":   perfiles.Data,
		"SuccessMsg": c.QueryParam("success"),
		"ErrorMsg":   c.QueryParam("error"),
		"CSRFToken":  csrfToken,
	})
}

// LoadPermissions - Carga los permisos de un perfil (AJAX)
func (h *PermisoHandler) LoadPermissions(c echo.Context) error {
	perfilID, err := strconv.Atoi(c.FormValue("perfil_id"))
	if err != nil || perfilID <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ID de perfil inválido",
		})
	}

	// Obtener permisos del perfil
	result, err := h.permisoService.GetPermissionsByPerfil(perfilID)
	if err != nil {
		log.Printf("[ERROR] LoadPermissions: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Error al cargar permisos",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// SavePermissions - Guarda los permisos de un perfil
func (h *PermisoHandler) SavePermissions(c echo.Context) error {
	// Obtener usuario que realiza la acción (para auditoría)
	auditorID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "No autenticado",
		})
	}

	// Parsear perfil_id
	perfilID, err := strconv.Atoi(c.FormValue("perfil_id"))
	if err != nil || perfilID <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ID de perfil inválido",
		})
	}

	// Parsear permisos del formulario
	permisos, err := h.parsePermisosForm(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Guardar permisos
	err = h.permisoService.SavePermissions(perfilID, permisos, auditorID)
	if err != nil {
		log.Printf("[ERROR] SavePermissions: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	log.Printf("[AUDIT] Usuario %d guardó permisos para perfil %d", auditorID, perfilID)

	return c.JSON(http.StatusOK, map[string]string{
		"success": "Permisos guardados exitosamente",
	})
}

// parsePermisosForm - Parsea los permisos enviados desde el formulario
func (h *PermisoHandler) parsePermisosForm(c echo.Context) ([]models.PermisoItemRequest, error) {
	// Obtener los IDs de módulo desde el formulario
	moduloIDs := c.FormValue("modulo_ids")
	if moduloIDs == "" {
		return nil, fmt.Errorf("no se recibieron módulos")
	}

	// Parsear IDs (vienen como "1,2,3,4")
	idStrings := strings.Split(moduloIDs, ",")
	var permisos []models.PermisoItemRequest

	for _, idStr := range idStrings {
		moduloID, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			continue
		}

		permiso := models.PermisoItemRequest{
			ModuloID:      moduloID,
			PuedeVer:      c.FormValue(fmt.Sprintf("ver_%d", moduloID)) == "on",
			PuedeCrear:    c.FormValue(fmt.Sprintf("crear_%d", moduloID)) == "on",
			PuedeEditar:   c.FormValue(fmt.Sprintf("editar_%d", moduloID)) == "on",
			PuedeEliminar: c.FormValue(fmt.Sprintf("eliminar_%d", moduloID)) == "on",
			PuedeDetalle:  c.FormValue(fmt.Sprintf("detalle_%d", moduloID)) == "on",
		}

		permisos = append(permisos, permiso)
	}

	return permisos, nil
}
