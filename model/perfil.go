package model

import "time"

type Perfil struct {
	UserID    int       `json:"user_id"`
	Foto      string    `json:"foto"`
	Bio       string    `json:"bio"`
	Ubicacion string    `json:"ubicacion"`
	SitioWeb  string    `json:"sitio_web"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CambioPassword struct {
	Actual    string `json:"actual" validate:"required"`
	Nueva     string `json:"nueva" validate:"required,min=6"`
	Confirmar string `json:"confirmar" validate:"required,eqfield=Nueva"`
}
