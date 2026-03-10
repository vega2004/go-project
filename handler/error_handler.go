package handler

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

type ErrorInfo struct {
	Code           int
	Title          string
	Message        string
	Details        string
	Timestamp      string
	ErrorID        string
	ShowReport     bool
	ShowStackTrace bool
	StackTrace     string
	RequestURI     string
	UserID         int
	UserRole       string
}

// ShowError - Muestra página de error personalizada
func ShowError(c echo.Context, code int, message string, err error) error {
	errorID := uuid.New().String()

	// Obtener información del usuario (si está autenticado)
	userID, _ := c.Get("user_id").(int)
	userRole, _ := c.Get("user_role").(string)

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

	// Registrar error en log
	fmt.Printf("[ERROR %d] [ID: %s] %s - Usuario: %d - Path: %s - Error: %v\n",
		code, errorID, message, userID, c.Request().URL.Path, err)

	// Si es error 500, guardar en log especial y enviar stack trace
	if code >= 500 {
		logErrorToFile(errorID, code, message, err, c)
	}

	// Preparar datos para el template
	data := map[string]interface{}{
		"Title":          title,
		"Code":           code,
		"Message":        message,
		"Details":        getErrorMessage(err),
		"Timestamp":      time.Now().Format("02/01/2006 15:04:05"),
		"ErrorID":        errorID,
		"ShowReport":     true,
		"ShowStackTrace": code >= 500 && os.Getenv("ENVIRONMENT") == "development",
		"StackTrace":     getStackTrace(err),
		"breadcrumbs": []map[string]string{
			{"name": "Inicio", "url": "/"},
			{"name": fmt.Sprintf("Error %d", code), "url": ""},
		},
		"UserID":     userID,
		"UserRole":   userRole,
		"RequestURI": c.Request().URL.Path,
	}

	// Establecer código de respuesta
	c.Response().Status = code

	// Intentar renderizar template de error, si falla mostrar mensaje simple
	err = c.Render(http.StatusOK, "error.html", data)
	if err != nil {
		// Fallback: mostrar error simple en texto
		return c.HTML(code, fmt.Sprintf(`
			<html>
			<head><title>Error %d</title></head>
			<body style="font-family: Arial; text-align: center; padding: 50px;">
				<h1 style="color: #dc3545;">Error %d</h1>
				<h2>%s</h2>
				<p>%s</p>
				<p><small>ID: %s</small></p>
				<a href="/dashboard" style="color: #5483B3;">Volver al Dashboard</a>
			</body>
			</html>
		`, code, code, title, message, errorID))
	}

	return nil
}

// CustomHTTPErrorHandler - Manejador personalizado de errores para Echo
func CustomHTTPErrorHandler(err error, c echo.Context) {
	// Código por defecto
	code := http.StatusInternalServerError
	message := "Ha ocurrido un error inesperado"

	// Verificar si es error HTTP
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = fmt.Sprintf("%v", he.Message)
	}

	// Manejar errores específicos
	switch code {
	case http.StatusNotFound:
		message = "La página que buscas no existe o ha sido movida"
	case http.StatusUnauthorized:
		message = "Necesitas iniciar sesión para acceder a esta página"
	case http.StatusForbidden:
		message = "No tienes permisos para acceder a esta página"
	case http.StatusMethodNotAllowed:
		message = "Método de solicitud no permitido"
	case http.StatusRequestTimeout:
		message = "La solicitud ha excedido el tiempo de espera"
	case http.StatusTooManyRequests:
		message = "Has realizado demasiadas solicitudes. Espera un momento"
	case http.StatusBadRequest:
		if message == "" {
			message = "La solicitud no es válida"
		}
	}

	// Verificar si la petición espera JSON
	if c.Request().Header.Get("Accept") == "application/json" ||
		c.Request().Header.Get("Content-Type") == "application/json" {
		c.JSON(code, map[string]interface{}{
			"error":   true,
			"code":    code,
			"message": message,
			"id":      uuid.New().String(),
		})
		return
	}

	// Mostrar página de error
	ShowError(c, code, message, err)
}

// logErrorToFile - Guarda errores en archivo de log
func logErrorToFile(errorID string, code int, message string, err error, c echo.Context) {
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

	// Obtener stack trace
	stack := debug.Stack()

	// Obtener información de la request
	userID, _ := c.Get("user_id").(int)
	userRole, _ := c.Get("user_role").(string)

	logEntry := fmt.Sprintf("[%s] [%s] ERROR %d\n",
		time.Now().Format("2006-01-02 15:04:05"),
		errorID,
		code,
	)
	logEntry += fmt.Sprintf("Message: %s\n", message)
	logEntry += fmt.Sprintf("Path: %s\n", c.Request().URL.Path)
	logEntry += fmt.Sprintf("Method: %s\n", c.Request().Method)
	logEntry += fmt.Sprintf("UserID: %d\n", userID)
	logEntry += fmt.Sprintf("UserRole: %s\n", userRole)
	logEntry += fmt.Sprintf("IP: %s\n", c.RealIP())
	logEntry += fmt.Sprintf("UserAgent: %s\n", c.Request().UserAgent())
	if err != nil {
		logEntry += fmt.Sprintf("Error: %v\n", err)
	}
	logEntry += fmt.Sprintf("Stack Trace:\n%s\n", stack)
	logEntry += strings.Repeat("-", 80) + "\n"

	f.WriteString(logEntry)
}

// getErrorMessage - Obtiene mensaje de error amigable
func getErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	// Errores comunes
	switch err.Error() {
	case "record not found":
		return "El registro solicitado no existe"
	case "duplicate key value violates unique constraint":
		return "Ya existe un registro con esos datos"
	case "connection refused":
		return "No se pudo conectar con el servidor"
	case "context deadline exceeded":
		return "La operación excedió el tiempo de espera"
	default:
		return err.Error()
	}
}

// getStackTrace - Obtiene stack trace como string
func getStackTrace(err error) string {
	if err == nil {
		return ""
	}
	return string(debug.Stack())
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
				err := fmt.Errorf("panic recuperado: %v", r)
				CustomHTTPErrorHandler(err, c)
			}
		}()
		return next(c)
	}
}
