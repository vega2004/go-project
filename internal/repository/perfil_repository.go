package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"tu-proyecto/internal/models"

	"github.com/lib/pq"
)

// ============================================
// INTERFAZ PÚBLICA - COMPLETA
// ============================================
type PerfilRepository interface {
	// CRUD para perfiles/roles
	Create(perfil *models.Perfil) error
	Update(perfil *models.Perfil) error
	Delete(id int) error
	FindByID(id int) (*models.Perfil, error)
	FindAll(filter *models.PerfilFilter) (*models.PerfilPaginatedResponse, error)
	IsInUse(id int) (bool, error)

	// Métodos para perfil de usuario (tabla perfiles_usuario)
	GetPerfilByUserID(userID int) (*models.PerfilUsuario, error)
	UpdatePerfil(userID int, perfil *models.PerfilUsuario) error
	UpdateFoto(userID int, fotoPath, nombreOriginal, mimeType string, sizeBytes int) error
}

// ============================================
// ESTRUCTURA PRIVADA
// ============================================
type perfilRepository struct {
	db *sql.DB
}

// ============================================
// CONSTRUCTOR
// ============================================
func NewPerfilRepository(db *sql.DB) PerfilRepository {
	return &perfilRepository{db: db}
}

// ============================================
// CRUD PARA PERFILES/ROLES (tabla perfiles)
// ============================================

// Create - Crear un nuevo perfil/rol
// Create - Crear un nuevo perfil/rol
func (r *perfilRepository) Create(perfil *models.Perfil) error {
	log.Printf("[DEBUG] Repository.Create - Insertando perfil: Nombre='%s', Descripcion='%s'", perfil.Nombre, perfil.Descripcion)

	query := `INSERT INTO perfiles (nombre, descripcion) VALUES ($1, $2) RETURNING id, created_at`
	err := r.db.QueryRow(query, perfil.Nombre, perfil.Descripcion).Scan(&perfil.ID, &perfil.CreatedAt)

	if err != nil {
		log.Printf("[ERROR] Repository.Create - Error: %v", err)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			log.Printf("[DEBUG] Error de duplicado - ya existe un perfil con nombre '%s'", perfil.Nombre)
			return fmt.Errorf("ya existe un perfil con el nombre '%s'", perfil.Nombre)
		}
		return fmt.Errorf("error al crear perfil: %w", err)
	}

	log.Printf("[DEBUG] Repository.Create - Perfil creado exitosamente con ID: %d", perfil.ID)
	return nil
}

// Update - Actualizar un perfil/rol existente
func (r *perfilRepository) Update(perfil *models.Perfil) error {
	query := `UPDATE perfiles SET nombre=$1, descripcion=$2 WHERE id=$3`
	result, err := r.db.Exec(query, perfil.Nombre, perfil.Descripcion, perfil.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("ya existe un perfil con el nombre '%s'", perfil.Nombre)
		}
		return fmt.Errorf("error al actualizar perfil: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("perfil con ID %d no encontrado", perfil.ID)
	}
	return nil
}

// Delete - Eliminar un perfil/rol
func (r *perfilRepository) Delete(id int) error {
	query := `DELETE FROM perfiles WHERE id=$1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar perfil: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("perfil con ID %d no encontrado", id)
	}
	return nil
}

// FindByID - Buscar perfil/rol por ID
func (r *perfilRepository) FindByID(id int) (*models.Perfil, error) {
	perfil := &models.Perfil{}
	query := `SELECT id, nombre, descripcion, created_at FROM perfiles WHERE id=$1`
	err := r.db.QueryRow(query, id).Scan(&perfil.ID, &perfil.Nombre, &perfil.Descripcion, &perfil.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("perfil con ID %d no encontrado", id)
		}
		return nil, fmt.Errorf("error al buscar perfil: %w", err)
	}
	return perfil, nil
}

// FindAll - Listar perfiles/roles con paginación
func (r *perfilRepository) FindAll(filter *models.PerfilFilter) (*models.PerfilPaginatedResponse, error) {
	if filter == nil {
		filter = &models.PerfilFilter{}
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	offset := (filter.Page - 1) * filter.PageSize

	where := []string{}
	args := []interface{}{}
	argPos := 1

	if filter.Nombre != "" {
		where = append(where, fmt.Sprintf("nombre ILIKE $%d", argPos))
		args = append(args, "%"+filter.Nombre+"%")
		argPos++
	}

	countQuery := "SELECT COUNT(*) FROM perfiles"
	if len(where) > 0 {
		countQuery += " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("error al contar perfiles: %w", err)
	}

	dataQuery := "SELECT id, nombre, descripcion, created_at FROM perfiles"
	if len(where) > 0 {
		dataQuery += " WHERE " + strings.Join(where, " AND ")
	}
	dataQuery += fmt.Sprintf(" ORDER BY nombre ASC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("error al listar perfiles: %w", err)
	}
	defer rows.Close()

	var perfiles []models.Perfil
	for rows.Next() {
		var p models.Perfil
		err := rows.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error al escanear perfil: %w", err)
		}
		perfiles = append(perfiles, p)
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize

	return &models.PerfilPaginatedResponse{
		Data:       perfiles,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// IsInUse - Verifica si un perfil/rol tiene usuarios asignados
func (r *perfilRepository) IsInUse(id int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE perfil_id = $1`
	err := r.db.QueryRow(query, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error al verificar uso del perfil: %w", err)
	}
	return count > 0, nil
}

