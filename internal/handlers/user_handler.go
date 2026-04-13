package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/service"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService   service.UserService
	perfilService service.PerfilService
}

func NewUserHandler(us service.UserService, ps service.PerfilService) *UserHandler {
	return &UserHandler{
		userService:   us,
		perfilService: ps,
	}
}

// ============================================
// INDEX - Listar usuarios
// ============================================
func (h *UserHandler) Index(c echo.Context) error {
	// Obtener parámetros de filtro
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	var activo *bool
	if c.QueryParam("activo") != "" {
		val := c.QueryParam("activo") == "true"
		activo = &val
	}

	perfilID, _ := strconv.Atoi(c.QueryParam("perfil_id"))

	filter := &models.UserFilter{
		Name:     c.QueryParam("name"),
		Email:    c.QueryParam("email"),
		PerfilID: perfilID,
		Activo:   activo,
		Page:     page,
		PageSize: 10,
	}

	// Obtener lista de usuarios
	result, err := h.userService.GetAll(filter)
	if err != nil {
		log.Printf("[ERROR] UserHandler.Index: %v", err)
		return c.Render(http.StatusOK, "error.html", map[string]interface{}{
			"Title":   "Error",
			"Code":    500,
			"Message": "Error al cargar usuarios",
		})
	}

	// Obtener lista de perfiles para el filtro
	perfiles, err := h.perfilService.GetAll(&models.PerfilFilter{Page: 1, PageSize: 100})
	if err != nil {
		log.Printf("[ERROR] UserHandler.Index - Error cargando perfiles: %v", err)
		perfiles = &models.PerfilPaginatedResponse{Data: []models.Perfil{}}
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "seguridad/usuarios.html", map[string]interface{}{
		"Title":      "Usuarios",
		"Usuarios":   result,
		"Perfiles":   perfiles.Data,
		"Filtros":    filter,
		"SuccessMsg": c.QueryParam("success"),
		"ErrorMsg":   c.QueryParam("error"),
		"CSRFToken":  csrfToken,
	})
}

// ============================================
// CREATE FORM - Mostrar formulario de creación
// ============================================
func (h *UserHandler) CreateForm(c echo.Context) error {
	// Obtener lista de perfiles
	perfiles, err := h.perfilService.GetAll(&models.PerfilFilter{Page: 1, PageSize: 100})
	if err != nil {
		log.Printf("[ERROR] UserHandler.CreateForm: %v", err)
		perfiles = &models.PerfilPaginatedResponse{Data: []models.Perfil{}}
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "seguridad/usuario_form.html", map[string]interface{}{
		"Title":     "Nuevo Usuario",
		"Usuario":   nil,
		"Perfiles":  perfiles.Data,
		"CSRFToken": csrfToken,
	})
}

// ============================================
// CREATE - Crear nuevo usuario
// ============================================
func (h *UserHandler) Create(c echo.Context) error {
	// Obtener usuario auditor
	auditorID, ok := c.Get("user_id").(int)
	if !ok {
		auditorID = 0
	}
	auditorName := c.Get("user_name").(string)

	// Parsear formulario
	perfilID, _ := strconv.Atoi(c.FormValue("perfil_id"))
	activo := c.FormValue("activo") == "on"

	req := &models.UserCreateRequest{
		Name:     strings.TrimSpace(c.FormValue("name")),
		Email:    strings.TrimSpace(c.FormValue("email")),
		Phone:    strings.TrimSpace(c.FormValue("phone")),
		Password: c.FormValue("password"),
		PerfilID: perfilID,
		Activo:   activo,
	}

	// Validaciones básicas
	if req.Name == "" {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=El nombre es requerido")
	}
	if req.Email == "" {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=El email es requerido")
	}
	if req.Password == "" {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=La contraseña es requerida")
	}
	if len(req.Password) < 6 {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=La contraseña debe tener al menos 6 caracteres")
	}
	if req.PerfilID <= 0 {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=Debe seleccionar un perfil")
	}

	// Crear usuario
	if err := h.userService.Create(req); err != nil {
		log.Printf("[WARN] Usuario %d (%s) intentó crear usuario '%s' y falló: %v",
			auditorID, auditorName, req.Email, err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d (%s) creó usuario '%s' con perfil %d",
		auditorID, auditorName, req.Email, req.PerfilID)

	return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?success=Usuario creado exitosamente")
}

// ============================================
// EDIT FORM - Mostrar formulario de edición
// ============================================
func (h *UserHandler) EditForm(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=ID inválido")
	}

	// Obtener usuario
	user, err := h.userService.GetByID(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=Usuario no encontrado")
	}

	// Obtener lista de perfiles
	perfiles, err := h.perfilService.GetAll(&models.PerfilFilter{Page: 1, PageSize: 100})
	if err != nil {
		log.Printf("[ERROR] UserHandler.EditForm: %v", err)
		perfiles = &models.PerfilPaginatedResponse{Data: []models.Perfil{}}
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "seguridad/usuario_form.html", map[string]interface{}{
		"Title":     "Editar Usuario",
		"Usuario":   user,
		"Perfiles":  perfiles.Data,
		"CSRFToken": csrfToken,
	})
}

