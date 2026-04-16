package handlers

import (
	"log"
	"net/http"
	"strconv"
	"tu-proyecto/internal/config"
	"tu-proyecto/internal/middleware"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/service"

	"github.com/labstack/echo/v4"
)

type PerfilHandler struct {
	service        service.PerfilService
	env            *config.Env
	csrfMiddleware *middleware.CSRFMiddleware
}

func NewPerfilHandler(s service.PerfilService, env *config.Env) *PerfilHandler {
	return &PerfilHandler{
		service:        s,
		env:            env,
		csrfMiddleware: middleware.NewCSRFMiddleware(env.IsProduction()),
	}
}

// ============================================
// CRUD PARA PERFILES/ROLES (tabla perfiles)
// ============================================

func (h *PerfilHandler) Index(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.Index - Iniciando")
	h.csrfMiddleware.SetToken(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	filter := &models.PerfilFilter{
		Nombre:   c.QueryParam("nombre"),
		Page:     page,
		PageSize: 10,
	}

	log.Printf("[DEBUG] PerfilHandler.Index - Filtros: Nombre=%s, Page=%d", filter.Nombre, filter.Page)

	result, err := h.service.GetAll(filter)
	if err != nil {
		log.Printf("[ERROR] PerfilHandler.Index: %v", err)
		return c.Render(http.StatusOK, "error.html", map[string]interface{}{
			"Title":   "Error",
			"Code":    500,
			"Message": "Error al cargar perfiles",
		})
	}

	log.Printf("[DEBUG] PerfilHandler.Index - Perfiles encontrados: %d", result.Total)

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	permisos := c.Get("permisos")
	if permisos == nil {
		permisos = make(map[string]models.Permiso)
	}

	return c.Render(http.StatusOK, "seguridad/perfiles.html", map[string]interface{}{
		"Title":      "Perfiles",
		"Perfiles":   result,
		"Filtros":    filter,
		"SuccessMsg": c.QueryParam("success"),
		"ErrorMsg":   c.QueryParam("error"),
		"CSRFToken":  csrfToken,
		"Permisos":   permisos,
	})
}

func (h *PerfilHandler) CreateForm(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.CreateForm - Iniciando")
	h.csrfMiddleware.SetToken(c)

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "seguridad/perfil_form.html", map[string]interface{}{
		"Title":     "Nuevo Perfil",
		"Perfil":    &models.Perfil{},
		"CSRFToken": csrfToken,
	})
}

