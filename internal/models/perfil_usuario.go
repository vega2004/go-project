package models

import "time"

// PerfilUsuario - Datos extras del perfil de usuario (tabla perfiles_usuario)
type PerfilUsuario struct {
	UserID             int       `json:"user_id"`
	FotoPath           string    `json:"foto_path"`
	FotoNombreOriginal string    `json:"foto_nombre_original,omitempty"`
	FotoMimeType       string    `json:"foto_mime_type,omitempty"`
	FotoSizeBytes      int       `json:"foto_size_bytes,omitempty"`
	Bio                string    `json:"bio"`
	Direccion          string    `json:"direccion"`
	TelefonoAlterno    string    `json:"telefono_alterno"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// PerfilConUsuario - Perfil de usuario con datos del usuario
type PerfilConUsuario struct {
	PerfilUsuario
	User
	RolNombre string `json:"rol_nombre"`
}

// CambioPassword - Estructura para cambio de contraseña
type CambioPassword struct {
	Actual    string `json:"actual" validate:"required"`
	Nueva     string `json:"nueva" validate:"required,min=6"`
	Confirmar string `json:"confirmar" validate:"required,eqfield=Nueva"`
}
