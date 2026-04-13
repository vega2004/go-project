package repository

import (
	"database/sql"
	"fmt"
	"tu-proyecto/internal/models"

	"github.com/lib/pq"
)

// ============================================
// INTERFAZ PRINCIPAL
// ============================================
type PermisoRepository interface {
	// Asignar/Actualizar permisos
	AssignPermissions(perfilID int, permisos []models.PermisoItemRequest) error

	// Obtener permisos de un perfil
	GetPermissionsByPerfil(perfilID int) ([]models.ModuloConPermisos, error)

	// Verificar permisos de un usuario
	UserHasPermission(userID int, moduloNombre string, permiso string) (bool, error)
	GetUserPermissions(userID int) (map[string]map[string]bool, error)
	IsAdmin(userID int) (bool, error)

	// Eliminar todos los permisos de un perfil
	DeletePermissionsByPerfil(perfilID int) error

	// Nuevos métodos
	GetUserPermissionsByUserID(userID int) ([]models.ModuloConPermisos, error)
	GetUserRole(userID int) (int, string, error)
	HasAnyPermission(userID int, moduloNombre string) (bool, error)
}

// ============================================
// ESTRUCTURA PRIVADA
// ============================================
type permisoRepository struct {
	db *sql.DB
}

// ============================================
// CONSTRUCTOR
// ============================================
func NewPermisoRepository(db *sql.DB) PermisoRepository {
	return &permisoRepository{db: db}
}

// ============================================
// ASSIGN PERMISSIONS
// ============================================
func (r *permisoRepository) AssignPermissions(perfilID int, permisos []models.PermisoItemRequest) error {
	if perfilID <= 0 {
		return fmt.Errorf("ID de perfil inválido: %d", perfilID)
	}
	if len(permisos) == 0 {
		return fmt.Errorf("no se recibieron permisos para asignar")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM permisos WHERE perfil_id = $1`, perfilID)
	if err != nil {
		return fmt.Errorf("error eliminando permisos antiguos: %w", err)
	}

	query := `
        INSERT INTO permisos (perfil_id, modulo_id, puede_ver, puede_crear, puede_editar, puede_eliminar, puede_detalle)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	for _, p := range permisos {
		_, err = tx.Exec(query, perfilID, p.ModuloID, p.PuedeVer, p.PuedeCrear,
			p.PuedeEditar, p.PuedeEliminar, p.PuedeDetalle)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				return fmt.Errorf("error insertando permisos para módulo %d: %s", p.ModuloID, pqErr.Message)
			}
			return fmt.Errorf("error insertando permisos: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error confirmando transacción: %w", err)
	}
	return nil
}

// ============================================
// GET PERMISSIONS BY PERFIL
// ============================================
func (r *permisoRepository) GetPermissionsByPerfil(perfilID int) ([]models.ModuloConPermisos, error) {
	if perfilID <= 0 {
		return nil, fmt.Errorf("ID de perfil inválido: %d", perfilID)
	}

	query := `
        SELECT 
            m.id, m.nombre, m.nombre_mostrar, m.ruta, m.icono, m.categoria, m.orden, m.activo, m.created_at,
            COALESCE(p.puede_ver, false) as puede_ver,
            COALESCE(p.puede_crear, false) as puede_crear,
            COALESCE(p.puede_editar, false) as puede_editar,
            COALESCE(p.puede_eliminar, false) as puede_eliminar,
            COALESCE(p.puede_detalle, false) as puede_detalle
        FROM modulos m
        LEFT JOIN permisos p ON p.modulo_id = m.id AND p.perfil_id = $1
        WHERE m.activo = true
        ORDER BY m.categoria, m.orden, m.id
    `

	rows, err := r.db.Query(query, perfilID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo permisos: %w", err)
	}
	defer rows.Close()

	var resultados []models.ModuloConPermisos
	for rows.Next() {
		var mp models.ModuloConPermisos
		err := rows.Scan(
			&mp.Modulo.ID, &mp.Modulo.Nombre, &mp.Modulo.NombreMostrar, &mp.Modulo.Ruta,
			&mp.Modulo.Icono, &mp.Modulo.Categoria, &mp.Modulo.Orden, &mp.Modulo.Activo, &mp.Modulo.CreatedAt,
			&mp.PuedeVer, &mp.PuedeCrear, &mp.PuedeEditar, &mp.PuedeEliminar, &mp.PuedeDetalle,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando resultado: %w", err)
		}
		resultados = append(resultados, mp)
	}
	return resultados, nil
}

