package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CSRFMiddleware struct {
	cookieName string
	isSecure   bool
}

// NewCSRFMiddleware - Constructor que recibe isSecure para entorno producción
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
	token := m.GenerateToken()
	c.SetCookie(&http.Cookie{
		Name:     m.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
		Secure:   m.isSecure,
	})
	c.Set("csrf_token", token)
}

func (m *CSRFMiddleware) Protect(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Method == "POST" || c.Request().Method == "PUT" || c.Request().Method == "DELETE" {
			token := c.FormValue("csrf_token")
			if token == "" {
				token = c.Request().Header.Get("X-CSRF-Token")
			}

			cookie, err := c.Cookie(m.cookieName)
			if err != nil || token == "" || cookie.Value != token {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Token CSRF inválido",
				})
			}
		}
		return next(c)
	}
}
