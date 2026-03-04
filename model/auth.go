package model

import "time"

type UserAuth struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Password  string    `json:"-"`
	RoleID    int       `json:"role_id"` // ← NUEVO
	CreatedAt time.Time `json:"created_at"`
}

type LoginForm struct {
	Email          string `form:"email" validate:"required,email"`
	Password       string `form:"password" validate:"required,min=6"`
	RecaptchaToken string `form:"recaptchaToken" validate:"required"`
}

// NUEVO: RegisterForm con campo Password
type RegisterForm struct {
	Name           string `form:"name" validate:"required,min=2,max=50"`
	Email          string `form:"email" validate:"required,email"`
	Phone          string `form:"phone" validate:"required,phone_format"`
	Password       string `form:"password" validate:"required,min=6"`
	RecaptchaToken string `form:"recaptchaToken" validate:"required"`
}

type Session struct {
	UserID       int
	Email        string
	Name         string
	RoleID       int    // ← NUEVO
	RoleNombre   string // ← NUEVO
	LastActivity time.Time
}
