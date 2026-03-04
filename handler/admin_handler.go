package handler

import (
	"net/http"
	"strconv"
	"tu-proyecto/model"
	"tu-proyecto/service"

	"github.com/labstack/echo/v4"
)

type AdminHandler struct {
	userService service.UserService
}

func NewAdminHandler(userService service.UserService) *AdminHandler {
	return &AdminHandler{
		userService: userService,
	}
}

// ShowUsers - Lista todos los usuarios (solo admin)
func (h *AdminHandler) ShowUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	filter := &model.UserFilter{
		Email:    c.QueryParam("email"),
		Name:     c.QueryParam("name"),
		Page:     page,
		PageSize: 5,
	}

	// Necesitas implementar GetAllUsers en UserService
	result, err := h.userService.GetAllUsers(filter)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/maintenance")
	}

	data := map[string]interface{}{
		"Title":       "Administración de Usuarios",
		"Users":       result,
		"breadcrumbs": c.Get("breadcrumbs"),
	}

	return c.Render(http.StatusOK, "admin_users.html", data)
}

// CreateUserForm - Mostrar formulario para crear usuario
func (h *AdminHandler) CreateUserForm(c echo.Context) error {
	data := map[string]interface{}{
		"Title":       "Crear Usuario",
		"breadcrumbs": c.Get("breadcrumbs"),
	}
	return c.Render(http.StatusOK, "admin_user_form.html", data)
}

// CreateUser - Procesar creación de usuario
func (h *AdminHandler) CreateUser(c echo.Context) error {
	var form model.RegisterForm
	if err := c.Bind(&form); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Datos inválidos"})
	}

	// Asignar rol (viene del formulario)
	roleID, _ := strconv.Atoi(c.FormValue("role_id"))

	err := h.userService.CreateUserByAdmin(&form, roleID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Usuario creado exitosamente"})
}

// EditUserForm - Mostrar formulario de edición
func (h *AdminHandler) EditUserForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/admin/users?error=Usuario no encontrado")
	}

	data := map[string]interface{}{
		"Title":       "Editar Usuario",
		"User":        user,
		"breadcrumbs": c.Get("breadcrumbs"),
	}
	return c.Render(http.StatusOK, "admin_user_form.html", data)
}

// UpdateUser - Actualizar usuario
func (h *AdminHandler) UpdateUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var form model.RegisterForm
	if err := c.Bind(&form); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Datos inválidos"})
	}

	roleID, _ := strconv.Atoi(c.FormValue("role_id"))

	err := h.userService.UpdateUser(id, &form, roleID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Usuario actualizado exitosamente"})
}

// DeleteUser - Eliminar usuario (con confirmación)
func (h *AdminHandler) DeleteUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	// Confirmación vía query param
	if c.QueryParam("confirm") != "true" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"require_confirm": true,
			"message":         "¿Está seguro de eliminar este usuario?",
		})
	}

	err := h.userService.DeleteUser(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Usuario eliminado exitosamente"})
}
