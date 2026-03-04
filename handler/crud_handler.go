package handler

import (
	"net/http"
	"strconv"
	"tu-proyecto/model"
	"tu-proyecto/service"

	"github.com/labstack/echo/v4"
)

type CrudHandler struct {
	crudService service.CrudService
}

func NewCrudHandler(crudService service.CrudService) *CrudHandler {
	return &CrudHandler{
		crudService: crudService,
	}
}

func (h *CrudHandler) ShowCrud(c echo.Context) error {
	userID := c.Get("user_id").(int)

	// Obtener filtros de la query
	filter := &model.PersonaFilter{
		Nombre:      c.QueryParam("nombre"),
		EstadoCivil: c.QueryParam("estadoCivil"),
		Page:        1,
		PageSize:    5,
	}

	// Si hay parámetro de página
	if page, err := strconv.Atoi(c.QueryParam("page")); err == nil && page > 0 {
		filter.Page = page
	}

	result, err := h.crudService.FindAll(filter)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/maintenance")
	}

	data := map[string]interface{}{
		"Title":       "CRUD de Personas",
		"UserID":      userID,
		"Data":        result,
		"Filtros":     filter,
		"breadcrumbs": c.Get("breadcrumbs"),
	}

	return c.Render(http.StatusOK, "crud.html", data)
}

func (h *CrudHandler) Create(c echo.Context) error {
	var persona model.Persona
	if err := c.Bind(&persona); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Datos inválidos"})
	}

	userID := c.Get("user_id").(int)

	if err := h.crudService.Create(&persona, userID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Registro creado exitosamente"})
}

func (h *CrudHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	var persona model.Persona
	if err := c.Bind(&persona); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Datos inválidos"})
	}

	persona.ID = id

	if err := h.crudService.Update(&persona); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Registro actualizado exitosamente"})
}

func (h *CrudHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID inválido"})
	}

	if err := h.crudService.Delete(id); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Registro eliminado exitosamente"})
}

func (h *CrudHandler) List(c echo.Context) error {
	filter := &model.PersonaFilter{
		Nombre:      c.QueryParam("nombre"),
		EstadoCivil: c.QueryParam("estadoCivil"),
		Page:        1,
		PageSize:    5,
	}

	if page, err := strconv.Atoi(c.QueryParam("page")); err == nil && page > 0 {
		filter.Page = page
	}

	result, err := h.crudService.FindAll(filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *CrudHandler) Filter(c echo.Context) error {
	var filter model.PersonaFilter
	if err := c.Bind(&filter); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Filtros inválidos"})
	}

	if filter.PageSize == 0 {
		filter.PageSize = 5
	}
	if filter.Page < 1 {
		filter.Page = 1
	}

	result, err := h.crudService.FindAll(&filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}