// ============================================
// USER HAS PERMISSION
// ============================================
func (r *permisoRepository) UserHasPermission(userID int, moduloNombre string, permiso string) (bool, error) {
	if userID <= 0 {
		return false, fmt.Errorf("ID de usuario inválido: %d", userID)
	}
	if moduloNombre == "" {
		return false, fmt.Errorf("nombre de módulo requerido")
	}

	query := `
        SELECT 
            CASE $3
                WHEN 'ver' THEN p.puede_ver
                WHEN 'crear' THEN p.puede_crear
                WHEN 'editar' THEN p.puede_editar
                WHEN 'eliminar' THEN p.puede_eliminar
                WHEN 'detalle' THEN p.puede_detalle
                ELSE false
            END
        FROM users u
        JOIN perfiles per ON u.perfil_id = per.id
        JOIN permisos p ON p.perfil_id = per.id
        JOIN modulos m ON p.modulo_id = m.id
        WHERE u.id = $1 AND m.nombre = $2
    `

	var tienePermiso bool
	err := r.db.QueryRow(query, userID, moduloNombre, permiso).Scan(&tienePermiso)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("error verificando permiso: %w", err)
	}
	return tienePermiso, nil
}

// ============================================
// GET USER PERMISSIONS (para frontend)
// ============================================
func (r *permisoRepository) GetUserPermissions(userID int) (map[string]map[string]bool, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("ID de usuario inválido: %d", userID)
	}

	query := `
        SELECT 
            m.nombre,
            p.puede_ver,
            p.puede_crear,
            p.puede_editar,
            p.puede_eliminar,
            p.puede_detalle
        FROM users u
        JOIN perfiles per ON u.perfil_id = per.id
        JOIN permisos p ON p.perfil_id = per.id
        JOIN modulos m ON p.modulo_id = m.id
        WHERE u.id = $1 AND m.activo = true
    `

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo permisos del usuario: %w", err)
	}
	defer rows.Close()

	resultado := make(map[string]map[string]bool)
	for rows.Next() {
		var moduloNombre string
		var puedeVer, puedeCrear, puedeEditar, puedeEliminar, puedeDetalle bool

		err := rows.Scan(&moduloNombre, &puedeVer, &puedeCrear, &puedeEditar, &puedeEliminar, &puedeDetalle)
		if err != nil {
			return nil, fmt.Errorf("error escaneando permiso: %w", err)
		}

		resultado[moduloNombre] = map[string]bool{
			"ver":      puedeVer,
			"crear":    puedeCrear,
			"editar":   puedeEditar,
			"eliminar": puedeEliminar,
			"detalle":  puedeDetalle,
		}
	}
	return resultado, nil
}

// ============================================
// IS ADMIN
// ============================================
func (r *permisoRepository) IsAdmin(userID int) (bool, error) {
	if userID <= 0 {
		return false, fmt.Errorf("ID de usuario inválido: %d", userID)
	}

	query := `
        SELECT EXISTS(
            SELECT 1 FROM users u
            JOIN perfiles per ON u.perfil_id = per.id
            WHERE u.id = $1 AND per.nombre = 'Administrador'
        )
    `
	var isAdmin bool
	err := r.db.QueryRow(query, userID).Scan(&isAdmin)
	if err != nil {
		return false, fmt.Errorf("error verificando admin: %w", err)
	}
	return isAdmin, nil
}

