package handlers

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ErrorInfo - Estructura con información del error
type ErrorInfo struct {
	Code           int         `json:"code"`
	Title          string      `json:"title"`
	Message        string      `json:"message"`
	Details        string      `json:"details,omitempty"`
	Timestamp      string      `json:"timestamp"`
	ErrorID        string      `json:"error_id"`
	RequestURI     string      `json:"request_uri,omitempty"`
	Method         string      `json:"method,omitempty"`
	UserID         interface{} `json:"user_id,omitempty"`
	UserRole       interface{} `json:"user_role,omitempty"`
	ShowStackTrace bool        `json:"-"`
	StackTrace     string      `json:"-"`
}

// ErrorHandler - Manejador global de errores
type ErrorHandler struct{}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// HandleError - Muestra página de error personalizada
func (h *ErrorHandler) HandleError(c echo.Context, code int, err error) error {
	errorID := uuid.New().String()

	// Obtener información del usuario (si está autenticado)
	userID := c.Get("user_id")
	userRole := c.Get("user_role")

	// Títulos según código de error
	titles := map[int]string{
		400: "Solicitud Incorrecta",
		401: "No Autorizado",
		403: "Acceso Denegado",
		404: "Página No Encontrada",
		405: "Método No Permitido",
		408: "Tiempo de Espera Agotado",
		429: "Demasiadas Solicitudes",
		500: "Error Interno del Servidor",
		502: "Bad Gateway",
		503: "Servicio No Disponible",
		504: "Gateway Timeout",
	}

	title := titles[code]
	if title == "" {
		title = "Error Inesperado"
	}

	// Mensajes amigables según código
	messages := map[int]string{
		400: "La solicitud no pudo ser procesada debido a datos inválidos.",
		401: "Necesitas iniciar sesión para acceder a esta página.",
		403: "No tienes permisos para acceder a esta página.",
		404: "La página que buscas no existe o ha sido movida.",
		405: "El método de solicitud no está permitido.",
		408: "La solicitud ha excedido el tiempo de espera.",
		429: "Has realizado demasiadas solicitudes. Espera un momento.",
		500: "Ocurrió un error inesperado. Nuestro equipo ha sido notificado.",
		502: "El servidor upstream respondió con un error.",
		503: "El servicio no está disponible momentáneamente.",
		504: "El servidor upstream no respondió a tiempo.",
	}

	message := messages[code]
	if message == "" {
		message = "Ha ocurrido un error inesperado."
	}

	// Registrar error en log
	h.logError(errorID, code, err, c)

	// Preparar datos para el template
	data := map[string]interface{}{
		"Title":          title,
		"Code":           code,
		"Message":        message,
		"Details":        h.getErrorMessage(err),
		"Timestamp":      time.Now().Format("02/01/2006 15:04:05"),
		"ErrorID":        errorID,
		"ShowStackTrace": os.Getenv("APP_ENV") == "development",
		"StackTrace":     h.getStackTrace(err),
		"RequestURI":     c.Request().RequestURI,
		"Method":         c.Request().Method,
		"UserID":         userID,
		"UserRole":       userRole,
		"breadcrumbs": []map[string]string{
			{"name": "Inicio", "url": "/"},
			{"name": fmt.Sprintf("Error %d", code), "url": ""},
		},
	}

	// Si la petición espera JSON, devolver JSON
	if c.Request().Header.Get("Accept") == "application/json" ||
		c.Request().Header.Get("Content-Type") == "application/json" {
		return c.JSON(code, ErrorInfo{
			Code:       code,
			Title:      title,
			Message:    message,
			Timestamp:  time.Now().Format("02/01/2006 15:04:05"),
			ErrorID:    errorID,
			RequestURI: c.Request().RequestURI,
			Method:     c.Request().Method,
		})
	}

	// Intentar renderizar template de error
	errRender := c.Render(code, "error.html", data)
	if errRender != nil {
		// Fallback: mostrar error simple en HTML
		return c.HTML(code, fmt.Sprintf(`
			<!DOCTYPE html>
			<html lang="es">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Error %d - %s</title>
				<style>
					body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background: #f5f5f5; }
					.container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
					h1 { color: #dc3545; margin-bottom: 20px; }
					.code { font-size: 72px; font-weight: bold; color: #dc3545; margin-bottom: 20px; }
					.message { font-size: 18px; color: #333; margin-bottom: 20px; }
					.id { font-size: 12px; color: #999; margin-top: 20px; }
					.btn { display: inline-block; padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
					.btn:hover { background: #0056b3; }
				</style>
			</head>
			<body>
				<div class="container">
					<div class="code">%d</div>
					<h1>%s</h1>
					<div class="message">%s</div>
					<div class="id">ID de error: %s</div>
					<a href="/dashboard" class="btn">Volver al Dashboard</a>
				</div>
			</body>
			</html>
		`, code, title, code, title, message, errorID))
	}

	return nil
}

