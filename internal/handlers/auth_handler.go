package handlers

import (
	"log"
	"net/http"
	"time"
	"tu-proyecto/internal/config"
	"tu-proyecto/internal/middleware"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/service"
	"tu-proyecto/internal/utils"

	"github.com/labstack/echo/v4"
)

// ============================================
// ESTRUCTURA Y CONSTRUCTOR
// ============================================

type AuthHandler struct {
	authService    service.AuthService
	sessionManager *utils.SessionManager
	jwtManager     *utils.JWTManager
	env            *config.Env
	csrfMiddleware *middleware.CSRFMiddleware
}

func NewAuthHandler(
	as service.AuthService,
	sm *utils.SessionManager,
	jwtManager *utils.JWTManager,
	env *config.Env,
) *AuthHandler {
	return &AuthHandler{
		authService:    as,
		sessionManager: sm,
		jwtManager:     jwtManager,
		env:            env,
		csrfMiddleware: middleware.NewCSRFMiddleware(env.IsProduction()),
	}
}

// ============================================
// SHOW LOGIN
// ============================================
func (h *AuthHandler) ShowLogin(c echo.Context) error {
	errorMsg := c.QueryParam("error")
	successMsg := c.QueryParam("success")

	h.csrfMiddleware.SetToken(c)

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"Title":        "Iniciar Sesión",
		"RecaptchaKey": h.env.RecaptchaSiteKey,
		"Error":        errorMsg,
		"Success":      successMsg,
		"CSRFToken":    csrfToken,
		"CurrentYear":  time.Now().Year(),
	})
}

// ============================================
// DO LOGIN
// ============================================
// ============================================
// DO LOGIN
// ============================================
func (h *AuthHandler) DoLogin(c echo.Context) error {
	var form models.LoginForm
	if err := c.Bind(&form); err != nil {
		return c.Redirect(http.StatusSeeOther, "/login?error=Error procesando formulario")
	}

	csrfToken := c.FormValue("csrf_token")
	if csrfToken == "" {
		return c.Redirect(http.StatusSeeOther, "/login?error=Token CSRF inválido")
	}

	ipAddress := c.RealIP()

	user, err := h.authService.Login(form.Email, form.Password, form.RecaptchaToken)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login?error="+err.Error())
	}

	// ============================================
	// PETICIÓN JSON (API) - DEVOLVER JWT
	// ============================================
	if c.Request().Header.Get("Accept") == "application/json" ||
		c.Request().Header.Get("Content-Type") == "application/json" {

		perfilNombre := h.authService.GetPerfilNombre(user.PerfilID)

		token, err := h.jwtManager.Generate(
			user.ID,
			user.Email,
			user.Name,
			user.PerfilID,
			perfilNombre,
		)
		if err != nil {
			log.Printf("[ERROR] Error generando JWT para usuario %d: %v", user.ID, err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Error generando token de autenticación",
			})
		}

		log.Printf("[AUDIT] Usuario %d (%s) autenticado por JWT desde IP %s (perfil: %s)",
			user.ID, user.Email, ipAddress, perfilNombre)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"token":   token,
			"user": map[string]interface{}{
				"id":     user.ID,
				"name":   user.Name,
				"email":  user.Email,
				"perfil": perfilNombre,
			},
		})
	}

	// ============================================
	// WEB TRADICIONAL - CREAR SES
	// ============================================
	h.sessionManager.ClearSession(c)

	if err := h.sessionManager.CreateSession(c, user); err != nil {
		log.Printf("[ERROR] Error creando sesión para usuario %d: %v", user.ID, err)
		return c.Redirect(http.StatusSeeOther, "/login?error=Error al iniciar sesión")
	}

	log.Printf("[AUDIT] Usuario %d (%s) inició sesión por sesión desde IP %s (perfil: %d)",
		user.ID, user.Email, ipAddress, user.PerfilID)

	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// ============================================
