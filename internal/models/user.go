package models

import "time"

// User - Usuario del sistema
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Password  string    `json:"-"` // No se incluye en JSON
	PerfilID  int       `json:"perfil_id"`
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserResponse - Respuesta para API (sin contraseña)
type UserResponse struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	PerfilID     int       `json:"perfil_id"`
	PerfilNombre string    `json:"perfil_nombre"`
	FotoPath     string    `json:"foto_path"` // ← Añadido campo foto
	Activo       bool      `json:"activo"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserWithProfile - Usuario con datos completos de perfil (CORREGIDO)
type UserWithProfile struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	PerfilID        int       `json:"perfil_id"`
	PerfilNombre    string    `json:"perfil_nombre"`
	Activo          bool      `json:"activo"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	FotoPath        string    `json:"foto_path"`
	Bio             string    `json:"bio"`
	Direccion       string    `json:"direccion"`
	TelefonoAlterno string    `json:"telefono_alterno"`
}

// UserFilter - Filtros para listar usuarios
type UserFilter struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	PerfilID int    `json:"perfil_id"`
	Activo   *bool  `json:"activo"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// UserPaginatedResponse - Respuesta paginada
type UserPaginatedResponse struct {
	Data       []UserResponse `json:"data"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
	HasNext    bool           `json:"has_next"`
	HasPrev    bool           `json:"has_prev"`
}

// UserCreateRequest - Creación de usuario
type UserCreateRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password" validate:"required,min=6"`
	PerfilID int    `json:"perfil_id" validate:"required,min=1"`
	Activo   bool   `json:"activo"`
}

// UserUpdateRequest - Actualización de usuario
type UserUpdateRequest struct {
	ID       int    `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password"` // Opcional, solo si se quiere cambiar
	PerfilID int    `json:"perfil_id" validate:"required,min=1"`
	Activo   bool   `json:"activo"`
}
