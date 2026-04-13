package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"tu-proyecto/internal/models"

	"github.com/labstack/echo/v4"
)

const sessionName = "user_session"

type SessionManager struct {
	secret     string
	isSecure   bool
	sessionAge time.Duration
}

func NewSessionManager(secret string, isProduction bool) *SessionManager {
	gob.Register(models.Session{})
	return &SessionManager{
		secret:     secret,
		isSecure:   isProduction,
		sessionAge: 24 * time.Hour,
	}
}

func (sm *SessionManager) sign(data []byte) string {
	h := hmac.New(sha256.New, []byte(sm.secret))
	h.Write(data)
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func (sm *SessionManager) verify(data []byte, signature string) bool {
	expected := sm.sign(data)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (sm *SessionManager) encodeSession(session *models.Session) (string, error) {
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("error serializando sesión: %w", err)
	}
	encoded := base64.URLEncoding.EncodeToString(sessionJSON)
	signature := sm.sign(sessionJSON)
	return encoded + "." + signature, nil
}

func (sm *SessionManager) decodeSession(value string) (*models.Session, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("formato de sesión inválido")
	}
	encoded := parts[0]
	signature := parts[1]
	sessionJSON, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("error decodificando sesión: %w", err)
	}
	if !sm.verify(sessionJSON, signature) {
		return nil, fmt.Errorf("firma de sesión inválida")
	}
	var session models.Session
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		return nil, fmt.Errorf("error deserializando sesión: %w", err)
	}
	return &session, nil
}

func (sm *SessionManager) CreateSession(c echo.Context, user *models.UserAuth) error {
	rolNombre := "usuario"
	switch user.RoleID {
	case 1:
		rolNombre = "administrador"
	case 2:
		rolNombre = "usuario"
	case 3:
		rolNombre = "editor"
	}
	session := &models.Session{
		UserID:       user.ID,
		Email:        user.Email,
		Name:         user.Name,
		RoleID:       user.RoleID,
		RoleNombre:   rolNombre,
		LastActivity: time.Now(),
	}
	c.Set(sessionName, session)
	encoded, err := sm.encodeSession(session)
	if err != nil {
		return fmt.Errorf("error codificando sesión: %w", err)
	}
	cookie := &http.Cookie{
		Name:     sessionName,
		Value:    encoded,
		Expires:  time.Now().Add(sm.sessionAge),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   sm.isSecure,
	}
	c.SetCookie(cookie)
	return nil
}

func (sm *SessionManager) GetSession(c echo.Context) (*models.Session, error) {
	if session, ok := c.Get(sessionName).(*models.Session); ok {
		return session, nil
	}
	cookie, err := c.Cookie(sessionName)
	if err != nil {
		return nil, fmt.Errorf("no hay sesión activa")
	}
	session, err := sm.decodeSession(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("error decodificando sesión: %w", err)
	}
	if time.Since(session.LastActivity) > sm.sessionAge {
		sm.ClearSession(c)
		return nil, fmt.Errorf("sesión expirada")
	}
	session.LastActivity = time.Now()
	c.Set(sessionName, session)
	return session, nil
}

func (sm *SessionManager) ClearSession(c echo.Context) {
	cookie := &http.Cookie{
		Name:     sessionName,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	}
	c.SetCookie(cookie)
	c.Set(sessionName, nil)
}

func (sm *SessionManager) UpdateSession(c echo.Context, updatedSession *models.Session) error {
	c.Set(sessionName, updatedSession)
	encoded, err := sm.encodeSession(updatedSession)
	if err != nil {
		return fmt.Errorf("error codificando sesión: %w", err)
	}
	cookie, err := c.Cookie(sessionName)
	if err != nil {
		cookie = &http.Cookie{}
	}
	cookie.Name = sessionName
	cookie.Value = encoded
	cookie.Expires = time.Now().Add(sm.sessionAge)
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Secure = sm.isSecure
	c.SetCookie(cookie)
	return nil
}
