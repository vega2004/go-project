package handlers

import (
	"log"
	"net/http"
	"strconv"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/service"

	"github.com/labstack/echo/v4"
)

type ModuloHandler struct {
	service service.ModuloService
}

func NewModuloHandler(s service.ModuloService) *ModuloHandler {
	return &ModuloHandler{service: s}
}

// ============================================
// FUNCIÓN AUXILIAR PARA OBTENER CATEGORÍAS
// ============================================
func (h *ModuloHandler) getCategoriasConError(c echo.Context) ([]string, string) {
	categorias, err := h.service.GetCategoriasDisponibles()
	if err != nil {
		log.Printf("[ERROR] Error obteniendo categorías: %v", err)
		return []string{"seguridad", "principal1", "principal2"}, "Error al cargar categorías"
	}
	return categorias, ""
}

// ============================================
// INDEX - Listar módulos
// ============================================
func (h *ModuloHandler) Index(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	filter := &models.ModuloFilter{
		Nombre:   c.QueryParam("nombre"),
		Page:     page,
		PageSize: 10,
	}

	result, err := h.service.GetAll(filter)
	if err != nil {
		log.Printf("[ERROR] ModuloHandler.Index: %v", err)
		return c.Render(http.StatusOK, "error.html", map[string]interface{}{
			"Title":   "Error",
			"Code":    500,
			"Message": "Error al cargar módulos",
		})
	}

	categorias, errCategorias := h.service.GetCategoriasDisponibles()
	if errCategorias != nil {
		log.Printf("[ERROR] Error obteniendo categorías: %v", errCategorias)
		categorias = []string{"seguridad", "principal1", "principal2"}
	}

	// Obtener CSRF token de forma segura
	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "seguridad/modulos.html", map[string]interface{}{
		"Title":      "Módulos",
		"Modulos":    result,
		"Filtros":    filter,
		"Categorias": categorias,
		"SuccessMsg": c.QueryParam("success"),
		"ErrorMsg":   c.QueryParam("error"),
		"CSRFToken":  csrfToken,
	})
}

// ============================================
// CREATE FORM - Mostrar formulario de creación
// ============================================
func (h *ModuloHandler) CreateForm(c echo.Context) error {
	categorias, errCategorias := h.service.GetCategoriasDisponibles()
	if errCategorias != nil {
		log.Printf("[ERROR] Error obteniendo categorías: %v", errCategorias)
		categorias = []string{"seguridad", "principal1", "principal2"}
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "seguridad/modulo_form.html", map[string]interface{}{
		"Title":      "Nuevo Módulo",
		"Modulo":     &models.Modulo{Orden: 0, Activo: true},
		"Categorias": categorias,
		"CSRFToken":  csrfToken,
	})
}

// ============================================
// CREATE - Crear nuevo módulo
// ============================================
func (h *ModuloHandler) Create(c echo.Context) error {
	// Validar usuario autenticado
	userID, ok := c.Get("user_id").(int)
	if !ok {
		log.Printf("[ERROR] No se pudo obtener user_id del contexto")
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=Error de autenticación")
	}

	userName, ok := c.Get("user_name").(string)
	if !ok {
		userName = "Desconocido"
	}

	activo := c.FormValue("activo") == "on"
	orden, _ := strconv.Atoi(c.FormValue("orden"))

	modulo := &models.Modulo{
		Nombre:        c.FormValue("nombre"),
		NombreMostrar: c.FormValue("nombre_mostrar"),
		Ruta:          c.FormValue("ruta"),
		Icono:         c.FormValue("icono"),
		Categoria:     c.FormValue("categoria"),
		Orden:         orden,
		Activo:        activo,
	}

	if err := h.service.Create(modulo); err != nil {
		log.Printf("[WARN] Usuario %d (%s) intentó crear módulo '%s' y falló: %v",
			userID, userName, modulo.Nombre, err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d (%s) creó módulo ID=%d (%s) - Ruta: %s",
		userID, userName, modulo.ID, modulo.Nombre, modulo.Ruta)

	return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?success=Módulo creado exitosamente")
}

// ============================================
// EDIT FORM - Mostrar formulario de edición
// ============================================
func (h *ModuloHandler) EditForm(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=ID inválido")
	}

	modulo, err := h.service.GetByID(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=Módulo no encontrado")
	}

	categorias, errCategorias := h.service.GetCategoriasDisponibles()
	if errCategorias != nil {
		log.Printf("[ERROR] Error obteniendo categorías: %v", errCategorias)
		categorias = []string{"seguridad", "principal1", "principal2"}
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "seguridad/modulo_form.html", map[string]interface{}{
		"Title":      "Editar Módulo",
		"Modulo":     modulo,
		"Categorias": categorias,
		"CSRFToken":  csrfToken,
	})
}

// ============================================
// UPDATE - Actualizar módulo
// ============================================
func (h *ModuloHandler) Update(c echo.Context) error {
	// Validar usuario autenticado
	userID, ok := c.Get("user_id").(int)
	if !ok {
		log.Printf("[ERROR] No se pudo obtener user_id del contexto")
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=Error de autenticación")
	}

	userName, ok := c.Get("user_name").(string)
	if !ok {
		userName = "Desconocido"
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=ID inválido")
	}

	activo := c.FormValue("activo") == "on"
	orden, _ := strconv.Atoi(c.FormValue("orden"))

	modulo := &models.Modulo{
		ID:            id,
		Nombre:        c.FormValue("nombre"),
		NombreMostrar: c.FormValue("nombre_mostrar"),
		Ruta:          c.FormValue("ruta"),
		Icono:         c.FormValue("icono"),
		Categoria:     c.FormValue("categoria"),
		Orden:         orden,
		Activo:        activo,
	}

	if err := h.service.Update(modulo); err != nil {
		log.Printf("[WARN] Usuario %d (%s) intentó actualizar módulo ID=%d y falló: %v",
			userID, userName, id, err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d (%s) actualizó módulo ID=%d (%s)",
		userID, userName, id, modulo.Nombre)

	return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?success=Módulo actualizado exitosamente")
}

// ============================================
// DELETE - Eliminar módulo (AJAX)
// ============================================
func (h *ModuloHandler) Delete(c echo.Context) error {
	// Validar usuario autenticado
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No autenticado"})
	}

	userName, ok := c.Get("user_name").(string)
	if !ok {
		userName = "Desconocido"
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	if err := h.service.Delete(id); err != nil {
		log.Printf("[WARN] Usuario %d (%s) intentó eliminar módulo ID=%d y falló: %v",
			userID, userName, id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	log.Printf("[AUDIT] Usuario %d (%s) eliminó módulo ID=%d",
		userID, userName, id)

	return c.JSON(http.StatusOK, map[string]string{"message": "Módulo eliminado exitosamente"})
}