// SHOW REGISTER
// ============================================
func (h *AuthHandler) ShowRegister(c echo.Context) error {
	errorMsg := c.QueryParam("error")

	h.csrfMiddleware.SetToken(c)

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	log.Printf("[DEBUG] RecaptchaKey: %s", h.env.RecaptchaSiteKey)

	return c.Render(http.StatusOK, "register.html", map[string]interface{}{
		"Title":        "Registro de Usuario",
		"RecaptchaKey": h.env.RecaptchaSiteKey,
		"Error":        errorMsg,
		"CSRFToken":    csrfToken,
		"CurrentYear":  time.Now().Year(),
	})
}

// ============================================
// DO REGISTER
// ============================================
func (h *AuthHandler) DoRegister(c echo.Context) error {
	var form models.RegisterForm
	if err := c.Bind(&form); err != nil {
		return c.Redirect(http.StatusSeeOther, "/register?error=Error procesando formulario")
	}

	csrfToken := c.FormValue("csrf_token")
	if csrfToken == "" {
		return c.Redirect(http.StatusSeeOther, "/register?error=Token CSRF inválido")
	}

	confirmPassword := c.FormValue("confirm_password")
	if form.Password != confirmPassword {
		return c.Redirect(http.StatusSeeOther, "/register?error=Las contraseñas no coinciden")
	}
	form.ConfirmPassword = confirmPassword

	ipAddress := c.RealIP()

	// ✅ PERMITIR TOKEN DE PRUEBA PARA DESARROLLO
	recaptchaToken := form.RecaptchaToken
	if recaptchaToken == "TEST_TOKEN_DESACTIVADO" {
		log.Println("[RECAPTCHA DEBUG] Usando token de prueba - validación omitida")
	}

	if err := h.authService.Register(&form, ipAddress); err != nil {
		return c.Redirect(http.StatusSeeOther, "/register?error="+err.Error())
	}

	user, err := h.authService.Login(form.Email, form.Password, recaptchaToken)
	if err == nil {
		if c.Request().Header.Get("Accept") == "application/json" ||
			c.Request().Header.Get("Content-Type") == "application/json" {

			perfilNombre := h.authService.GetPerfilNombre(user.PerfilID)

			token, err := h.jwtManager.Generate(
				user.ID,
				user.Email,
				user.Name,
				user.PerfilID,
				perfilNombre,
			)
			if err == nil {
				log.Printf("[AUDIT] Nuevo usuario registrado por JWT: %d (%s) desde IP %s",
					user.ID, user.Email, ipAddress)
				return c.JSON(http.StatusOK, map[string]interface{}{
					"success": true,
					"token":   token,
					"user": map[string]interface{}{
						"id":     user.ID,
						"name":   user.Name,
						"email":  user.Email,
						"perfil": perfilNombre,
					},
				})
			}
		}

		if err := h.sessionManager.CreateSession(c, user); err != nil {
			log.Printf("[ERROR] Error creando sesión para nuevo usuario %d: %v", user.ID, err)
		} else {
			log.Printf("[AUDIT] Nuevo usuario registrado por sesión: %d (%s) desde IP %s (perfil: %d)",
				user.ID, user.Email, ipAddress, user.PerfilID)
			return c.Redirect(http.StatusSeeOther, "/dashboard")
		}
	}

	return c.Redirect(http.StatusSeeOther, "/login?success=Registro exitoso, por favor inicie sesión")
}

// ============================================
// LOGOUT
// ============================================
func (h *AuthHandler) Logout(c echo.Context) error {
	userID := c.Get("user_id")
	userEmail := c.Get("user_email")
	ipAddress := c.RealIP()

	if userID != nil {
		log.Printf("[AUDIT] Usuario %d (%v) cerró sesión desde IP %s", userID, userEmail, ipAddress)
	}

	h.sessionManager.ClearSession(c)
	return c.Redirect(http.StatusSeeOther, "/login")
}

// ============================================
// MAINTENANCE
// ============================================
func (h *AuthHandler) Maintenance(c echo.Context) error {
	return c.Render(http.StatusOK, "maintenance.html", map[string]interface{}{
		"Title": "🔧 Sitio en Mantenimiento",
	})
}

// ============================================
// SUCCESS
// ============================================
func (h *AuthHandler) Success(c echo.Context) error {
	return c.Render(http.StatusOK, "success.html", map[string]interface{}{
		"Title": "Operación Exitosa",
	})
}
