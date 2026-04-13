package models

import "time"

type Modulo struct {
	ID            int       `json:"id"`
	Nombre        string    `json:"nombre"`
	NombreMostrar string    `json:"nombre_mostrar"`
	Ruta          string    `json:"ruta"`
	Icono         string    `json:"icono"`
	Categoria     string    `json:"categoria"`
	Orden         int       `json:"orden"`
	Activo        bool      `json:"activo"`
	CreatedAt     time.Time `json:"created_at"`
}

type ModuloFilter struct {
	Nombre   string `json:"nombre"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type ModuloPaginatedResponse struct {
	Data       []Modulo `json:"data"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}
