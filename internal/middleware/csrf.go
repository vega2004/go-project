package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CSRFMiddleware struct {
	cookieName string
	isSecure   bool
}

func NewCSRFMiddleware(isSecure bool) *CSRFMiddleware {
	return &CSRFMiddleware{
		cookieName: "csrf_token",
		isSecure:   isSecure,
	}
}

func (m *CSRFMiddleware) GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (m *CSRFMiddleware) SetToken(c echo.Context) {
	// Verificar si ya existe un token en la cookie
	cookie, err := c.Cookie(m.cookieName)
	var token string

	if err == nil && cookie.Value != "" {
		// Reutilizar token existente
		token = cookie.Value
		log.Printf("[CSRF] Reutilizando token existente: %s", token)
	} else {
		// Generar nuevo token
		token = m.GenerateToken()
		c.SetCookie(&http.Cookie{
			Name:     m.cookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: false,
			SameSite: http.SameSiteLaxMode,
			Secure:   m.isSecure,
		})
		log.Printf("[CSRF] Nuevo token generado: %s", token)
	}

	c.Set("csrf_token", token)
}

func (m *CSRFMiddleware) Protect(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Method == "POST" || c.Request().Method == "PUT" || c.Request().Method == "DELETE" {
			// ✅ PRIORIDAD: Primero intentar obtener del header
			token := c.Request().Header.Get("X-CSRF-Token")

			// Si no está en header, intentar del formulario
			if token == "" {
				token = c.FormValue("csrf_token")
			}

			cookie, err := c.Cookie(m.cookieName)

			log.Printf("[CSRF DEBUG] Method: %s", c.Request().Method)
			log.Printf("[CSRF DEBUG] Header X-CSRF-Token: %s", c.Request().Header.Get("X-CSRF-Token"))
			log.Printf("[CSRF DEBUG] Form token: %s", c.FormValue("csrf_token"))
			log.Printf("[CSRF DEBUG] Token usado: %s", token)
			if err == nil {
				log.Printf("[CSRF DEBUG] Cookie token: %s", cookie.Value)
			}

			if err != nil || token == "" || cookie.Value != token {
				log.Printf("[CSRF ERROR] Token inválido - err: %v, token vacío: %v", err, token == "")
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Token CSRF inválido",
				})
			}
		}
		return next(c)
	}
}
