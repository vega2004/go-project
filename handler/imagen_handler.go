package handler

import (
	"net/http"
	"strconv"
	"tu-proyecto/model"
	"tu-proyecto/service"
	"tu-proyecto/utils"

	"github.com/labstack/echo/v4"
)

type ImagenHandler struct {
	imagenService service.ImagenService
}

func NewImagenHandler(imagenService service.ImagenService) *ImagenHandler {
	return &ImagenHandler{
		imagenService: imagenService,
	}
}

// ShowCarrusel - Muestra la página del carrusel
func (h *ImagenHandler) ShowCarrusel(c echo.Context) error {
	userID := c.Get("user_id").(int)
	userName := c.Get("user_name").(string)
	userRole := c.Get("user_role").(string)

	imagenes, err := h.imagenService.GetAll()
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/maintenance")
	}

	data := map[string]interface{}{
		"Title":       "Gestión de Carrusel",
		"UserID":      userID,
		"UserName":    userName,
		"UserRole":    userRole,
		"Imagenes":    imagenes,
		"Success":     c.QueryParam("success"),
		"Error":       c.QueryParam("error"),
		"breadcrumbs": c.Get("breadcrumbs"),
	}

	return c.Render(http.StatusOK, "carrusel.html", data)
}

// Upload - Sube una nueva imagen (VERSIÓN CORREGIDA - USA REDIRECT)
func (h *ImagenHandler) Upload(c echo.Context) error {
	userID := c.Get("user_id").(int)
	userRole := c.Get("user_role").(string)

	// Verificar permisos
	if userRole != "admin" && userRole != "editor" {
		return c.Redirect(http.StatusSeeOther, "/carrusel?error=No tienes permisos para subir imágenes")
	}

	// Obtener el archivo
	file, err := c.FormFile("imagen")
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/carrusel?error=No se recibió ninguna imagen")
	}

	// Validar límite de imágenes
	imagenes, err := h.imagenService.GetAll()
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/carrusel?error=Error al verificar límite de imágenes")
	}

	if len(imagenes) >= 10 {
		return c.Redirect(http.StatusSeeOther, "/carrusel?error=Máximo 10 imágenes permitidas")
	}

	// Abrir el archivo
	src, err := file.Open()
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/carrusel?error=Error al abrir la imagen")
	}
	defer src.Close()

	// Validar y guardar la imagen
	uploadedFile, err := utils.SaveUploadedFile(src, file)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/carrusel?error="+err.Error())
	}

	// Crear registro en base de datos
	imagen := &model.Imagen{
		Nombre: uploadedFile.OriginalName,
		Ruta:   uploadedFile.Path,
		Activo: true,
		UserID: userID,
	}

	if err := h.imagenService.Save(imagen, userID); err != nil {
		// Si falla la BD, eliminar el archivo subido
		utils.DeleteFile(uploadedFile.Path)
		return c.Redirect(http.StatusSeeOther, "/carrusel?error=Error al guardar en base de datos")
	}

	// ✅ REDIRECCIÓN CON MENSAJE DE ÉXITO (NO JSON)
	return c.Redirect(http.StatusSeeOther, "/carrusel?success=Imagen subida exitosamente")
}

// Delete - Elimina una imagen (VERSIÓN CORREGIDA - USA REDIRECT)
func (h *ImagenHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/carrusel?error=ID inválido")
	}

	userRole := c.Get("user_role").(string)
	if userRole != "admin" && userRole != "editor" {
		return c.Redirect(http.StatusSeeOther, "/carrusel?error=No tienes permisos para eliminar imágenes")
	}

	// Obtener la imagen antes de eliminarla (para borrar el archivo)
	imagenes, err := h.imagenService.GetAll()
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/carrusel?error=Error al obtener la imagen")
	}

	var imagenEliminar *model.Imagen
	for _, img := range imagenes {
		if img.ID == id {
			imagenEliminar = &img
			break
		}
	}

	if imagenEliminar == nil {
		return c.Redirect(http.StatusSeeOther, "/carrusel?error=Imagen no encontrada")
	}

	// Eliminar de la base de datos
	if err := h.imagenService.Delete(id); err != nil {
		return c.Redirect(http.StatusSeeOther, "/carrusel?error="+err.Error())
	}

	// Eliminar el archivo físico
	utils.DeleteFile(imagenEliminar.Ruta)

	// ✅ REDIRECCIÓN CON MENSAJE DE ÉXITO
	return c.Redirect(http.StatusSeeOther, "/carrusel?success=Imagen eliminada exitosamente")
}

// GetCarruselJSON - Devuelve las imágenes en formato JSON para Fetch API
func (h *ImagenHandler) GetCarruselJSON(c echo.Context) error {
	imagenes, err := h.imagenService.GetForCarrusel()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Error al cargar imágenes",
		})
	}

	return c.JSON(http.StatusOK, imagenes)
}

// Reorder - Reordena las imágenes (drag & drop)
func (h *ImagenHandler) Reorder(c echo.Context) error {
	var ordenes []int
	if err := c.Bind(&ordenes); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Datos inválidos",
		})
	}

	userRole := c.Get("user_role").(string)
	if userRole != "admin" && userRole != "editor" {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "No tienes permisos para reordenar",
		})
	}

	if err := h.imagenService.Reorder(ordenes); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Orden actualizado exitosamente",
	})
}
