package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type UserForm struct {
	Name           string `form:"name" validate:"required,min=2,max=50,no_special_chars"`
	Email          string `form:"email" validate:"required,email"`
	Phone          string `form:"phone" validate:"required,phone_format"`
	RecaptchaToken string `form:"g-recaptcha-response" validate:"required"`
}
