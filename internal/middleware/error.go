package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type ErrorInfo struct {
	ID         string    `json:"id"`
	Code       int       `json:"code"`
	Path       string    `json:"path"`
	Method     string    `json:"method"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	UserID     int       `json:"user_id"`
	Error      string    `json:"error"`
	StackTrace string    `json:"stack_trace,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

type ErrorMiddleware struct {
	mu           sync.Mutex
	errorCounts  map[string][]time.Time
	isDev        bool
	alertWebhook string
}

func NewErrorMiddleware(isDev bool, alertWebhook string) *ErrorMiddleware {
	m := &ErrorMiddleware{
		errorCounts:  make(map[string][]time.Time),
		isDev:        isDev,
		alertWebhook: alertWebhook,
	}

	// Limpieza periódica de contadores
	go m.cleanupCounters()

	return m
}

func (m *ErrorMiddleware) Recover(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic recuperado: %v", r)
				stackTrace := string(debug.Stack())
				log.Printf("[PANIC] %v\n%s", err, stackTrace)

				errorInfo := m.buildErrorInfo(c, http.StatusInternalServerError, err, stackTrace)
				m.saveErrorToFile(errorInfo)

				if !m.isDev {
					m.sendAlert(errorInfo)
				}

				m.handleError(c, http.StatusInternalServerError, err, stackTrace)
			}
		}()
		return next(c)
	}
}

func (m *ErrorMiddleware) ErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	errorInfo := m.buildErrorInfo(c, code, err, "")
	m.saveErrorToFile(errorInfo)

	if code >= 500 && !m.isDev {
		m.sendAlert(errorInfo)
	}

	m.handleError(c, code, err, "")
}

func (m *ErrorMiddleware) NotFound(c echo.Context) error {
	err := fmt.Errorf("página no encontrada: %s", c.Request().URL.Path)
	return m.handleError(c, http.StatusNotFound, err, "")
}

func (m *ErrorMiddleware) MethodNotAllowed(c echo.Context) error {
	err := fmt.Errorf("método no permitido: %s", c.Request().Method)
	return m.handleError(c, http.StatusMethodNotAllowed, err, "")
}

func (m *ErrorMiddleware) buildErrorInfo(c echo.Context, code int, err error, stackTrace string) ErrorInfo {
	userID := 0
	if uid, ok := c.Get("user_id").(int); ok {
		userID = uid
	}

	return ErrorInfo{
		ID:         fmt.Sprintf("%d-%d", time.Now().UnixNano(), code),
		Code:       code,
		Path:       c.Request().URL.Path,
		Method:     c.Request().Method,
		IP:         c.RealIP(),
		UserAgent:  c.Request().UserAgent(),
		UserID:     userID,
		Error:      err.Error(),
		StackTrace: stackTrace,
		Timestamp:  time.Now(),
	}
}

func (m *ErrorMiddleware) saveErrorToFile(errorInfo ErrorInfo) {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.MkdirAll("logs", 0755)
	}

	filename := fmt.Sprintf("logs/errors_%s.log", time.Now().Format("2006-01-02"))

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error abriendo archivo de log: %v", err)
		return
	}
	defer f.Close()

	data, _ := json.MarshalIndent(errorInfo, "", "  ")
	f.WriteString(string(data) + "\n")
	f.WriteString(strings.Repeat("-", 80) + "\n")
}

func (m *ErrorMiddleware) sendAlert(errorInfo ErrorInfo) {
	if m.alertWebhook == "" {
		return
	}

	// Enviar alerta a webhook (Slack, Discord, etc.)
	go func() {
		// Implementar según necesidad
		log.Printf("[ALERT] Error %d: %s", errorInfo.Code, errorInfo.Error)
	}()
}

func (m *ErrorMiddleware) handleError(c echo.Context, code int, err error, stackTrace string) error {
	errorID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), code)
	message := m.getErrorMessage(code)

	if m.isRateLimited(c.RealIP()) {
		return c.JSON(http.StatusTooManyRequests, map[string]string{
			"error": "Demasiados errores, intente más tarde",
		})
	}

	if m.isJSONRequest(c) {
		return c.JSON(code, map[string]interface{}{
			"success":  false,
			"code":     code,
			"message":  message,
			"error_id": errorID,
		})
	}

	data := map[string]interface{}{
		"Title":     m.getErrorTitle(code),
		"Code":      code,
		"Message":   message,
		"ErrorID":   errorID,
		"Timestamp": time.Now().Format("02/01/2006 15:04:05"),
		"Path":      c.Request().URL.Path,
		"IsDev":     m.isDev,
	}

	if m.isDev && err != nil {
		data["Details"] = err.Error()
		if stackTrace != "" {
			data["StackTrace"] = stackTrace
		}
	}

	return c.Render(code, "error.html", data)
}

func (m *ErrorMiddleware) isJSONRequest(c echo.Context) bool {
	return c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		c.Request().Header.Get("Accept") == "application/json"
}

func (m *ErrorMiddleware) isRateLimited(ip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	window := now.Add(-1 * time.Minute)

	// Limpiar intentos antiguos
	var recent []time.Time
	for _, t := range m.errorCounts[ip] {
		if t.After(window) {
			recent = append(recent, t)
		}
	}

	if len(recent) > 30 {
		return true
	}

	m.errorCounts[ip] = append(recent, now)
	return false
}

func (m *ErrorMiddleware) cleanupCounters() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for ip, attempts := range m.errorCounts {
			var recent []time.Time
			for _, t := range attempts {
				if t.After(now.Add(-5 * time.Minute)) {
					recent = append(recent, t)
				}
			}
			if len(recent) == 0 {
				delete(m.errorCounts, ip)
			} else {
				m.errorCounts[ip] = recent
			}
		}
		m.mu.Unlock()
	}
}

func (m *ErrorMiddleware) getErrorMessage(code int) string {
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

	if msg, ok := messages[code]; ok {
		return msg
	}
	return "Ha ocurrido un error inesperado."
}

func (m *ErrorMiddleware) getErrorTitle(code int) string {
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

	if title, ok := titles[code]; ok {
		return title
	}
	return "Error"
}
