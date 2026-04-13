package models

import (
	"time"
)

// ============================================
// RESPUESTAS ESTÁNDAR PARA API
// ============================================

// Response - Estructura base para todas las respuestas JSON
type Response struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse - Respuesta con paginación
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Code       int         `json:"code"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
}

// ValidationErrorResponse - Respuesta para errores de validación
type ValidationErrorResponse struct {
	Success bool              `json:"success"`
	Code    int               `json:"code"`
	Errors  map[string]string `json:"errors"`
}

// ============================================
// RESPUESTAS DE AUTENTICACIÓN
// ============================================

// LoginResponse - Respuesta después de login exitoso
type LoginResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Redirect  string `json:"redirect"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
	UserRole  string `json:"user_role"`
}

// LogoutResponse - Respuesta después de logout
type LogoutResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Redirect string `json:"redirect"`
}

// RegisterResponse - Respuesta después de registro
type RegisterResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Redirect string `json:"redirect"`
	UserID   int    `json:"user_id,omitempty"`
}

// ============================================
// RESPUESTAS DE CRUD
// ============================================

// CreateResponse - Respuesta después de crear un registro
type CreateResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	ID      int         `json:"id"`
	Data    interface{} `json:"data,omitempty"`
}

// UpdateResponse - Respuesta después de actualizar un registro
type UpdateResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	ID      int         `json:"id"`
	Data    interface{} `json:"data,omitempty"`
}

// DeleteResponse - Respuesta después de eliminar un registro
type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      int    `json:"id"`
}

// DetailResponse - Respuesta con detalle de un registro
type DetailResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// ListResponse - Respuesta con lista de registros
type ListResponse struct {
	Success bool          `json:"success"`
	Data    []interface{} `json:"data"`
	Total   int           `json:"total"`
}

// ============================================
// RESPUESTAS DE PERMISOS
// ============================================

// PermisosResponse - Respuesta con permisos de un usuario
type PermisosResponse struct {
	Success  bool                   `json:"success"`
	Permisos map[string]PermisoInfo `json:"permisos"`
	Modulos  []string               `json:"modulos"`
}

// PermisoInfo - Información de un permiso específico
type PermisoInfo struct {
	Ver      bool `json:"ver"`
	Crear    bool `json:"crear"`
	Editar   bool `json:"editar"`
	Eliminar bool `json:"eliminar"`
	Detalle  bool `json:"detalle"`
}

// ============================================
// RESPUESTAS DE DASHBOARD
// ============================================

// DashboardStats - Estadísticas del dashboard
type DashboardStats struct {
	TotalUsuarios   int `json:"total_usuarios"`
	TotalPerfiles   int `json:"total_perfiles"`
	TotalModulos    int `json:"total_modulos"`
	UsuariosActivos int `json:"usuarios_activos"`
	VisitasHoy      int `json:"visitas_hoy"`
}

// DashboardResponse - Respuesta completa del dashboard
type DashboardResponse struct {
	Success        bool             `json:"success"`
	Stats          DashboardStats   `json:"stats"`
	RecentActivity []ActivityItem   `json:"recent_activity"`
	Modulos        []ModuloMenuItem `json:"modulos"`
	UserInfo       UserInfo         `json:"user_info"`
}

// ActivityItem - Item de actividad reciente
type ActivityItem struct {
	Usuario   string    `json:"usuario"`
	Accion    string    `json:"accion"`
	Modulo    string    `json:"modulo,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip,omitempty"`
}

// ModuloMenuItem - Item del menú de módulos
type ModuloMenuItem struct {
	Nombre string `json:"nombre"`
	Ruta   string `json:"ruta"`
	Icono  string `json:"icono"`
	Color  string `json:"color"`
}

// UserInfo - Información del usuario autenticado
type UserInfo struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
	Rol    string `json:"rol"`
}

// ============================================
// RESPUESTAS DE ERROR
// ============================================

// ErrorResponse - Respuesta de error estándar
type ErrorResponse struct {
	Success   bool   `json:"success"`
	Code      int    `json:"code"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	ErrorID   string `json:"error_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

// ValidationError - Error de validación de campo
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorsResponse - Respuesta con múltiples errores de validación
type ValidationErrorsResponse struct {
	Success bool              `json:"success"`
	Code    int               `json:"code"`
	Errors  []ValidationError `json:"errors"`
}

// ============================================
// FUNCIONES AUXILIARES PARA CREAR RESPUESTAS
// ============================================

// NewSuccessResponse - Crea una respuesta de éxito
func NewSuccessResponse(message string, data interface{}) *Response {
	return &Response{
		Success: true,
		Code:    200,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse - Crea una respuesta de error
func NewErrorResponse(code int, message, errorDetail string) *Response {
	return &Response{
		Success: false,
		Code:    code,
		Message: message,
		Error:   errorDetail,
	}
}

// NewPaginatedResponse - Crea una respuesta paginada
func NewPaginatedResponse(data interface{}, total, page, pageSize, totalPages int, hasNext, hasPrev bool) *PaginatedResponse {
	return &PaginatedResponse{
		Success:    true,
		Code:       200,
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}
}

// NewValidationErrorResponse - Crea una respuesta de error de validación
func NewValidationErrorResponse(errors map[string]string) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		Success: false,
		Code:    422,
		Errors:  errors,
	}
}

// NewCreateResponse - Crea una respuesta de creación exitosa
func NewCreateResponse(message string, id int, data interface{}) *CreateResponse {
	return &CreateResponse{
		Success: true,
		Message: message,
		ID:      id,
		Data:    data,
	}
}

// NewUpdateResponse - Crea una respuesta de actualización exitosa
func NewUpdateResponse(message string, id int, data interface{}) *UpdateResponse {
	return &UpdateResponse{
		Success: true,
		Message: message,
		ID:      id,
		Data:    data,
	}
}

// NewDeleteResponse - Crea una respuesta de eliminación exitosa
func NewDeleteResponse(message string, id int) *DeleteResponse {
	return &DeleteResponse{
		Success: true,
		Message: message,
		ID:      id,
	}
}

// NewDetailResponse - Crea una respuesta con detalle
func NewDetailResponse(data interface{}) *DetailResponse {
	return &DetailResponse{
		Success: true,
		Data:    data,
	}
}

// NewLoginResponse - Crea una respuesta de login exitoso
func NewLoginResponse(message, redirect, userName, userEmail, userRole string) *LoginResponse {
	return &LoginResponse{
		Success:   true,
		Message:   message,
		Redirect:  redirect,
		UserName:  userName,
		UserEmail: userEmail,
		UserRole:  userRole,
	}
}

// NewLogoutResponse - Crea una respuesta de logout
func NewLogoutResponse(message, redirect string) *LogoutResponse {
	return &LogoutResponse{
		Success:  true,
		Message:  message,
		Redirect: redirect,
	}
}
