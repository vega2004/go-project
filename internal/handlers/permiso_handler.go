package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"tu-proyecto/internal/config"
	"tu-proyecto/internal/middleware"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/service"

	"github.com/labstack/echo/v4"
)

type PermisoHandler struct {
	permisoService service.PermisoService
	perfilService  service.PerfilService
	env            *config.Env
	csrfMiddleware *middleware.CSRFMiddleware
}

func NewPermisoHandler(ps service.PermisoService, pfs service.PerfilService, env *config.Env) *PermisoHandler {
	return &PermisoHandler{
		permisoService: ps,
		perfilService:  pfs,
		env:            env,
		csrfMiddleware: middleware.NewCSRFMiddleware(env.IsProduction()),
	}
}

// Index - Muestra la página principal de gestión de permisos
// Index - Muestra la página principal de gestión de permisos
func (h *PermisoHandler) Index(c echo.Context) error {
	// Generar token CSRF
	h.csrfMiddleware.SetToken(c)

	// ✅ Obtener permisos del contexto
	permisos := c.Get("permisos")
	if permisos == nil {
		permisos = make(map[string]models.Permiso)
	}

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
		"Permisos":   permisos, // ✅ AGREGAR ESTO
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

// SavePermissions - Guarda los permisos de un perfil (CORREGIDO para JSON)
func (h *PermisoHandler) SavePermissions(c echo.Context) error {
	// Obtener usuario que realiza la acción (para auditoría)
	auditorID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "No autenticado",
		})
	}

	// Leer el body completo para depurar
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("[ERROR] Error leyendo body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Error al leer datos",
		})
	}
	log.Printf("[DEBUG] Body recibido: %s", string(body))

	// Parsear JSON
	var req struct {
		PerfilID  int                         `json:"perfil_id"`
		Permisos  []models.PermisoItemRequest `json:"permisos"`
		CsrfToken string                      `json:"csrf_token"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("[ERROR] Error parseando JSON: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Error al parsear datos: " + err.Error(),
		})
	}

	log.Printf("[DEBUG] PerfilID: %d", req.PerfilID)
	log.Printf("[DEBUG] Cantidad de permisos: %d", len(req.Permisos))

	if req.PerfilID <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ID de perfil inválido",
		})
	}

	// Guardar permisos
	err = h.permisoService.SavePermissions(req.PerfilID, req.Permisos, auditorID)
	if err != nil {
		log.Printf("[ERROR] SavePermissions: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	log.Printf("[AUDIT] Usuario %d guardó permisos para perfil %d", auditorID, req.PerfilID)

	return c.JSON(http.StatusOK, map[string]string{
		"success": "Permisos guardados exitosamente",
	})
}

// parsePermisosForm - Ya no se usa con JSON, pero lo mantenemos por si acaso
func (h *PermisoHandler) parsePermisosForm(c echo.Context) ([]models.PermisoItemRequest, error) {
	moduloIDs := c.FormValue("modulo_ids")
	if moduloIDs == "" {
		return nil, fmt.Errorf("no se recibieron módulos")
	}

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
