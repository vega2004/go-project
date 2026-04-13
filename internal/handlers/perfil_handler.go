package handlers

import (
	"log"
	"net/http"
	"strconv"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/service"

	"github.com/labstack/echo/v4"
)

type PerfilHandler struct {
	service service.PerfilService
}

func NewPerfilHandler(s service.PerfilService) *PerfilHandler {
	return &PerfilHandler{service: s}
}

// ============================================
// CRUD PARA PERFILES/ROLES (tabla perfiles)
// ============================================

func (h *PerfilHandler) Index(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	filter := &models.PerfilFilter{
		Nombre:   c.QueryParam("nombre"),
		Page:     page,
		PageSize: 10,
	}

	result, err := h.service.GetAll(filter)
	if err != nil {
		log.Printf("[ERROR] PerfilHandler.Index: %v", err)
		return c.Render(http.StatusOK, "error.html", map[string]interface{}{
			"Title":   "Error",
			"Code":    500,
			"Message": "Error al cargar perfiles",
		})
	}

	return c.Render(http.StatusOK, "seguridad/perfiles.html", map[string]interface{}{
		"Title":      "Perfiles",
		"Perfiles":   result,
		"Filtros":    filter,
		"SuccessMsg": c.QueryParam("success"),
		"ErrorMsg":   c.QueryParam("error"),
		"CSRFToken":  c.Get("csrf_token"),
	})
}

func (h *PerfilHandler) CreateForm(c echo.Context) error {
	return c.Render(http.StatusOK, "seguridad/perfil_form.html", map[string]interface{}{
		"Title":     "Nuevo Perfil",
		"Perfil":    &models.Perfil{},
		"CSRFToken": c.Get("csrf_token"),
	})
}

func (h *PerfilHandler) Create(c echo.Context) error {
	userID := c.Get("user_id").(int)
	userName := c.Get("user_name").(string)

	perfil := &models.Perfil{
		Nombre:      c.FormValue("nombre"),
		Descripcion: c.FormValue("descripcion"),
	}

	if err := h.service.Create(perfil); err != nil {
		log.Printf("[WARN] Usuario %d (%s) intentó crear perfil '%s' y falló: %v",
			userID, userName, perfil.Nombre, err)
		return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d (%s) creó perfil ID=%d (%s)",
		userID, userName, perfil.ID, perfil.Nombre)

	return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?success=Perfil creado exitosamente")
}

func (h *PerfilHandler) EditForm(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?error=ID inválido")
	}

	perfil, err := h.service.GetByID(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?error=Perfil no encontrado")
	}

	return c.Render(http.StatusOK, "seguridad/perfil_form.html", map[string]interface{}{
		"Title":     "Editar Perfil",
		"Perfil":    perfil,
		"CSRFToken": c.Get("csrf_token"),
	})
}

func (h *PerfilHandler) Update(c echo.Context) error {
	userID := c.Get("user_id").(int)
	userName := c.Get("user_name").(string)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/seguridad/perfiles?error=ID inválido")
	}

	perfil := &models.Perfil{
		ID:          id,
		Nombre:      c.FormValue("nombre"),
		Descripcion: c.FormValue("descripcion"),
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
	userID := c.Get("user_id").(int)
	userName := c.Get("user_name").(string)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	if id == 1 {
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

// ShowPerfil - Muestra el perfil del usuario autenticado (CORREGIDO)
func (h *PerfilHandler) ShowPerfil(c echo.Context) error {
	userID := c.Get("user_id").(int)
	userName := c.Get("user_name").(string)
	userEmail := c.Get("user_email").(string)
	userRole := c.Get("user_role").(string)

	// ✅ Usar PerfilUsuario, no Perfil
	perfilUsuario, err := h.service.GetPerfil(userID)
	if err != nil {
		log.Printf("[ERROR] PerfilHandler.ShowPerfil: %v", err)
		return c.Redirect(http.StatusSeeOther, "/dashboard?error=Error al cargar perfil")
	}

	// Obtener foto del perfil
	fotoPath := "/static/uploads/perfil/default-avatar.png"
	if perfilUsuario.FotoPath != "" {
		fotoPath = perfilUsuario.FotoPath
	}

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "perfil.html", map[string]interface{}{
		"Title":     "Mi Perfil",
		"UserID":    userID,
		"UserName":  userName,
		"UserEmail": userEmail,
		"UserRole":  userRole,
		"Perfil":    perfilUsuario,
		"FotoPath":  fotoPath,
		"Success":   c.QueryParam("success"),
		"Error":     c.QueryParam("error"),
		"CSRFToken": csrfToken,
		"breadcrumbs": []map[string]string{
			{"name": "Inicio", "url": "/"},
			{"name": "Dashboard", "url": "/dashboard"},
			{"name": "Mi Perfil", "url": ""},
		},
	})
}

// UpdatePerfil - Actualiza datos del perfil (CORREGIDO)
func (h *PerfilHandler) UpdatePerfil(c echo.Context) error {
	userID := c.Get("user_id").(int)

	// ✅ Usar PerfilUsuario
	perfilUsuario := &models.PerfilUsuario{
		Bio:       c.FormValue("bio"),
		Direccion: c.FormValue("ubicacion"),
		// Nota: SitioWeb no está en PerfilUsuario, podrías agregarlo o ignorarlo
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
	userID := c.Get("user_id").(int)

	file, err := c.FormFile("foto")
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/perfil?error=No se seleccionó ninguna imagen")
	}

	src, err := file.Open()
	if err != nil {
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
	userID := c.Get("user_id").(int)

	actual := c.FormValue("actual")
	nueva := c.FormValue("nueva")
	confirmar := c.FormValue("confirmar")

	if nueva != confirmar {
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
	userID := c.Get("user_id").(int)

	if err := h.service.DeleteFoto(userID); err != nil {
		log.Printf("[ERROR] PerfilHandler.DeleteFoto: %v", err)
		return c.Redirect(http.StatusSeeOther, "/perfil?error="+err.Error())
	}

	log.Printf("[AUDIT] Usuario %d eliminó su foto de perfil", userID)
	return c.Redirect(http.StatusSeeOther, "/perfil?success=Foto eliminada correctamente")
}
