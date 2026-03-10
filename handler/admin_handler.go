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

	result, err := h.userService.GetAllUsers(filter)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/maintenance")
	}

	data := map[string]interface{}{
		"Title":       "Administración de Usuarios",
		"Users":       result,
		"Filtros":     filter,
		"Success":     c.QueryParam("success"), // ← Mensaje de éxito
		"Error":       c.QueryParam("error"),   // ← Mensaje de error
		"breadcrumbs": c.Get("breadcrumbs"),
	}

	return c.Render(http.StatusOK, "admin_users.html", data)
}

// CreateUserForm - Mostrar formulario para crear usuario
func (h *AdminHandler) CreateUserForm(c echo.Context) error {
	data := map[string]interface{}{
		"Title":       "Crear Usuario",
		"Error":       c.QueryParam("error"), // ← Error del formulario
		"breadcrumbs": c.Get("breadcrumbs"),
	}
	return c.Render(http.StatusOK, "admin_user_form.html", data)
}

// CreateUser - Procesar creación de usuario (CORREGIDO)
func (h *AdminHandler) CreateUser(c echo.Context) error {
	var form model.RegisterForm
	if err := c.Bind(&form); err != nil {
		return c.Redirect(http.StatusSeeOther, "/admin/users/create?error=Error en formulario")
	}

	// Asignar rol
	roleID, _ := strconv.Atoi(c.FormValue("role_id"))
	if roleID < 1 || roleID > 3 {
		roleID = 2
	}

	err := h.userService.CreateUserByAdmin(&form, roleID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/admin/users/create?error="+err.Error())
	}

	// ✅ REDIRECCIÓN A LISTA CON MENSAJE DE ÉXITO
	return c.Redirect(http.StatusSeeOther, "/admin/users?success=Usuario creado exitosamente")
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
		"Error":       c.QueryParam("error"),
		"breadcrumbs": c.Get("breadcrumbs"),
	}
	return c.Render(http.StatusOK, "admin_user_form.html", data)
}

// UpdateUser - Actualizar usuario (CORREGIDO)
func (h *AdminHandler) UpdateUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var form model.RegisterForm
	if err := c.Bind(&form); err != nil {
		return c.Redirect(http.StatusSeeOther, "/admin/users/edit/"+strconv.Itoa(id)+"?error=Error en formulario")
	}

	roleID, _ := strconv.Atoi(c.FormValue("role_id"))
	if roleID < 1 || roleID > 3 {
		roleID = 2
	}

	err := h.userService.UpdateUser(id, &form, roleID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/admin/users/edit/"+strconv.Itoa(id)+"?error="+err.Error())
	}

	// ✅ REDIRECCIÓN A LISTA CON MENSAJE DE ÉXITO
	return c.Redirect(http.StatusSeeOther, "/admin/users?success=Usuario actualizado correctamente")
}

// DeleteUser - Eliminar usuario (mejorado)
func (h *AdminHandler) DeleteUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	// Verificar confirmación
	if c.QueryParam("confirm") != "true" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Se requiere confirmación",
		})
	}

	err := h.userService.DeleteUser(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// ✅ RESPUESTA JSON (el fetch recargará la página)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Usuario eliminado exitosamente",
	})
}
