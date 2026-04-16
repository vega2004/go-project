package handlers

import (
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

type ModuloHandler struct {
	service        service.ModuloService
	env            *config.Env
	csrfMiddleware *middleware.CSRFMiddleware
}

func NewModuloHandler(s service.ModuloService, env *config.Env) *ModuloHandler {
	log.Println("[DEBUG] ModuloHandler inicializado")
	return &ModuloHandler{
		service:        s,
		env:            env,
		csrfMiddleware: middleware.NewCSRFMiddleware(env.IsProduction()),
	}
}

// ============================================
// INDEX - Listar módulos
// ============================================
// ============================================
// INDEX - Listar módulos
// ============================================
func (h *ModuloHandler) Index(c echo.Context) error {
	log.Println("[DEBUG] ModuloHandler.Index - Iniciando")
	h.csrfMiddleware.SetToken(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	filter := &models.ModuloFilter{
		Nombre:   c.QueryParam("nombre"),
		Page:     page,
		PageSize: 10,
	}

	log.Printf("[DEBUG] Index - Filtros: Nombre=%s, Page=%d", filter.Nombre, filter.Page)

	result, err := h.service.GetAll(filter)
	if err != nil {
		log.Printf("[ERROR] ModuloHandler.Index: %v", err)
		return c.Render(http.StatusOK, "error.html", map[string]interface{}{
			"Title":   "Error",
			"Code":    500,
			"Message": "Error al cargar módulos",
		})
	}

	// ✅ FILTRAR: Excluir módulos "Inicio", "Seguridad", "Principal 1" y "Principal 2"
	modulosExcluidos := map[string]bool{
		"Inicio":      true,
		"Seguridad":   true,
		"Principal 1": true,
		"Principal 2": true,
	}

	filteredData := []models.Modulo{}
	for _, m := range result.Data {
		if !modulosExcluidos[m.Nombre] {
			filteredData = append(filteredData, m)
		}
	}

	result.Data = filteredData
	result.Total = len(filteredData)

	log.Printf("[DEBUG] Index - Módulos encontrados (después de filtrar): %d", result.Total)

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	permisos := c.Get("permisos")
	if permisos == nil {
		permisos = make(map[string]models.Permiso)
	}

	return c.Render(http.StatusOK, "seguridad/modulos.html", map[string]interface{}{
		"Title":      "Módulos",
		"Modulos":    result,
		"Filtros":    filter,
		"SuccessMsg": c.QueryParam("success"),
		"ErrorMsg":   c.QueryParam("error"),
		"CSRFToken":  csrfToken,
		"Permisos":   permisos,
	})
}

// ============================================
// CREATE FORM - Mostrar formulario de creación
// ============================================
func (h *ModuloHandler) CreateForm(c echo.Context) error {
	log.Println("[DEBUG] ModuloHandler.CreateForm - Iniciando")
	h.csrfMiddleware.SetToken(c)

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	modulo := &models.Modulo{
		Activo: true,
	}

	log.Printf("[DEBUG] CreateForm - Mostrando formulario para nuevo módulo (Activo por defecto: %v)", modulo.Activo)

	return c.Render(http.StatusOK, "seguridad/modulo_form.html", map[string]interface{}{
		"Title":     "Nuevo Módulo",
		"Modulo":    modulo,
		"CSRFToken": csrfToken,
	})
}

// ============================================
// CREATE - Crear nuevo módulo
// ============================================
func (h *ModuloHandler) Create(c echo.Context) error {
	log.Println("[DEBUG] ========================================")
	log.Println("[DEBUG] ModuloHandler.Create - INICIANDO CREACIÓN")
	log.Println("[DEBUG] ========================================")

	// ✅ LOG 1: Verificar autenticación
	userID, ok := c.Get("user_id").(int)
	if !ok {
		log.Printf("[ERROR] ❌ No se pudo obtener user_id del contexto")
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=Error de autenticación")
	}
	log.Printf("[DEBUG] ✅ UserID obtenido: %d", userID)

	userName, ok := c.Get("user_name").(string)
	if !ok {
		userName = "Desconocido"
	}
	log.Printf("[DEBUG] ✅ UserName obtenido: %s", userName)

	// ✅ LOG 2: Verificar datos recibidos del formulario
	log.Println("[DEBUG] --- DATOS RECIBIDOS DEL FORMULARIO ---")
	log.Printf("[DEBUG]   nombre:      '%s'", c.FormValue("nombre"))
	log.Printf("[DEBUG]   descripcion: '%s'", c.FormValue("descripcion"))
	log.Printf("[DEBUG]   activo:      '%s'", c.FormValue("activo"))
	log.Println("[DEBUG] ---------------------------------------")

	nombre := strings.TrimSpace(c.FormValue("nombre"))
	descripcion := strings.TrimSpace(c.FormValue("descripcion"))
	activo := c.FormValue("activo") == "on"

	log.Printf("[DEBUG] --- DATOS PROCESADOS ---")
	log.Printf("[DEBUG]   Nombre:      '%s' (longitud: %d)", nombre, len(nombre))
	log.Printf("[DEBUG]   Descripcion: '%s'", descripcion)
	log.Printf("[DEBUG]   Activo:      %v", activo)
	log.Println("[DEBUG] -------------------------")

	// ✅ LOG 3: Validaciones
	if nombre == "" {
		log.Printf("[WARN] ⚠️ Validación fallida: nombre vacío")
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=El nombre del módulo es requerido")
	}
	if len(nombre) < 3 {
		log.Printf("[WARN] ⚠️ Validación fallida: nombre muy corto (%d caracteres)", len(nombre))
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=El nombre debe tener al menos 3 caracteres")
	}
	log.Println("[DEBUG] ✅ Validaciones pasadas correctamente")

	// ✅ LOG 4: Crear objeto módulo
	modulo := &models.Modulo{
		Nombre:      nombre,
		Descripcion: descripcion,
		Activo:      activo,
	}
	log.Printf("[DEBUG] 📦 Objeto Modulo creado: %+v", modulo)

	// ✅ LOG 5: Llamar al servicio
	log.Println("[DEBUG] 🔄 Llamando a h.service.Create(modulo)...")
	if err := h.service.Create(modulo); err != nil {
		log.Printf("[ERROR] ❌ Error en service.Create: %v", err)
		log.Printf("[WARN] Usuario %d (%s) intentó crear módulo '%s' y falló: %v",
			userID, userName, nombre, err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error="+err.Error())
	}

	// ✅ LOG 6: Éxito
	log.Println("[DEBUG] ✅ service.Create ejecutado SIN ERRORES")
	log.Printf("[DEBUG] 📌 ID asignado al módulo: %d", modulo.ID)
	log.Printf("[AUDIT] ✅✅✅ Usuario %d (%s) creó módulo ID=%d (%s) - Activo=%v",
		userID, userName, modulo.ID, modulo.Nombre, modulo.Activo)
	log.Println("[DEBUG] ========================================")

	return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?success=Módulo creado exitosamente")
}

// ============================================
// EDIT FORM - Mostrar formulario de edición
// ============================================
func (h *ModuloHandler) EditForm(c echo.Context) error {
	log.Println("[DEBUG] ModuloHandler.EditForm - Iniciando")
	h.csrfMiddleware.SetToken(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("[ERROR] EditForm - ID inválido: %v", err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=ID inválido")
	}

	log.Printf("[DEBUG] EditForm - Buscando módulo ID: %d", id)

	modulo, err := h.service.GetByID(id)
	if err != nil {
		log.Printf("[ERROR] EditForm - Módulo no encontrado: %v", err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=Módulo no encontrado")
	}

	log.Printf("[DEBUG] EditForm - Módulo encontrado: ID=%d, Nombre=%s, Activo=%v",
		modulo.ID, modulo.Nombre, modulo.Activo)

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "seguridad/modulo_form.html", map[string]interface{}{
		"Title":     "Editar Módulo",
		"Modulo":    modulo,
		"CSRFToken": csrfToken,
	})
}

// ============================================
// UPDATE - Actualizar módulo
// ============================================
func (h *ModuloHandler) Update(c echo.Context) error {
	log.Println("[DEBUG] ========================================")
	log.Println("[DEBUG] ModuloHandler.Update - INICIANDO ACTUALIZACIÓN")
	log.Println("[DEBUG] ========================================")

	userID, ok := c.Get("user_id").(int)
	if !ok {
		log.Printf("[ERROR] ❌ No se pudo obtener user_id del contexto")
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=Error de autenticación")
	}
	log.Printf("[DEBUG] ✅ UserID: %d", userID)

	userName, ok := c.Get("user_name").(string)
	if !ok {
		userName = "Desconocido"
	}
	log.Printf("[DEBUG] ✅ UserName: %s", userName)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("[ERROR] Update - ID inválido: %v", err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=ID inválido")
	}
	log.Printf("[DEBUG] 📌 ID a actualizar: %d", id)

	// Datos del formulario
	log.Println("[DEBUG] --- DATOS RECIBIDOS DEL FORMULARIO ---")
	log.Printf("[DEBUG]   nombre:      '%s'", c.FormValue("nombre"))
	log.Printf("[DEBUG]   descripcion: '%s'", c.FormValue("descripcion"))
	log.Printf("[DEBUG]   activo:      '%s'", c.FormValue("activo"))
	log.Println("[DEBUG] ---------------------------------------")

	nombre := strings.TrimSpace(c.FormValue("nombre"))
	descripcion := strings.TrimSpace(c.FormValue("descripcion"))
	activo := c.FormValue("activo") == "on"

	log.Printf("[DEBUG] --- DATOS PROCESADOS ---")
	log.Printf("[DEBUG]   ID:          %d", id)
	log.Printf("[DEBUG]   Nombre:      '%s'", nombre)
	log.Printf("[DEBUG]   Descripcion: '%s'", descripcion)
	log.Printf("[DEBUG]   Activo:      %v", activo)
	log.Println("[DEBUG] -------------------------")

	if nombre == "" {
		log.Printf("[WARN] ⚠️ Validación fallida: nombre vacío")
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=El nombre del módulo es requerido")
	}
	if len(nombre) < 3 {
		log.Printf("[WARN] ⚠️ Validación fallida: nombre muy corto")
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error=El nombre debe tener al menos 3 caracteres")
	}

	modulo := &models.Modulo{
		ID:          id,
		Nombre:      nombre,
		Descripcion: descripcion,
		Activo:      activo,
	}
	log.Printf("[DEBUG] 📦 Objeto Modulo para actualizar: %+v", modulo)

	log.Println("[DEBUG] 🔄 Llamando a h.service.Update(modulo)...")
	if err := h.service.Update(modulo); err != nil {
		log.Printf("[ERROR] ❌ Error en service.Update: %v", err)
		log.Printf("[WARN] Usuario %d (%s) intentó actualizar módulo ID=%d y falló: %v",
			userID, userName, id, err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?error="+err.Error())
	}

	log.Printf("[AUDIT] ✅✅✅ Usuario %d (%s) actualizó módulo ID=%d (%s) - Activo=%v",
		userID, userName, id, modulo.Nombre, modulo.Activo)
	log.Println("[DEBUG] ========================================")

	return c.Redirect(http.StatusSeeOther, "/seguridad/modulos?success=Módulo actualizado exitosamente")
}

// ============================================
// DELETE - Eliminar módulo (AJAX)
// ============================================
func (h *ModuloHandler) Delete(c echo.Context) error {
	log.Println("[DEBUG] ModuloHandler.Delete - Iniciando")

	userID, ok := c.Get("user_id").(int)
	if !ok {
		log.Printf("[ERROR] Delete - No autenticado")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No autenticado"})
	}

	userName, ok := c.Get("user_name").(string)
	if !ok {
		userName = "Desconocido"
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("[ERROR] Delete - ID inválido: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	log.Printf("[DEBUG] Delete - Eliminando módulo ID: %d por usuario %d (%s)", id, userID, userName)

	if err := h.service.Delete(id); err != nil {
		log.Printf("[ERROR] Delete - Error en service.Delete: %v", err)
		log.Printf("[WARN] Usuario %d (%s) intentó eliminar módulo ID=%d y falló: %v",
			userID, userName, id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	log.Printf("[AUDIT] ✅ Usuario %d (%s) eliminó módulo ID=%d", userID, userName, id)

	return c.JSON(http.StatusOK, map[string]string{"message": "Módulo eliminado exitosamente"})
}

// ============================================
// DETALLE - Obtener módulo por ID (JSON para AJAX)
// ============================================
func (h *ModuloHandler) Detalle(c echo.Context) error {
	log.Println("[DEBUG] ModuloHandler.Detalle - Iniciando")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("[ERROR] Detalle - ID inválido: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ID inválido",
		})
	}

	log.Printf("[DEBUG] Detalle - Buscando módulo ID: %d", id)

	modulo, err := h.service.GetByID(id)
	if err != nil {
		log.Printf("[ERROR] Detalle - Módulo no encontrado: %v", err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Módulo no encontrado",
		})
	}

	log.Printf("[DEBUG] Detalle - Módulo encontrado: ID=%d, Nombre=%s", modulo.ID, modulo.Nombre)

	// Formatear fecha para el frontend
	fechaFormateada := modulo.CreatedAt.Format("02/01/2006 15:04:05")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":          modulo.ID,
		"nombre":      modulo.Nombre,
		"descripcion": modulo.Descripcion,
		"activo":      modulo.Activo,
		"created_at":  fechaFormateada,
	})
}
