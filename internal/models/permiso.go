package models

import "time"

// Permiso - Representa un permiso individual para un perfil en un módulo
type Permiso struct {
	ID            int       `json:"id"`
	PerfilID      int       `json:"perfil_id"`
	ModuloID      int       `json:"modulo_id"`
	ModuloNombre  string    `json:"modulo_nombre"` // ← Agregado
	PuedeVer      bool      `json:"puede_ver"`
	PuedeCrear    bool      `json:"puede_crear"`
	PuedeEditar   bool      `json:"puede_editar"`
	PuedeEliminar bool      `json:"puede_eliminar"`
	PuedeDetalle  bool      `json:"puede_detalle"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PermisoRequest - Estructura para guardar múltiples permisos
type PermisoRequest struct {
	PerfilID int                  `json:"perfil_id" validate:"required"`
	Permisos []PermisoItemRequest `json:"permisos" validate:"required,dive"`
}

// PermisoItemRequest - Permiso individual en la petición
type PermisoItemRequest struct {
	ModuloID      int  `json:"modulo_id" validate:"required"`
	PuedeVer      bool `json:"puede_ver"`
	PuedeCrear    bool `json:"puede_crear"`
	PuedeEditar   bool `json:"puede_editar"`
	PuedeEliminar bool `json:"puede_eliminar"`
	PuedeDetalle  bool `json:"puede_detalle"`
}

// ModuloConPermisos - Módulo con sus permisos para un perfil específico
type ModuloConPermisos struct {
	Modulo        Modulo `json:"modulo"`
	PuedeVer      bool   `json:"puede_ver"`
	PuedeCrear    bool   `json:"puede_crear"`
	PuedeEditar   bool   `json:"puede_editar"`
	PuedeEliminar bool   `json:"puede_eliminar"`
	PuedeDetalle  bool   `json:"puede_detalle"`
}

// PermisosPorPerfilResponse - Respuesta para la pantalla de permisos
type PermisosPorPerfilResponse struct {
	Perfil   Perfil              `json:"perfil"`
	Permisos []ModuloConPermisos `json:"permisos"`
	Total    int                 `json:"total"`
}
