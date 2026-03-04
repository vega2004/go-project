package model

import "time"

type Persona struct {
	ID          int       `json:"id"`
	Nombre      string    `json:"nombre" validate:"required,min=2,max=100"`
	EstadoCivil string    `json:"estado_civil" validate:"required,oneof=soltero casado divorciado viudo"`
	UserID      int       `json:"user_id"` // Usuario que creó el registro
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PersonaFilter struct {
	Nombre      string `json:"nombre"`
	EstadoCivil string `json:"estado_civil"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"` // Default 5
}

type PaginatedResponse struct {
	Data       []Persona `json:"data"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
	HasNext    bool      `json:"has_next"`
	HasPrev    bool      `json:"has_prev"`
}
