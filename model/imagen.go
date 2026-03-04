package model

import "time"

type Imagen struct {
	ID        int       `json:"id"`
	Nombre    string    `json:"nombre"`
	Ruta      string    `json:"ruta"`
	Orden     int       `json:"orden"`
	Activo    bool      `json:"activo"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CarruselResponse struct {
	Imagenes []Imagen `json:"imagenes"`
	Total    int      `json:"total"`
}