// ============================================
// UPDATE - Actualizar usuario
// ============================================
// Update - Actualizar usuario
func (h *UserHandler) Update(c echo.Context) error {
	// Obtener usuario auditor
	auditorID, ok := c.Get("user_id").(int)
	if !ok {
		auditorID = 0
	}
	auditorName := c.Get("user_name").(string)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=ID inválido")
	}

	perfilID, _ := strconv.Atoi(c.FormValue("perfil_id"))
	activo := c.FormValue("activo") == "on"

	req := &models.UserUpdateRequest{
		ID:       id,
		Name:     strings.TrimSpace(c.FormValue("name")),
		Email:    strings.TrimSpace(c.FormValue("email")),
		Phone:    strings.TrimSpace(c.FormValue("phone")),
		Password: c.FormValue("password"),
		PerfilID: perfilID,
		Activo:   activo,
	}

	// Validaciones básicas
	if req.Name == "" {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=El nombre es requerido")
	}
	if req.Email == "" {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=El email es requerido")
	}
	if req.PerfilID <= 0 {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=Debe seleccionar un perfil")
	}

	// ✅ CORREGIDO: pasar auditorID
	if err := h.userService.Update(req, auditorID); err != nil {
		log.Printf("[WARN] Usuario %d (%s) intentó actualizar usuario ID=%d y falló: %v",
			auditorID, auditorName, id, err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d (%s) actualizó usuario ID=%d",
		auditorID, auditorName, id)

	return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?success=Usuario actualizado exitosamente")
}

// ============================================
// DELETE - Eliminar usuario (AJAX)
// ============================================
func (h *UserHandler) Delete(c echo.Context) error {
	// Obtener usuario auditor
	auditorID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No autenticado"})
	}
	auditorName := c.Get("user_name").(string)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	// No permitir eliminar el propio usuario
	if id == auditorID {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No puedes eliminar tu propio usuario"})
	}

	// Eliminar usuario
	if err := h.userService.Delete(id, auditorID); err != nil {
		log.Printf("[WARN] Usuario %d (%s) intentó eliminar usuario ID=%d y falló: %v",
			auditorID, auditorName, id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	log.Printf("[AUDIT] Usuario %d (%s) eliminó usuario ID=%d",
		auditorID, auditorName, id)

	return c.JSON(http.StatusOK, map[string]string{"message": "Usuario eliminado exitosamente"})
}

// ============================================
// DETAIL - Ver detalle de usuario (AJAX)
// ============================================
func (h *UserHandler) Detail(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Usuario no encontrado"})
	}

	// Obtener nombre del perfil
	perfiles, _ := h.perfilService.GetAll(&models.PerfilFilter{Page: 1, PageSize: 100})
	perfilNombre := ""
	for _, p := range perfiles.Data {
		if p.ID == user.PerfilID {
			perfilNombre = p.Nombre
			break
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":            user.ID,
		"name":          user.Name,
		"email":         user.Email,
		"phone":         user.Phone,
		"perfil_id":     user.PerfilID,
		"perfil_nombre": perfilNombre,
		"activo":        user.Activo,
		"created_at":    user.CreatedAt.Format("02/01/2006 15:04:05"),
		"updated_at":    user.UpdatedAt.Format("02/01/2006 15:04:05"),
	})
}

// ============================================
// TOGGLE STATUS - Cambiar estado activo/inactivo (AJAX)
// ============================================
func (h *UserHandler) ToggleStatus(c echo.Context) error {
	// Obtener usuario auditor
	auditorID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No autenticado"})
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	var req struct {
		Activo bool `json:"activo"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Datos inválidos"})
	}

	// No permitir desactivar el propio usuario
	if id == auditorID && !req.Activo {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No puedes desactivar tu propio usuario"})
	}

	if err := h.userService.UpdateStatus(id, req.Activo, auditorID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	status := "activado"
	if !req.Activo {
		status = "desactivado"
	}
	log.Printf("[AUDIT] Usuario %d cambió estado de usuario %d a %s", auditorID, id, status)

	return c.JSON(http.StatusOK, map[string]string{"message": "Estado actualizado exitosamente"})
}
