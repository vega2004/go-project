// internal/models/perfil.go
package models

import "time"

type Perfil struct {
	ID          int       `json:"id"`
	Nombre      string    `json:"nombre"`
	Descripcion string    `json:"descripcion"`
	CreatedAt   time.Time `json:"created_at"`
}

type PerfilFilter struct {
	Nombre   string `json:"nombre"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type PerfilPaginatedResponse struct {
	Data       []Perfil `json:"data"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}
