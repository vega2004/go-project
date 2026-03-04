package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

// CreateSession - Crea una nueva sesión
func (sm *SessionManager) CreateSession(c echo.Context, user *model.UserAuth) error {
	rolNombre := "user"
	switch user.RoleID {
	case 1:
		rolNombre = "admin"
	case 2:
		rolNombre = "user"
	case 3:
		rolNombre = "editor"
	}

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

	// Serializar sesión para la cookie
	sessionJSON, _ := json.Marshal(session)

	// Configurar cookie
	cookie := new(http.Cookie)
	cookie.Name = sessionName
	cookie.Value = base64.URLEncoding.EncodeToString(sessionJSON)
	cookie.Expires = time.Now().Add(sessionLength)
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode

	// Secure solo en producción (HTTPS)
	if os.Getenv("RAILWAY_ENVIRONMENT") != "" {
		cookie.Secure = true
	}

	c.SetCookie(cookie)

	fmt.Printf("✅ Sesión creada para usuario %d\n", user.ID)
	return nil
}

// GetSession - Obtiene la sesión actual
func (sm *SessionManager) GetSession(c echo.Context) (*model.Session, error) {
	// Primero intentar obtener del contexto
	if session, ok := c.Get(sessionName).(*model.Session); ok {
		return session, nil
	}

	// Si no está en contexto, intentar desde la cookie
	cookie, err := c.Cookie(sessionName)
	if err != nil {
		return nil, fmt.Errorf("no hay cookie de sesión")
	}

	// Decodificar cookie
	decoded, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("error decodificando cookie")
	}

	var session model.Session
	if err := json.Unmarshal(decoded, &session); err != nil {
		return nil, fmt.Errorf("error deserializando sesión")
	}

	// Verificar expiración
	if time.Since(session.LastActivity) > sessionLength {
		sm.ClearSession(c)
		return nil, fmt.Errorf("sesión expirada")
	}

	// Actualizar última actividad
	session.LastActivity = time.Now()

	// Guardar en contexto para futuros usos
	c.Set(sessionName, &session)

	return &session, nil
}

// ClearSession - Elimina la sesión
func (sm *SessionManager) ClearSession(c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = sessionName
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	cookie.Path = "/"
	cookie.HttpOnly = true
	c.SetCookie(cookie)
	c.Set(sessionName, nil)
	fmt.Println("✅ Sesión eliminada")
}