// logError - Registra el error en archivo de log
func (h *ErrorHandler) logError(errorID string, code int, err error, c echo.Context) {
	// Crear directorio logs si no existe
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, 0755)
	}

	logFile := fmt.Sprintf("logs/errors_%s.log", time.Now().Format("2006-01-02"))

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error abriendo archivo de log: %v\n", err)
		return
	}
	defer f.Close()

	// Obtener información del usuario
	userID := c.Get("user_id")
	userRole := c.Get("user_role")

	logEntry := fmt.Sprintf(`
========================================
[%s] [ID: %s] ERROR %d
Path: %s
Method: %s
IP: %s
UserAgent: %s
UserID: %v
UserRole: %v
Error: %v
Stack Trace:
%s
========================================
`,
		time.Now().Format("2006-01-02 15:04:05"),
		errorID,
		code,
		c.Request().URL.Path,
		c.Request().Method,
		c.RealIP(),
		c.Request().UserAgent(),
		userID,
		userRole,
		err,
		string(debug.Stack()),
	)

	f.WriteString(logEntry)
}

// getErrorMessage - Obtiene mensaje de error amigable
func (h *ErrorHandler) getErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// Errores comunes de base de datos
	switch {
	case strings.Contains(errStr, "duplicate key"):
		return "Ya existe un registro con esos datos."
	case strings.Contains(errStr, "foreign key"):
		return "No se puede eliminar porque hay registros relacionados."
	case strings.Contains(errStr, "not found"):
		return "El registro solicitado no existe."
	case strings.Contains(errStr, "connection refused"):
		return "No se pudo conectar con la base de datos."
	case strings.Contains(errStr, "timeout"):
		return "La operación excedió el tiempo de espera."
	default:
		return errStr
	}
}

// getStackTrace - Obtiene stack trace como string
func (h *ErrorHandler) getStackTrace(err error) string {
	if err == nil {
		return ""
	}
	return string(debug.Stack())
}

// ============================================
// FUNCIONES GLOBALES PARA USAR EN TODA LA APP
// ============================================

// ShowError - Función global para mostrar errores
func ShowError(c echo.Context, code int, message string, err error) error {
	handler := NewErrorHandler()
	return handler.HandleError(c, code, err)
}

// NotFoundHandler - Manejador para rutas no encontradas
func NotFoundHandler(c echo.Context) error {
	return ShowError(c, http.StatusNotFound, "Página no encontrada", nil)
}

// MethodNotAllowedHandler - Manejador para métodos no permitidos
func MethodNotAllowedHandler(c echo.Context) error {
	return ShowError(c, http.StatusMethodNotAllowed, "Método no permitido", nil)
}

// PanicHandler - Recupera de panics y muestra error 500
func PanicHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch v := r.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("%v", v)
				}
				ShowError(c, http.StatusInternalServerError, "Error interno del servidor", err)
			}
		}()
		return next(c)
	}
}

// CustomHTTPErrorHandler - Manejador personalizado de errores para Echo
func CustomHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Ha ocurrido un error inesperado"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = fmt.Sprintf("%v", he.Message)
	}

	ShowError(c, code, message, err)
}
