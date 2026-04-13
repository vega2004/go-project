package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response - Estructura estándar de respuesta
type Response struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JSONResponse - Envía respuesta JSON estandarizada
func JSONResponse(c echo.Context, status int, success bool, message string, data interface{}, err string) error {
	return c.JSON(status, Response{
		Success: success,
		Code:    status,
		Message: message,
		Data:    data,
		Error:   err,
	})
}

// Success - Respuesta de éxito
func Success(c echo.Context, message string, data interface{}) error {
	return JSONResponse(c, http.StatusOK, true, message, data, "")
}

// Created - Respuesta de creación exitosa
func Created(c echo.Context, message string, data interface{}) error {
	return JSONResponse(c, http.StatusCreated, true, message, data, "")
}

// BadRequest - Respuesta de error 400
func BadRequest(c echo.Context, message string) error {
	return JSONResponse(c, http.StatusBadRequest, false, "", nil, message)
}

// Unauthorized - Respuesta de error 401
func Unauthorized(c echo.Context, message string) error {
	return JSONResponse(c, http.StatusUnauthorized, false, "", nil, message)
}

// Forbidden - Respuesta de error 403
func Forbidden(c echo.Context, message string) error {
	return JSONResponse(c, http.StatusForbidden, false, "", nil, message)
}

// NotFound - Respuesta de error 404
func NotFound(c echo.Context, message string) error {
	return JSONResponse(c, http.StatusNotFound, false, "", nil, message)
}

// InternalError - Respuesta de error 500
func InternalError(c echo.Context, message string) error {
	return JSONResponse(c, http.StatusInternalServerError, false, "", nil, message)
}

// ValidationError - Respuesta para errores de validación
func ValidationError(c echo.Context, errors map[string]string) error {
	return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
		"success": false,
		"code":    422,
		"errors":  errors,
	})
}
