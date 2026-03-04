package handler

import (
	"net/http"
	"tu-proyecto/model"
	"tu-proyecto/service"
	"tu-proyecto/utils"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService    service.AuthService
	sessionManager *utils.SessionManager
}

func NewAuthHandler(authService service.AuthService, sm *utils.SessionManager) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		sessionManager: sm,
	}
}

func (h *AuthHandler) ShowLogin(c echo.Context) error {
	errorMsg := c.QueryParam("error")

	data := map[string]interface{}{
		"Title":        "Iniciar Sesión",
		"RecaptchaKey": "6LdjX1gsAAAAAAmGXFf5nW7EkXqx9xukAwyJR35Z",
		"Error":        errorMsg,
		"breadcrumbs":  c.Get("breadcrumbs"),
	}

	return c.Render(http.StatusOK, "login.html", data)
}

func (h *AuthHandler) DoLogin(c echo.Context) error {
	var form model.LoginForm

	if err := c.Bind(&form); err != nil {
		return c.Redirect(http.StatusSeeOther, "/login?error=Error procesando formulario")
	}

	user, err := h.authService.Login(form.Email, form.Password, form.RecaptchaToken)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login?error="+err.Error())
	}

	// Crear sesión
	if err := h.sessionManager.CreateSession(c, user); err != nil {
		return c.Redirect(http.StatusSeeOther, "/login?error=Error creando sesión")
	}

	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// NUEVO: Mostrar formulario de registro
func (h *AuthHandler) ShowRegister(c echo.Context) error {
	errorMsg := c.QueryParam("error")

	data := map[string]interface{}{
		"Title":        "Registro de Usuario",
		"RecaptchaKey": "6LdjX1gsAAAAAAmGXFf5nW7EkXqx9xukAwyJR35Z",
		"Error":        errorMsg,
		"breadcrumbs":  c.Get("breadcrumbs"),
	}

	return c.Render(http.StatusOK, "register.html", data)
}

// NUEVO: Procesar registro
func (h *AuthHandler) DoRegister(c echo.Context) error {
	var form model.RegisterForm // ← AHORA USA RegisterForm

	if err := c.Bind(&form); err != nil {
		return c.Redirect(http.StatusSeeOther, "/register?error=Error procesando formulario")
	}

	err := h.authService.Register(&form) // ← PASA RegisterForm
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/register?error="+err.Error())
	}

	// Auto-login después del registro
	user, err := h.authService.Login(form.Email, form.Password, form.RecaptchaToken)
	if err == nil {
		h.sessionManager.CreateSession(c, user)
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	}

	return c.Redirect(http.StatusSeeOther, "/login?error=Registro exitoso, por favor inicie sesión")
}

func (h *AuthHandler) Logout(c echo.Context) error {
	h.sessionManager.ClearSession(c)
	return c.Redirect(http.StatusSeeOther, "/login")
}
