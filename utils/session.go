package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"
	"tu-proyecto/model"

	"github.com/labstack/echo/v4"
)

const (
	sessionName   = "user_session"
	sessionLength = 24 * time.Hour
)

type SessionManager struct {
	secret string
}

func NewSessionManager() *SessionManager {
	// Registrar el tipo para gob
	gob.Register(model.Session{})

	return &SessionManager{
		secret: generateSecret(),
	}
}

func generateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// CreateSession - Crea una nueva sesión con información del usuario y su rol
func (sm *SessionManager) CreateSession(c echo.Context, user *model.UserAuth) error {
	// Determinar nombre del rol basado en role_id
	rolNombre := "user"
	switch user.RoleID {
	case 1:
		rolNombre = "admin"
	case 2:
		rolNombre = "user"
	case 3:
		rolNombre = "editor"
	}

	// Crear sesión con toda la información
	session := &model.Session{
		UserID:       user.ID,
		Email:        user.Email,
		Name:         user.Name,
		RoleID:       user.RoleID,
		RoleNombre:   rolNombre,
		LastActivity: time.Now(),
	}

	// Guardar en contexto
	c.Set(sessionName, session)

	// Configurar cookie segura
	cookie := new(http.Cookie)
	cookie.Name = sessionName
	cookie.Value = fmt.Sprintf("%d", user.ID)
	cookie.Expires = time.Now().Add(sessionLength)
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode

	c.SetCookie(cookie)

	return nil
}

// GetSession - Obtiene la sesión actual y actualiza última actividad
func (sm *SessionManager) GetSession(c echo.Context) (*model.Session, error) {
	session, ok := c.Get(sessionName).(*model.Session)
	if !ok {
		return nil, fmt.Errorf("sesión no encontrada")
	}

	// Verificar expiración
	if time.Since(session.LastActivity) > sessionLength {
		sm.ClearSession(c)
		return nil, fmt.Errorf("sesión expirada")
	}

	// Actualizar última actividad
	session.LastActivity = time.Now()

	return session, nil
}

// ClearSession - Elimina la sesión actual
func (sm *SessionManager) ClearSession(c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = sessionName
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	cookie.Path = "/"
	cookie.HttpOnly = true
	c.SetCookie(cookie)
	c.Set(sessionName, nil)
}

// GetUserID - Obtiene el ID del usuario de la sesión
func (sm *SessionManager) GetUserID(c echo.Context) (int, error) {
	session, err := sm.GetSession(c)
	if err != nil {
		return 0, err
	}
	return session.UserID, nil
}

// GetUserRole - Obtiene el rol del usuario de la sesión
func (sm *SessionManager) GetUserRole(c echo.Context) (int, string, error) {
	session, err := sm.GetSession(c)
	if err != nil {
		return 0, "", err
	}
	return session.RoleID, session.RoleNombre, nil
}

// IsAdmin - Verifica si el usuario actual es administrador
func (sm *SessionManager) IsAdmin(c echo.Context) bool {
	session, err := sm.GetSession(c)
	if err != nil {
		return false
	}
	return session.RoleID == 1
}

// RefreshSession - Renueva la sesión (útil después de cambios de rol)
func (sm *SessionManager) RefreshSession(c echo.Context, user *model.UserAuth) error {
	// Eliminar sesión actual
	sm.ClearSession(c)

	// Crear nueva sesión con datos actualizados
	return sm.CreateSession(c, user)
}

// GetSessionData - Obtiene todos los datos de la sesión (útil para templates)
func (sm *SessionManager) GetSessionData(c echo.Context) map[string]interface{} {
	session, err := sm.GetSession(c)
	if err != nil {
		return map[string]interface{}{
			"Authenticated": false,
		}
	}

	return map[string]interface{}{
		"Authenticated": true,
		"UserID":        session.UserID,
		"UserName":      session.Name,
		"UserEmail":     session.Email,
		"UserRoleID":    session.RoleID,
		"UserRole":      session.RoleNombre,
	}
}
