package handlers

import (
	"log"
	"net/http"
	"tu-proyecto/internal/config"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/service"
	"tu-proyecto/internal/utils"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService    service.AuthService
	sessionManager *utils.SessionManager
	jwtManager     *utils.JWTManager // ← NUEVO: para JWT
	env            *config.Env
}

// NewAuthHandler - Constructor ACTUALIZADO
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
	}
}

// ShowLogin - Muestra formulario de login
func (h *AuthHandler) ShowLogin(c echo.Context) error {
	errorMsg := c.QueryParam("error")

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"Title":        "Iniciar Sesión",
		"RecaptchaKey": h.env.RecaptchaSiteKey,
		"Error":        errorMsg,
		"CSRFToken":    csrfToken,
	})
}

// DoLogin - Procesa el login (ACTUALIZADO con soporte JWT)
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
	// NUEVO: Si la petición espera JSON (API), devolver JWT
	// ============================================
	if c.Request().Header.Get("Accept") == "application/json" ||
		c.Request().Header.Get("Content-Type") == "application/json" {

		// Generar JWT
		token, err := h.jwtManager.Generate(
			user.ID,
			user.Email,
			user.Name,
			user.RoleID,
			h.authService.GetRolName(user.RoleID),
		)
		if err != nil {
			log.Printf("[ERROR] Error generando JWT para usuario %d: %v", user.ID, err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Error generando token de autenticación",
			})
		}

		log.Printf("[AUDIT] Usuario %d (%s) autenticado por JWT desde IP %s", user.ID, user.Email, ipAddress)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"token":   token,
			"user": map[string]interface{}{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  h.authService.GetRolName(user.RoleID),
			},
		})
	}

	// ============================================
	// PARA WEB TRADICIONAL: Crear sesión
	// ============================================
	h.sessionManager.ClearSession(c)

	if err := h.sessionManager.CreateSession(c, user); err != nil {
		log.Printf("[ERROR] Error creando sesión para usuario %d: %v", user.ID, err)
		return c.Redirect(http.StatusSeeOther, "/login?error=Error al iniciar sesión")
	}

	log.Printf("[AUDIT] Usuario %d (%s) inició sesión por sesión desde IP %s", user.ID, user.Email, ipAddress)

	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// ShowRegister - Muestra formulario de registro
func (h *AuthHandler) ShowRegister(c echo.Context) error {
	errorMsg := c.QueryParam("error")

	csrfToken := c.Get("csrf_token")
	if csrfToken == nil {
		csrfToken = ""
	}

	return c.Render(http.StatusOK, "register.html", map[string]interface{}{
		"Title":        "Registro de Usuario",
		"RecaptchaKey": h.env.RecaptchaSiteKey,
		"Error":        errorMsg,
		"CSRFToken":    csrfToken,
	})
}

// DoRegister - Procesa el registro (ACTUALIZADO con soporte JWT)
func (h *AuthHandler) DoRegister(c echo.Context) error {
	var form models.RegisterForm
	if err := c.Bind(&form); err != nil {
		return c.Redirect(http.StatusSeeOther, "/register?error=Error procesando formulario")
	}

	csrfToken := c.FormValue("csrf_token")
	if csrfToken == "" {
		return c.Redirect(http.StatusSeeOther, "/register?error=Token CSRF inválido")
	}

	ipAddress := c.RealIP()

	if err := h.authService.Register(&form, ipAddress); err != nil {
		return c.Redirect(http.StatusSeeOther, "/register?error="+err.Error())
	}

	// Auto-login después del registro
	user, err := h.authService.Login(form.Email, form.Password, form.RecaptchaToken)
	if err == nil {
		// Si es petición JSON, devolver JWT
		if c.Request().Header.Get("Accept") == "application/json" ||
			c.Request().Header.Get("Content-Type") == "application/json" {

			token, err := h.jwtManager.Generate(
				user.ID,
				user.Email,
				user.Name,
				user.RoleID,
				h.authService.GetRolName(user.RoleID),
			)
			if err == nil {
				log.Printf("[AUDIT] Nuevo usuario registrado por JWT: %d (%s) desde IP %s", user.ID, user.Email, ipAddress)
				return c.JSON(http.StatusOK, map[string]interface{}{
					"success": true,
					"token":   token,
					"user": map[string]interface{}{
						"id":    user.ID,
						"name":  user.Name,
						"email": user.Email,
						"role":  h.authService.GetRolName(user.RoleID),
					},
				})
			}
		}

		// Web tradicional: crear sesión
		if err := h.sessionManager.CreateSession(c, user); err != nil {
			log.Printf("[ERROR] Error creando sesión para nuevo usuario %d: %v", user.ID, err)
		} else {
			log.Printf("[AUDIT] Nuevo usuario registrado por sesión: %d (%s) desde IP %s", user.ID, user.Email, ipAddress)
			return c.Redirect(http.StatusSeeOther, "/dashboard")
		}
	}

	return c.Redirect(http.StatusSeeOther, "/login?success=Registro exitoso, por favor inicie sesión")
}

// Logout - Cierra sesión
func (h *AuthHandler) Logout(c echo.Context) error {
	userID := c.Get("user_id")
	ipAddress := c.RealIP()
	if userID != nil {
		log.Printf("[AUDIT] Usuario %v cerró sesión desde IP %s", userID, ipAddress)
	}
	h.sessionManager.ClearSession(c)
	return c.Redirect(http.StatusSeeOther, "/login")
}

// Maintenance - Página de mantenimiento
func (h *AuthHandler) Maintenance(c echo.Context) error {
	return c.Render(http.StatusOK, "maintenance.html", map[string]interface{}{
		"Title": "🔧 Sitio en Mantenimiento",
	})
}

// Success - Página de éxito
func (h *AuthHandler) Success(c echo.Context) error {
	return c.Render(http.StatusOK, "success.html", map[string]interface{}{
		"Title": "Operación Exitosa",
	})
}

// ShowPerfilJSON - Devuelve el perfil del usuario en formato JSON (para API con JWT)
func (h *PerfilHandler) ShowPerfilJSON(c echo.Context) error {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No autenticado"})
	}

	perfil, err := h.service.GetPerfil(userID)
	if err != nil {
		log.Printf("[ERROR] PerfilHandler.ShowPerfilJSON: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al cargar perfil"})
	}

	// Obtener datos del usuario
	userName := c.Get("user_name")
	userEmail := c.Get("user_email")
	userRole := c.Get("user_role")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user_id":          userID,
			"name":             userName,
			"email":            userEmail,
			"role":             userRole,
			"bio":              perfil.Bio,
			"direccion":        perfil.Direccion,
			"telefono_alterno": perfil.TelefonoAlterno,
			"foto_path":        perfil.FotoPath,
			"updated_at":       perfil.UpdatedAt,
		},
	})
}
