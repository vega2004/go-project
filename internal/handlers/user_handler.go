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

// ============================================
// ESTRUCTURA Y CONSTRUCTOR
// ============================================

type UserHandler struct {
	userService    service.UserService
	perfilService  service.PerfilService
	env            *config.Env
	csrfMiddleware *middleware.CSRFMiddleware
}

func NewUserHandler(us service.UserService, ps service.PerfilService, env *config.Env) *UserHandler {
	return &UserHandler{
		userService:    us,
		perfilService:  ps,
		env:            env,
		csrfMiddleware: middleware.NewCSRFMiddleware(env.IsProduction()),
	}
}

// ============================================
// INDEX - Listar usuarios
// ============================================
func (h *UserHandler) Index(c echo.Context) error {
	log.Println("[DEBUG] Index - Iniciando")
	h.csrfMiddleware.SetToken(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	perfilID, _ := strconv.Atoi(c.QueryParam("perfil_id"))

	filter := &models.UserFilter{
		Name:     c.QueryParam("name"),
		Email:    c.QueryParam("email"),
		PerfilID: perfilID,
		Activo:   c.QueryParam("activo"),
		Page:     page,
		PageSize: 10,
	}

	log.Printf("[DEBUG] Index - Filtros: Name=%s, Email=%s, PerfilID=%d, Page=%d",
		filter.Name, filter.Email, filter.PerfilID, filter.Page)

	result, err := h.userService.GetAll(filter)
	if err != nil {
		log.Printf("[ERROR] UserHandler.Index: %v", err)
		return c.Render(http.StatusOK, "error.html", map[string]interface{}{
			"Title":   "Error",
			"Code":    500,
			"Message": "Error al cargar usuarios",
		})
	}

	log.Printf("[DEBUG] Index - Usuarios encontrados: %d", result.Total)

	perfiles, err := h.perfilService.GetAll(&models.PerfilFilter{Page: 1, PageSize: 100})
	if err != nil {
		log.Printf("[ERROR] UserHandler.Index - Error cargando perfiles: %v", err)
		perfiles = &models.PerfilPaginatedResponse{Data: []models.Perfil{}}
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	permisos := c.Get("permisos")
	if permisos == nil {
		permisos = make(map[string]models.Permiso)
	}

	return c.Render(http.StatusOK, "seguridad/usuarios.html", map[string]interface{}{
		"Title":      "Usuarios",
		"Usuarios":   result,
		"Perfiles":   perfiles.Data,
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
func (h *UserHandler) CreateForm(c echo.Context) error {
	log.Println("[DEBUG] CreateForm - Iniciando")
	h.csrfMiddleware.SetToken(c)

	perfiles, err := h.perfilService.GetAll(&models.PerfilFilter{Page: 1, PageSize: 100})
	if err != nil {
		log.Printf("[ERROR] UserHandler.CreateForm: %v", err)
		perfiles = &models.PerfilPaginatedResponse{Data: []models.Perfil{}}
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	permisos := c.Get("permisos")
	if permisos == nil {
		permisos = make(map[string]models.Permiso)
	}

	return c.Render(http.StatusOK, "seguridad/usuario_form.html", map[string]interface{}{
		"Title":     "Nuevo Usuario",
		"Usuario":   nil,
		"Perfiles":  perfiles.Data,
		"CSRFToken": csrfToken,
		"Permisos":  permisos,
	})
}

// ============================================
// CREATE - Crear nuevo usuario
// ============================================
// ============================================
// CREATE - Crear nuevo usuario
// ============================================
func (h *UserHandler) Create(c echo.Context) error {
	log.Println("[DEBUG] Create - Iniciando")

	auditorID, ok := c.Get("user_id").(int)
	if !ok {
		auditorID = 0
	}
	auditorName := c.Get("user_name").(string)
	if auditorName == "" {
		auditorName = "Desconocido"
	}

	perfilID, _ := strconv.Atoi(c.FormValue("perfil_id"))
	activo := c.FormValue("activo") == "on"

	req := &models.UserCreateRequest{
		Name:            strings.TrimSpace(c.FormValue("name")),
		Email:           strings.TrimSpace(c.FormValue("email")),
		Phone:           strings.TrimSpace(c.FormValue("phone")),
		Password:        c.FormValue("password"),
		ConfirmPassword: c.FormValue("confirm_password"),
		PerfilID:        perfilID,
		Activo:          activo,
	}

	log.Printf("[DEBUG] Create - Datos: Name=%s, Email=%s, PerfilID=%d, Activo=%v",
		req.Name, req.Email, req.PerfilID, req.Activo)

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
	if req.Password != req.ConfirmPassword {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=Las contraseñas no coinciden")
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

	// ✅ OBTENER EL USUARIO RECIÉN CREADO PARA SABER SU ID
	user, err := h.userService.GetByEmail(req.Email)
	if err != nil {
		log.Printf("[WARN] No se pudo obtener el usuario recién creado: %v", err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?success=Usuario creado, pero error al guardar foto")
	}

	log.Printf("[DEBUG] Usuario creado con ID: %d", user.ID)

	// ✅ PROCESAR LA FOTO DE PERFIL
	file, err := c.FormFile("foto")
	if err == nil {
		src, err := file.Open()
		if err == nil {
			defer src.Close()

			// Usar el servicio de perfil para subir la foto
			ruta, err := h.perfilService.UpdateFoto(user.ID, src, file)
			if err != nil {
				log.Printf("[WARN] Error al subir foto para usuario %d: %v", user.ID, err)
			} else {
				log.Printf("[INFO] Foto subida correctamente para usuario %d: %s", user.ID, ruta)
			}
		}
	} else {
		log.Printf("[DEBUG] No se recibió foto para el usuario %d", user.ID)
	}

	log.Printf("[AUDIT] Usuario %d (%s) creó usuario '%s' con perfil %d",
		auditorID, auditorName, req.Email, req.PerfilID)

	return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?success=Usuario creado exitosamente")
}

// ============================================
// EDIT FORM - Mostrar formulario de edición
// ============================================
func (h *UserHandler) EditForm(c echo.Context) error {
	log.Println("[DEBUG] EditForm - Iniciando")
	h.csrfMiddleware.SetToken(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("[ERROR] EditForm - ID inválido: %v", err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=ID inválido")
	}

	log.Printf("[DEBUG] EditForm - Buscando usuario ID: %d", id)

	user, err := h.userService.GetByID(id)
	if err != nil {
		log.Printf("[ERROR] EditForm - Usuario no encontrado: %v", err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=Usuario no encontrado")
	}

	log.Printf("[DEBUG] EditForm - Usuario encontrado: ID=%d, Name=%s, Email=%s, PerfilID=%d",
		user.ID, user.Name, user.Email, user.PerfilID)

	perfiles, err := h.perfilService.GetAll(&models.PerfilFilter{Page: 1, PageSize: 100})
	if err != nil {
		log.Printf("[ERROR] UserHandler.EditForm: %v", err)
		perfiles = &models.PerfilPaginatedResponse{Data: []models.Perfil{}}
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	permisos := c.Get("permisos")
	if permisos == nil {
		permisos = make(map[string]models.Permiso)
	}

	return c.Render(http.StatusOK, "seguridad/usuario_form.html", map[string]interface{}{
		"Title":     "Editar Usuario",
		"Usuario":   user,
		"Perfiles":  perfiles.Data,
		"CSRFToken": csrfToken,
		"Permisos":  permisos,
	})
}

// ============================================
// UPDATE - Actualizar usuario
// ============================================
// ============================================
// UPDATE - Actualizar usuario
// ============================================
func (h *UserHandler) Update(c echo.Context) error {
	log.Println("[DEBUG] Update - Iniciando")

	auditorID, ok := c.Get("user_id").(int)
	if !ok {
		auditorID = 0
	}
	auditorName := c.Get("user_name").(string)
	if auditorName == "" {
		auditorName = "Desconocido"
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("[ERROR] Update - ID inválido: %v", err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=ID inválido")
	}

	log.Printf("[DEBUG] Update - ID recibido: %d, Auditor: %d (%s)", id, auditorID, auditorName)

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

	log.Printf("[DEBUG] Update - Datos recibidos: Name=%s, Email=%s, PerfilID=%d, Activo=%v, Password=%s",
		req.Name, req.Email, req.PerfilID, req.Activo,
		func() string {
			if req.Password != "" {
				return "***"
			} else {
				return "vacío"
			}
		}())

	if req.Name == "" {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=El nombre es requerido")
	}
	if req.Email == "" {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=El email es requerido")
	}
	if req.PerfilID <= 0 {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=Debe seleccionar un perfil")
	}

	if req.Password != "" && len(req.Password) < 6 {
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error=La contraseña debe tener al menos 6 caracteres")
	}

	// Actualizar usuario
	if err := h.userService.Update(req, auditorID); err != nil {
		log.Printf("[WARN] Usuario %d (%s) intentó actualizar usuario ID=%d y falló: %v",
			auditorID, auditorName, id, err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?error="+err.Error())
	}

	// ✅ ACTUALIZAR LA FOTO DE PERFIL
	file, err := c.FormFile("foto")
	if err == nil {
		src, err := file.Open()
		if err == nil {
			defer src.Close()

			// Usar el servicio de perfil para subir la nueva foto
			ruta, err := h.perfilService.UpdateFoto(id, src, file)
			if err != nil {
				log.Printf("[WARN] Error al actualizar foto para usuario %d: %v", id, err)
			} else {
				log.Printf("[INFO] Foto actualizada correctamente para usuario %d: %s", id, ruta)
			}
		}
	} else {
		log.Printf("[DEBUG] No se recibió nueva foto para el usuario %d", id)
	}

	log.Printf("[AUDIT] Usuario %d (%s) actualizó usuario ID=%d",
		auditorID, auditorName, id)

	return c.Redirect(http.StatusSeeOther, "/seguridad/usuarios?success=Usuario actualizado exitosamente")
}

// ============================================
// DELETE - Eliminar usuario (AJAX)
// ============================================
func (h *UserHandler) Delete(c echo.Context) error {
	log.Println("[DEBUG] Delete - Iniciando")

	auditorID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No autenticado"})
	}
	auditorName := c.Get("user_name").(string)
	if auditorName == "" {
		auditorName = "Desconocido"
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	log.Printf("[DEBUG] Delete - Eliminando usuario ID: %d por auditor %d (%s)", id, auditorID, auditorName)

	if id == auditorID {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No puedes eliminar tu propio usuario"})
	}

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
	log.Println("[DEBUG] Detail - Iniciando")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("[ERROR] Detail - ID inválido: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	log.Printf("[DEBUG] Detail - Buscando usuario ID: %d", id)

	user, err := h.userService.GetByID(id)
	if err != nil {
		log.Printf("[ERROR] Detail - Usuario no encontrado: %v", err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Usuario no encontrado"})
	}

	log.Printf("[DEBUG] Detail - Usuario encontrado: ID=%d, Name=%s, Email=%s",
		user.ID, user.Name, user.Email)

	perfiles, _ := h.perfilService.GetAll(&models.PerfilFilter{Page: 1, PageSize: 100})
	perfilNombre := ""
	for _, p := range perfiles.Data {
		if p.ID == user.PerfilID {
			perfilNombre = p.Nombre
			break
		}
	}

	log.Printf("[DEBUG] Detail - Perfil encontrado: %s", perfilNombre)

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
	log.Println("[DEBUG] ToggleStatus - Iniciando")

	auditorID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No autenticado"})
	}
	auditorName := c.Get("user_name").(string)
	if auditorName == "" {
		auditorName = "Desconocido"
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

	log.Printf("[DEBUG] ToggleStatus - Usuario ID: %d, Nuevo estado: %v, Auditor: %d (%s)",
		id, req.Activo, auditorID, auditorName)

	if id == auditorID && !req.Activo {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No puedes desactivar tu propio usuario"})
	}

	if err := h.userService.UpdateStatus(id, req.Activo, auditorID); err != nil {
		log.Printf("[WARN] Usuario %d (%s) intentó cambiar estado de usuario %d y falló: %v",
			auditorID, auditorName, id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	status := "activado"
	if !req.Activo {
		status = "desactivado"
	}
	log.Printf("[AUDIT] Usuario %d (%s) %s usuario %d", auditorID, auditorName, status, id)

	return c.JSON(http.StatusOK, map[string]string{"message": "Estado actualizado exitosamente"})
}