// ============================================
// DELETE PERMISSIONS BY PERFIL
// ============================================
func (r *permisoRepository) DeletePermissionsByPerfil(perfilID int) error {
	if perfilID <= 0 {
		return fmt.Errorf("ID de perfil inválido: %d", perfilID)
	}

	query := `DELETE FROM permisos WHERE perfil_id = $1`
	_, err := r.db.Exec(query, perfilID)
	if err != nil {
		return fmt.Errorf("error eliminando permisos: %w", err)
	}
	return nil
}

// ============================================
// GET USER PERMISSIONS BY USER ID
// ============================================
func (r *permisoRepository) GetUserPermissionsByUserID(userID int) ([]models.ModuloConPermisos, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("ID de usuario inválido: %d", userID)
	}

	query := `
        SELECT 
            m.id, m.nombre, m.nombre_mostrar, m.ruta, m.icono, m.categoria, m.orden, m.activo, m.created_at,
            COALESCE(p.puede_ver, false) as puede_ver,
            COALESCE(p.puede_crear, false) as puede_crear,
            COALESCE(p.puede_editar, false) as puede_editar,
            COALESCE(p.puede_eliminar, false) as puede_eliminar,
            COALESCE(p.puede_detalle, false) as puede_detalle
        FROM users u
        JOIN perfiles per ON u.perfil_id = per.id
        CROSS JOIN modulos m
        LEFT JOIN permisos p ON p.modulo_id = m.id AND p.perfil_id = per.id
        WHERE u.id = $1 AND m.activo = true
        ORDER BY m.categoria, m.orden, m.id
    `

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo permisos del usuario: %w", err)
	}
	defer rows.Close()

	var resultados []models.ModuloConPermisos
	for rows.Next() {
		var mp models.ModuloConPermisos
		err := rows.Scan(
			&mp.Modulo.ID, &mp.Modulo.Nombre, &mp.Modulo.NombreMostrar, &mp.Modulo.Ruta,
			&mp.Modulo.Icono, &mp.Modulo.Categoria, &mp.Modulo.Orden, &mp.Modulo.Activo, &mp.Modulo.CreatedAt,
			&mp.PuedeVer, &mp.PuedeCrear, &mp.PuedeEditar, &mp.PuedeEliminar, &mp.PuedeDetalle,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando resultado: %w", err)
		}
		resultados = append(resultados, mp)
	}
	return resultados, nil
}

// ============================================
// GET USER ROLE
// ============================================
func (r *permisoRepository) GetUserRole(userID int) (int, string, error) {
	if userID <= 0 {
		return 0, "", fmt.Errorf("ID de usuario inválido: %d", userID)
	}

	var roleID int
	var roleName string
	query := `
        SELECT u.perfil_id, p.nombre
        FROM users u
        JOIN perfiles p ON u.perfil_id = p.id
        WHERE u.id = $1
    `
	err := r.db.QueryRow(query, userID).Scan(&roleID, &roleName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", fmt.Errorf("usuario no encontrado")
		}
		return 0, "", fmt.Errorf("error obteniendo rol: %w", err)
	}
	return roleID, roleName, nil
}

// ============================================
// HAS ANY PERMISSION
// ============================================
func (r *permisoRepository) HasAnyPermission(userID int, moduloNombre string) (bool, error) {
	if userID <= 0 {
		return false, fmt.Errorf("ID de usuario inválido: %d", userID)
	}
	if moduloNombre == "" {
		return false, fmt.Errorf("nombre de módulo requerido")
	}

	query := `
        SELECT EXISTS(
            SELECT 1 
            FROM users u
            JOIN perfiles per ON u.perfil_id = per.id
            JOIN permisos p ON p.perfil_id = per.id
            JOIN modulos m ON p.modulo_id = m.id
            WHERE u.id = $1 
                AND m.nombre = $2
                AND (p.puede_ver = true OR p.puede_crear = true OR p.puede_editar = true 
                     OR p.puede_eliminar = true OR p.puede_detalle = true)
        )
    `
	var hasPermission bool
	err := r.db.QueryRow(query, userID, moduloNombre).Scan(&hasPermission)
	if err != nil {
		return false, fmt.Errorf("error verificando permisos: %w", err)
	}
	return hasPermission, nil
}