// ============================================
// MÉTODOS PARA PERFIL DE USUARIO (tabla perfiles_usuario)
// ============================================

// GetPerfilByUserID - Obtiene el perfil de un usuario (foto, bio, dirección)
// GetPerfilByUserID - Obtiene el perfil de un usuario (foto, bio, dirección)
func (r *perfilRepository) GetPerfilByUserID(userID int) (*models.PerfilUsuario, error) {
	perfil := &models.PerfilUsuario{}
	query := `
        SELECT 
            user_id, 
            COALESCE(foto_path, '/static/uploads/perfil/default-avatar.png') as foto_path,
            COALESCE(foto_nombre_original, '') as foto_nombre_original,
            COALESCE(foto_mime_type, '') as foto_mime_type,
            COALESCE(foto_size_bytes, 0) as foto_size_bytes,
            COALESCE(bio, '') as bio,
            COALESCE(direccion, '') as direccion,
            COALESCE(telefono_alterno, '') as telefono_alterno,
            updated_at
        FROM perfiles_usuario 
        WHERE user_id = $1
    `
	err := r.db.QueryRow(query, userID).Scan(
		&perfil.UserID, &perfil.FotoPath, &perfil.FotoNombreOriginal, &perfil.FotoMimeType, &perfil.FotoSizeBytes,
		&perfil.Bio, &perfil.Direccion, &perfil.TelefonoAlterno, &perfil.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Crear perfil por defecto si no existe
			return r.createDefaultPerfil(userID)
		}
		return nil, fmt.Errorf("error al obtener perfil de usuario: %w", err)
	}
	return perfil, nil
}

// createDefaultPerfil - Crea un perfil de usuario por defecto
func (r *perfilRepository) createDefaultPerfil(userID int) (*models.PerfilUsuario, error) {
	query := `
		INSERT INTO perfiles_usuario (user_id, foto_path, updated_at) 
		VALUES ($1, '/static/uploads/perfil/default-avatar.png', NOW())
		RETURNING user_id, foto_path, updated_at
	`
	perfil := &models.PerfilUsuario{
		UserID:          userID,
		FotoPath:        "/static/uploads/perfil/default-avatar.png",
		Bio:             "",
		Direccion:       "",
		TelefonoAlterno: "",
	}
	err := r.db.QueryRow(query, userID).Scan(&perfil.UserID, &perfil.FotoPath, &perfil.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error al crear perfil por defecto: %w", err)
	}
	return perfil, nil
}

// UpdatePerfil - Actualiza los datos del perfil de usuario (bio, dirección, teléfono alterno)
func (r *perfilRepository) UpdatePerfil(userID int, perfil *models.PerfilUsuario) error {
	query := `
		UPDATE perfiles_usuario 
		SET bio = $1, direccion = $2, telefono_alterno = $3, updated_at = NOW()
		WHERE user_id = $4
	`
	result, err := r.db.Exec(query, perfil.Bio, perfil.Direccion, perfil.TelefonoAlterno, userID)
	if err != nil {
		return fmt.Errorf("error al actualizar perfil de usuario: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Si no existe, crearlo
		_, err = r.db.Exec(`
			INSERT INTO perfiles_usuario (user_id, bio, direccion, telefono_alterno, foto_path, updated_at)
			VALUES ($1, $2, $3, $4, '/static/uploads/perfil/default-avatar.png', NOW())
		`, userID, perfil.Bio, perfil.Direccion, perfil.TelefonoAlterno)
		if err != nil {
			return fmt.Errorf("error al crear perfil de usuario: %w", err)
		}
	}
	return nil
}

// UpdateFoto - Actualiza la foto de perfil del usuario
func (r *perfilRepository) UpdateFoto(userID int, fotoPath, nombreOriginal, mimeType string, sizeBytes int) error {
	query := `
		UPDATE perfiles_usuario 
		SET foto_path = $1, foto_nombre_original = $2, foto_mime_type = $3, foto_size_bytes = $4, updated_at = NOW()
		WHERE user_id = $5
	`
	result, err := r.db.Exec(query, fotoPath, nombreOriginal, mimeType, sizeBytes, userID)
	if err != nil {
		return fmt.Errorf("error al actualizar foto de perfil: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Si no existe, crear perfil con la foto
		_, err = r.db.Exec(`
			INSERT INTO perfiles_usuario (user_id, foto_path, foto_nombre_original, foto_mime_type, foto_size_bytes, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
		`, userID, fotoPath, nombreOriginal, mimeType, sizeBytes)
		if err != nil {
			return fmt.Errorf("error al crear perfil con foto: %w", err)
		}
	}
	return nil
}
