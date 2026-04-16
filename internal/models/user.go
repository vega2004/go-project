package models

import "time"

// ============================================
// USUARIO PRINCIPAL
// ============================================

// User - Usuario del sistema
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Password  string    `json:"-"` // No se incluye en JSON por seguridad
	PerfilID  int       `json:"perfil_id"`
	FotoPath  string    `json:"foto_path"` // Ruta de la foto de perfil
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================
// RESPUESTAS PARA API (sin datos sensibles)
// ============================================

// UserResponse - Respuesta para listados (sin contraseña)
type UserResponse struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	PerfilID     int       `json:"perfil_id"`
	PerfilNombre string    `json:"perfil_nombre"`
	FotoPath     string    `json:"foto_path"`
	Activo       bool      `json:"activo"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserDetailResponse - Respuesta para detalle (con más datos)
type UserDetailResponse struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	PerfilID     int       `json:"perfil_id"`
	PerfilNombre string    `json:"perfil_nombre"`
	FotoPath     string    `json:"foto_path"`
	Activo       bool      `json:"activo"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ============================================
// SOLICITUDES (CREATE / UPDATE)
// ============================================

// UserCreateRequest - Creación de nuevo usuario
type UserCreateRequest struct {
	Name            string `json:"name" form:"name" validate:"required,min=2,max=100"`
	Email           string `json:"email" form:"email" validate:"required,email"`
	Phone           string `json:"phone" form:"phone"`
	Password        string `json:"password" form:"password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password" validate:"required,eqfield=Password"`
	PerfilID        int    `json:"perfil_id" form:"perfil_id" validate:"required,min=1"`
	Activo          bool   `json:"activo" form:"activo"`
}

// UserUpdateRequest - Actualización de usuario existente
type UserUpdateRequest struct {
	ID       int    `json:"id" form:"id" validate:"required"`
	Name     string `json:"name" form:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" form:"email" validate:"required,email"`
	Phone    string `json:"phone" form:"phone"`
	Password string `json:"password" form:"password"` // Opcional, solo si se quiere cambiar
	PerfilID int    `json:"perfil_id" form:"perfil_id" validate:"required,min=1"`
	Activo   bool   `json:"activo" form:"activo"`
}

// UserChangePasswordRequest - Cambio de contraseña
type UserChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" form:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" form:"new_password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password" validate:"required,eqfield=NewPassword"`
}

// ============================================
// FILTROS Y PAGINACIÓN
// ============================================

// UserFilter - Filtros para listar usuarios
type UserFilter struct {
	Name     string `json:"name" form:"name"`
	Email    string `json:"email" form:"email"`
	PerfilID int    `json:"perfil_id" form:"perfil_id"`
	Activo   string `json:"activo" form:"activo"` // "true", "false", o vacío = todos
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
}

// UserPaginatedResponse - Respuesta paginada para listados
type UserPaginatedResponse struct {
	Data       []UserResponse `json:"data"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
	HasNext    bool           `json:"has_next"`
	HasPrev    bool           `json:"has_prev"`
}

// ============================================
// USUARIO CON PERFIL COMPLETO
// ============================================

// UserWithProfile - Usuario con datos completos del perfil
type UserWithProfile struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	PerfilID        int       `json:"perfil_id"`
	PerfilNombre    string    `json:"perfil_nombre"`
	FotoPath        string    `json:"foto_path"`
	Activo          bool      `json:"activo"`
	Bio             string    `json:"bio"`
	Direccion       string    `json:"direccion"`
	TelefonoAlterno string    `json:"telefono_alterno"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ============================================
// FUNCIONES AUXILIARES
// ============================================

// IsActive - Retorna true si el usuario está activo
func (u *User) IsActive() bool {
	return u.Activo
}

// GetFotoPath - Retorna la ruta de la foto o la default
func (u *User) GetFotoPath() string {
	if u.FotoPath == "" {
		return "/static/uploads/perfil/default-avatar.png"
	}
	return u.FotoPath
}
