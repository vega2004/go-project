package handler

import (
	"net/http"
	"tu-proyecto/model"
	"tu-proyecto/service"
	"tu-proyecto/utils"

	"github.com/labstack/echo/v4"
)

type PerfilHandler struct {
	perfilService  service.PerfilService
	sessionManager *utils.SessionManager
}

func NewPerfilHandler(perfilService service.PerfilService, sm *utils.SessionManager) *PerfilHandler {
	return &PerfilHandler{
		perfilService:  perfilService,
		sessionManager: sm,
	}
}

// ShowPerfil - Muestra la página de perfil
func (h *PerfilHandler) ShowPerfil(c echo.Context) error {
	userID := c.Get("user_id").(int)
	userName := c.Get("user_name").(string)
	userEmail := c.Get("user_email").(string)
	userRole := c.Get("user_role").(string)

	perfil, err := h.perfilService.GetPerfil(userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/dashboard?error=Error al cargar perfil")
	}

	data := map[string]interface{}{
		"Title":       "Mi Perfil",
		"UserID":      userID,
		"UserName":    userName,
		"UserEmail":   userEmail,
		"UserRole":    userRole,
		"Perfil":      perfil,
		"Success":     c.QueryParam("success"),
		"Error":       c.QueryParam("error"),
		"breadcrumbs": c.Get("breadcrumbs"),
	}

	return c.Render(http.StatusOK, "perfil.html", data)
}

// UpdatePerfil - Actualiza datos del perfil
func (h *PerfilHandler) UpdatePerfil(c echo.Context) error {
	userID := c.Get("user_id").(int)

	perfil := &model.Perfil{
		Bio:       c.FormValue("bio"),
		Ubicacion: c.FormValue("ubicacion"),
		SitioWeb:  c.FormValue("sitio_web"),
	}

	err := h.perfilService.UpdatePerfil(userID, perfil)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/perfil?error="+err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/perfil?success=Datos actualizados correctamente")
}

// UploadFoto - Sube una nueva foto de perfil (VERSIÓN FINAL CORREGIDA)
// UploadFoto - Sube una nueva foto de perfil (VERSIÓN CORREGIDA)
func (h *PerfilHandler) UploadFoto(c echo.Context) error {
	userID := c.Get("user_id").(int)

	// Obtener el archivo del formulario
	file, err := c.FormFile("foto")
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/perfil?error=No se seleccionó ninguna imagen")
	}

	// Abrir el archivo
	src, err := file.Open()
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/perfil?error=Error al abrir la imagen")
	}
	defer src.Close()

	// Llamar al servicio con los parámetros correctos
	_, err = h.perfilService.UpdateFoto(userID, src, file) // ← IGNORAMOS ruta
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/perfil?error="+err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/perfil?success=Foto actualizada correctamente")
}

// ChangePassword - Cambia la contraseña
func (h *PerfilHandler) ChangePassword(c echo.Context) error {
	userID := c.Get("user_id").(int)

	// Obtener valores del formulario
	actual := c.FormValue("actual")
	nueva := c.FormValue("nueva")
	confirmar := c.FormValue("confirmar")

	// Validar que las contraseñas coincidan
	if nueva != confirmar {
		return c.Redirect(http.StatusSeeOther, "/perfil?error=Las contraseñas no coinciden")
	}

	form := &model.CambioPassword{
		Actual:    actual,
		Nueva:     nueva,
		Confirmar: confirmar,
	}

	err := h.perfilService.ChangePassword(userID, form)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/perfil?error="+err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/perfil?success=Contraseña cambiada correctamente")
}

// DeleteFoto - Elimina la foto de perfil (opcional)
func (h *PerfilHandler) DeleteFoto(c echo.Context) error {
	userID := c.Get("user_id").(int)

	err := h.perfilService.DeleteFoto(userID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/perfil?error="+err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/perfil?success=Foto eliminada correctamente")
}

// GetPerfilJSON - Obtiene datos del perfil en formato JSON (para AJAX)
func (h *PerfilHandler) GetPerfilJSON(c echo.Context) error {
	userID := c.Get("user_id").(int)

	perfil, err := h.perfilService.GetPerfil(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Error al cargar perfil",
		})
	}

	return c.JSON(http.StatusOK, perfil)
}
