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

func (h *ImagenHandler) ShowCarrusel(c echo.Context) error {
	userID := c.Get("user_id").(int)

	imagenes, err := h.imagenService.GetAll()
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/maintenance")
	}

	data := map[string]interface{}{
		"Title":       "Gestión de Carrusel",
		"UserID":      userID,
		"Imagenes":    imagenes,
		"breadcrumbs": c.Get("breadcrumbs"),
	}

	return c.Render(http.StatusOK, "carrusel.html", data)
}

func (h *ImagenHandler) Upload(c echo.Context) error {
	userID := c.Get("user_id").(int)

	file, err := c.FormFile("imagen")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No se recibió ninguna imagen"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al abrir la imagen"})
	}
	defer src.Close()

	uploadedFile, err := utils.SaveUploadedFile(src, file)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	imagen := &model.Imagen{
		Nombre: uploadedFile.OriginalName,
		Ruta:   uploadedFile.Path,
		Activo: true,
	}

	if err := h.imagenService.Save(imagen, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Imagen subida exitosamente",
		"imagen":  imagen,
	})
}

func (h *ImagenHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	if err := h.imagenService.Delete(id); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Imagen eliminada exitosamente"})
}

func (h *ImagenHandler) Reorder(c echo.Context) error {
	var ordenes []int
	if err := c.Bind(&ordenes); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Datos inválidos"})
	}

	if err := h.imagenService.Reorder(ordenes); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Orden actualizado exitosamente"})
}

func (h *ImagenHandler) GetCarruselJSON(c echo.Context) error {
	imagenes, err := h.imagenService.GetForCarrusel()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, imagenes)
}