func (h *PerfilHandler) Create(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.Create - Iniciando")

	userID := c.Get("user_id").(int)
	userName := c.Get("user_name").(string)

	nombre := c.FormValue("nombre")
	descripcion := c.FormValue("descripcion")

	log.Printf("[DEBUG] PerfilHandler.Create - Nombre recibido: '%s'", nombre)
	log.Printf("[DEBUG] PerfilHandler.Create - Descripción recibida: '%s'", descripcion)
	log.Printf("[DEBUG] PerfilHandler.Create - Usuario: %d (%s)", userID, userName)

	perfil := &models.Perfil{
		Nombre:      nombre,
		Descripcion: descripcion,
	}

	log.Printf("[DEBUG] PerfilHandler.Create - Llamando a service.Create...")
	if err := h.service.Create(perfil); err != nil {
		log.Printf("[ERROR] PerfilHandler.Create - Error al crear perfil: %v", err)
		log.Printf("[WARN] Usuario %d (%s) intentó crear perfil '%s' y falló: %v",
			userID, userName, perfil.Nombre, err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d (%s) creó perfil ID=%d (%s)",
		userID, userName, perfil.ID, perfil.Nombre)

	return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?success=Perfil creado exitosamente")
}

func (h *PerfilHandler) EditForm(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.EditForm - Iniciando")
	h.csrfMiddleware.SetToken(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("[ERROR] EditForm - ID inválido: %v", err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?error=ID inválido")
	}

	log.Printf("[DEBUG] EditForm - Buscando perfil ID: %d", id)

	perfil, err := h.service.GetByID(id)
	if err != nil {
		log.Printf("[ERROR] EditForm - Perfil no encontrado: %v", err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?error=Perfil no encontrado")
	}

	log.Printf("[DEBUG] EditForm - Perfil encontrado: ID=%d, Nombre=%s", perfil.ID, perfil.Nombre)

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "seguridad/perfil_form.html", map[string]interface{}{
		"Title":     "Editar Perfil",
		"Perfil":    perfil,
		"CSRFToken": csrfToken,
	})
}

func (h *PerfilHandler) Update(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.Update - Iniciando")

	userID := c.Get("user_id").(int)
	userName := c.Get("user_name").(string)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("[ERROR] Update - ID inválido: %v", err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?error=ID inválido")
	}

	nombre := c.FormValue("nombre")
	descripcion := c.FormValue("descripcion")

	log.Printf("[DEBUG] Update - ID: %d, Nombre: '%s', Descripción: '%s'", id, nombre, descripcion)

	perfil := &models.Perfil{
		ID:          id,
		Nombre:      nombre,
		Descripcion: descripcion,
	}

	if err := h.service.Update(perfil); err != nil {
		log.Printf("[WARN] Usuario %d (%s) intentó actualizar perfil ID=%d y falló: %v",
			userID, userName, id, err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d (%s) actualizó perfil ID=%d (%s)",
		userID, userName, id, perfil.Nombre)

	return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?success=Perfil actualizado exitosamente")
}

func (h *PerfilHandler) Delete(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.Delete - Iniciando")

	userID := c.Get("user_id").(int)
	userName := c.Get("user_name").(string)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("[ERROR] Delete - ID inválido: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	log.Printf("[DEBUG] Delete - Eliminando perfil ID: %d por usuario %d (%s)", id, userID, userName)

	if id == 1 {
		log.Printf("[DEBUG] Delete - Intento de eliminar perfil de Administrador bloqueado")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No se puede eliminar el perfil de Administrador por defecto",
		})
	}

	if err := h.service.Delete(id); err != nil {
		log.Printf("[WARN] Usuario %d (%s) intentó eliminar perfil ID=%d y falló: %v",
			userID, userName, id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	log.Printf("[AUDIT] Usuario %d (%s) eliminó perfil ID=%d",
		userID, userName, id)

	return c.JSON(http.StatusOK, map[string]string{"message": "Perfil eliminado exitosamente"})
}

// ============================================
// MÉTODOS PARA PERFIL DE USUARIO AUTENTICADO
// ============================================

// ShowPerfil - Muestra el perfil del usuario autenticado
func (h *PerfilHandler) ShowPerfil(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.ShowPerfil - Iniciando")

	h.csrfMiddleware.SetToken(c)

	userID := c.Get("user_id").(int)
	userName := c.Get("user_name").(string)
	userEmail := c.Get("user_email").(string)
	userPerfil := c.Get("user_perfil").(string)

	log.Printf("[DEBUG] ShowPerfil - Usuario ID: %d, Nombre: %s", userID, userName)

	perfilUsuario, err := h.service.GetPerfil(userID)
	if err != nil {
		log.Printf("[ERROR] PerfilHandler.ShowPerfil: %v", err)
		return c.Redirect(http.StatusSeeOther, "/dashboard?error=Error al cargar perfil")
	}

	fotoPath := "/static/uploads/perfil/default-avatar.png"
	if perfilUsuario.FotoPath != "" {
		fotoPath = perfilUsuario.FotoPath
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "perfil.html", map[string]interface{}{
		"Title":      "Mi Perfil",
		"UserID":     userID,
		"UserName":   userName,
		"UserEmail":  userEmail,
		"UserPerfil": userPerfil,
		"Perfil":     perfilUsuario,
		"FotoPath":   fotoPath,
		"Success":    c.QueryParam("success"),
		"Error":      c.QueryParam("error"),
		"CSRFToken":  csrfToken,
		"breadcrumbs": []map[string]string{
			{"name": "Inicio", "url": "/"},
			{"name": "Dashboard", "url": "/dashboard"},
			{"name": "Mi Perfil", "url": ""},
		},
	})
}

// UpdatePerfil - Actualiza datos del perfil
// UpdatePerfil - Actualiza datos del perfil
func (h *PerfilHandler) UpdatePerfil(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.UpdatePerfil - Iniciando")

	userID := c.Get("user_id").(int)
	bio := c.FormValue("bio")
	direccion := c.FormValue("ubicacion")
	telefonoAlterno := c.FormValue("telefono_alterno") // ✅ AGREGAR ESTA LÍNEA

	log.Printf("[DEBUG] UpdatePerfil - UserID: %d, Bio: %s, Direccion: %s, Telefono: %s",
		userID, bio, direccion, telefonoAlterno)

	perfilUsuario := &models.PerfilUsuario{
		Bio:             bio,
		Direccion:       direccion,
		TelefonoAlterno: telefonoAlterno, // ✅ AGREGAR ESTA LÍNEA
	}

	if err := h.service.UpdatePerfil(userID, perfilUsuario); err != nil {
		log.Printf("[ERROR] PerfilHandler.UpdatePerfil: %v", err)
		return c.Redirect(http.StatusSeeOther, "/perfil?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d actualizó su perfil", userID)
	return c.Redirect(http.StatusSeeOther, "/perfil?success=Perfil actualizado correctamente")
}

// UploadFoto - Sube una nueva foto de perfil
func (h *PerfilHandler) UploadFoto(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.UploadFoto - Iniciando")

	userID := c.Get("user_id").(int)

	file, err := c.FormFile("foto")
	if err != nil {
		log.Printf("[DEBUG] UploadFoto - No se recibió foto: %v", err)
		return c.Redirect(http.StatusSeeOther, "/perfil?error=No se seleccionó ninguna imagen")
	}

	log.Printf("[DEBUG] UploadFoto - Archivo recibido: %s, Tamaño: %d bytes", file.Filename, file.Size)

	src, err := file.Open()
	if err != nil {
		log.Printf("[ERROR] UploadFoto - Error al abrir archivo: %v", err)
		return c.Redirect(http.StatusSeeOther, "/perfil?error=Error al abrir la imagen")
	}
	defer src.Close()

	ruta, err := h.service.UpdateFoto(userID, src, file)
	if err != nil {
		log.Printf("[ERROR] PerfilHandler.UploadFoto: %v", err)
		return c.Redirect(http.StatusSeeOther, "/perfil?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d actualizó su foto de perfil: %s", userID, ruta)
	return c.Redirect(http.StatusSeeOther, "/perfil?success=Foto actualizada correctamente")
}

// ChangePassword - Cambia la contraseña del usuario
func (h *PerfilHandler) ChangePassword(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.ChangePassword - Iniciando")

	userID := c.Get("user_id").(int)

	actual := c.FormValue("actual")
	nueva := c.FormValue("nueva")
	confirmar := c.FormValue("confirmar")

	log.Printf("[DEBUG] ChangePassword - UserID: %d", userID)

	if nueva != confirmar {
		log.Printf("[DEBUG] ChangePassword - Las contraseñas no coinciden")
		return c.Redirect(http.StatusSeeOther, "/perfil?error=Las contraseñas no coinciden")
	}

	form := &models.CambioPassword{
		Actual:    actual,
		Nueva:     nueva,
		Confirmar: confirmar,
	}

	if err := h.service.ChangePassword(userID, form); err != nil {
		log.Printf("[ERROR] PerfilHandler.ChangePassword: %v", err)
		return c.Redirect(http.StatusSeeOther, "/perfil?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d cambió su contraseña", userID)
	return c.Redirect(http.StatusSeeOther, "/perfil?success=Contraseña cambiada correctamente")
}

// DeleteFoto - Elimina la foto de perfil
func (h *PerfilHandler) DeleteFoto(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.DeleteFoto - Iniciando")

	userID := c.Get("user_id").(int)

	if err := h.service.DeleteFoto(userID); err != nil {
		log.Printf("[ERROR] PerfilHandler.DeleteFoto: %v", err)
		return c.Redirect(http.StatusSeeOther, "/perfil?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d eliminó su foto de perfil", userID)
	return c.Redirect(http.StatusSeeOther, "/perfil?success=Foto eliminada correctamente")
}

// ShowPerfilJSON - Devuelve el perfil del usuario en formato JSON
func (h *PerfilHandler) ShowPerfilJSON(c echo.Context) error {
	log.Println("[DEBUG] PerfilHandler.ShowPerfilJSON - Iniciando")

	userID, ok := c.Get("user_id").(int)
	if !ok {
		log.Printf("[ERROR] ShowPerfilJSON - No autenticado")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No autenticado"})
	}

	log.Printf("[DEBUG] ShowPerfilJSON - UserID: %d", userID)

	perfil, err := h.service.GetPerfil(userID)
	if err != nil {
		log.Printf("[ERROR] PerfilHandler.ShowPerfilJSON: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al cargar perfil"})
	}

	userName := c.Get("user_name")
	userEmail := c.Get("user_email")
	userPerfil := c.Get("user_perfil")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user_id":          userID,
			"name":             userName,
			"email":            userEmail,
			"perfil":           userPerfil,
			"bio":              perfil.Bio,
			"direccion":        perfil.Direccion,
			"telefono_alterno": perfil.TelefonoAlterno,
			"foto_path":        perfil.FotoPath,
			"updated_at":       perfil.UpdatedAt,
		},
	})
}
