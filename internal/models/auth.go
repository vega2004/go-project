package models

import "time"

// UserAuth - Usuario para autenticación
type UserAuth struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Password  string    `json:"-"`         // No se incluye en JSON
	PerfilID  int       `json:"perfil_id"` // Perfil del usuario
	Activo    bool      `json:"activo"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginForm - Formulario de login
type LoginForm struct {
	Email          string `form:"email" validate:"required,email"`
	Password       string `form:"password" validate:"required"`
	RecaptchaToken string `form:"recaptchaToken" validate:"required"`
}

// RegisterForm - Formulario de registro
type RegisterForm struct {
	Name            string `form:"name" validate:"required,min=2,max=50"`
	Email           string `form:"email" validate:"required,email"`
	Phone           string `form:"phone" validate:"required"`
	Password        string `form:"password" validate:"required,min=6"`
	ConfirmPassword string `form:"confirm_password" validate:"required,eqfield=Password"`
	RecaptchaToken  string `form:"recaptchaToken" validate:"required"`
}

// Session - Estructura de sesión
type Session struct {
	UserID       int                `json:"user_id"`
	Email        string             `json:"email"`
	Name         string             `json:"name"`
	PerfilID     int                `json:"perfil_id"`
	PerfilNombre string             `json:"perfil_nombre"`
	LastActivity time.Time          `json:"last_activity"`
	Permisos     map[string]Permiso `json:"permisos"` // Permisos del usuario (desde permiso.go)
}
