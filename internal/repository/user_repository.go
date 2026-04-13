package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"tu-proyecto/internal/models"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// ============================================
// INTERFAZ PÚBLICA
// ============================================
type UserRepository interface {
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id int) error // ← auditorID no es necesario aquí
	FindByID(id int) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindAll(filter *models.UserFilter) (*models.UserPaginatedResponse, error)
	UpdatePassword(id int, newPassword string) error
	UpdateStatus(id int, activo bool) error
	IsEmailUnique(email string, excludeID int) (bool, error)

	// Nuevos métodos
	GetUserWithProfile(id int) (*models.UserWithProfile, error) // ← Usa models
	CountActiveAdmins() (int, error)
	ExistsByID(id int) (bool, error)
}

// ============================================
// ESTRUCTURA PRIVADA
// ============================================
type userRepository struct {
	db *sql.DB
}

// ============================================
// CONSTRUCTOR
// ============================================
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// ============================================
// CREATE
// ============================================
func (r *userRepository) Create(user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hasheando contraseña: %w", err)
	}

	query := `
        INSERT INTO users (name, email, phone, password, perfil_id, activo, created_at, updated_at) 
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) 
        RETURNING id, created_at, updated_at
    `

	err = r.db.QueryRow(query, user.Name, user.Email, user.Phone, string(hashedPassword),
		user.PerfilID, user.Activo).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("ya existe un usuario con el email '%s'", user.Email)
		}
		return fmt.Errorf("error creando usuario: %w", err)
	}

	// Crear perfil de usuario asociado
	_, err = r.db.Exec(`
        INSERT INTO perfiles_usuario (user_id, foto_path, updated_at) 
        VALUES ($1, '/static/uploads/perfil/default-avatar.png', NOW())
    `, user.ID)
	if err != nil {
		// No fallamos por esto, solo logueamos
		fmt.Printf("Warning: no se pudo crear perfil_usuario para user_id %d: %v\n", user.ID, err)
	}

	return nil
}

// ============================================
// UPDATE
// ============================================
func (r *userRepository) Update(user *models.User) error {
	exists, err := r.ExistsByID(user.ID)
	if err != nil {
		return fmt.Errorf("error verificando existencia: %w", err)
	}
	if !exists {
		return fmt.Errorf("usuario con ID %d no encontrado", user.ID)
	}

	query := `
        UPDATE users 
        SET name = $1, email = $2, phone = $3, perfil_id = $4, activo = $5, updated_at = NOW()
        WHERE id = $6
    `
	result, err := r.db.Exec(query, user.Name, user.Email, user.Phone, user.PerfilID, user.Activo, user.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("ya existe un usuario con el email '%s'", user.Email)
		}
		return fmt.Errorf("error actualizando usuario: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("usuario con ID %d no encontrado", user.ID)
	}
	return nil
}

// ============================================
// DELETE (CORREGIDO - sin auditorID innecesario)
// ============================================
func (r *userRepository) Delete(id int) error {
	user, err := r.FindByID(id)
	if err != nil {
		return fmt.Errorf("usuario no encontrado: %w", err)
	}

	// Verificar si es administrador
	if user.PerfilID == 1 {
		adminCount, err := r.CountActiveAdmins()
		if err != nil {
			return fmt.Errorf("error verificando administradores: %w", err)
		}
		if adminCount <= 1 {
			return fmt.Errorf("no se puede eliminar el último administrador del sistema")
		}
	}

	// Primero eliminar perfil_usuario
	_, err = r.db.Exec(`DELETE FROM perfiles_usuario WHERE user_id = $1`, id)
	if err != nil {
		return fmt.Errorf("error eliminando perfil de usuario: %w", err)
	}

	// Luego eliminar usuario
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error eliminando usuario: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("usuario con ID %d no encontrado", id)
	}
	return nil
}

// ============================================
// FIND BY ID
// ============================================
func (r *userRepository) FindByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `
        SELECT id, name, email, phone, perfil_id, activo, created_at, updated_at
        FROM users WHERE id = $1
    `
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone,
		&user.PerfilID, &user.Activo, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario con ID %d no encontrado", id)
		}
		return nil, fmt.Errorf("error buscando usuario: %w", err)
	}
	return user, nil
}

// ============================================
// FIND BY EMAIL
// ============================================
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
        SELECT id, name, email, phone, perfil_id, activo, created_at, updated_at
        FROM users WHERE email = $1
    `
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone,
		&user.PerfilID, &user.Activo, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario con email '%s' no encontrado", email)
		}
		return nil, fmt.Errorf("error buscando usuario: %w", err)
	}
	return user, nil
}

// ============================================
// UPDATE PASSWORD
// ============================================
func (r *userRepository) UpdatePassword(id int, newPassword string) error {
	exists, err := r.ExistsByID(id)
	if err != nil {
		return fmt.Errorf("error verificando existencia: %w", err)
	}
	if !exists {
		return fmt.Errorf("usuario con ID %d no encontrado", id)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hasheando contraseña: %w", err)
	}

	query := `UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.Exec(query, string(hashedPassword), id)
	if err != nil {
		return fmt.Errorf("error actualizando contraseña: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("usuario con ID %d no encontrado", id)
	}
	return nil
}

// ============================================
// UPDATE STATUS
// ============================================
func (r *userRepository) UpdateStatus(id int, activo bool) error {
	exists, err := r.ExistsByID(id)
	if err != nil {
		return fmt.Errorf("error verificando existencia: %w", err)
	}
	if !exists {
		return fmt.Errorf("usuario con ID %d no encontrado", id)
	}

	query := `UPDATE users SET activo = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.Exec(query, activo, id)
	if err != nil {
		return fmt.Errorf("error actualizando estado: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("usuario con ID %d no encontrado", id)
	}
	return nil
}

// ============================================
// IS EMAIL UNIQUE
// ============================================
func (r *userRepository) IsEmailUnique(email string, excludeID int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1 AND id != $2`
	err := r.db.QueryRow(query, email, excludeID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error verificando email único: %w", err)
	}
	return count == 0, nil
}

// ============================================
// COUNT ACTIVE ADMINS
// ============================================
func (r *userRepository) CountActiveAdmins() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE perfil_id = 1 AND activo = true`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error contando administradores: %w", err)
	}
	return count, nil
}

// ============================================
// EXISTS BY ID
// ============================================
func (r *userRepository) ExistsByID(id int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := r.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error verificando existencia: %w", err)
	}
	return exists, nil
}

// ============================================
// GET USER WITH PROFILE
// ============================================
func (r *userRepository) GetUserWithProfile(id int) (*models.UserWithProfile, error) {
	query := `
        SELECT 
            u.id, u.name, u.email, u.phone, u.perfil_id, u.activo, u.created_at, u.updated_at,
            p.nombre as perfil_nombre,
            COALESCE(pu.foto_path, '/static/uploads/perfil/default-avatar.png') as foto_path,
            COALESCE(pu.bio, '') as bio,
            COALESCE(pu.direccion, '') as direccion,
            COALESCE(pu.telefono_alterno, '') as telefono_alterno
        FROM users u
        JOIN perfiles p ON u.perfil_id = p.id
        LEFT JOIN perfiles_usuario pu ON u.id = pu.user_id
        WHERE u.id = $1
    `

	var result models.UserWithProfile
	err := r.db.QueryRow(query, id).Scan(
		&result.ID, &result.Name, &result.Email, &result.Phone,
		&result.PerfilID, &result.Activo, &result.CreatedAt, &result.UpdatedAt,
		&result.PerfilNombre,
		&result.FotoPath, &result.Bio, &result.Direccion, &result.TelefonoAlterno,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario con ID %d no encontrado", id)
		}
		return nil, fmt.Errorf("error obteniendo usuario con perfil: %w", err)
	}
	return &result, nil
}

// ============================================
// FIND ALL - Listar usuarios con paginación y filtros
// ============================================
func (r *userRepository) FindAll(filter *models.UserFilter) (*models.UserPaginatedResponse, error) {
	if filter == nil {
		filter = &models.UserFilter{}
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

	if filter.Name != "" {
		where = append(where, fmt.Sprintf("u.name ILIKE $%d", argPos))
		args = append(args, "%"+filter.Name+"%")
		argPos++
	}

	if filter.Email != "" {
		where = append(where, fmt.Sprintf("u.email ILIKE $%d", argPos))
		args = append(args, "%"+filter.Email+"%")
		argPos++
	}

	if filter.PerfilID > 0 {
		where = append(where, fmt.Sprintf("u.perfil_id = $%d", argPos))
		args = append(args, filter.PerfilID)
		argPos++
	}

	if filter.Activo != nil {
		where = append(where, fmt.Sprintf("u.activo = $%d", argPos))
		args = append(args, *filter.Activo)
		argPos++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	countQuery := `SELECT COUNT(*) FROM users u` + whereClause
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("error contando usuarios: %w", err)
	}

	dataQuery := `
        SELECT 
            u.id, u.name, u.email, u.phone, u.perfil_id, u.activo, u.created_at,
            p.nombre as perfil_nombre,
            COALESCE(pu.foto_path, '/static/uploads/perfil/default-avatar.png') as foto_path
        FROM users u
        JOIN perfiles p ON u.perfil_id = p.id
        LEFT JOIN perfiles_usuario pu ON u.id = pu.user_id
    ` + whereClause + `
        ORDER BY u.id DESC
        LIMIT $%d OFFSET $%d
    `
	dataQuery = fmt.Sprintf(dataQuery, argPos, argPos+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("error listando usuarios: %w", err)
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var u models.UserResponse
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.PerfilID, &u.Activo, &u.CreatedAt, &u.PerfilNombre, &u.FotoPath)
		if err != nil {
			return nil, fmt.Errorf("error escaneando usuario: %w", err)
		}
		users = append(users, u)
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize

	return &models.UserPaginatedResponse{
		Data:       users,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
		HasNext:    filter.Page < totalPages,
		HasPrev:    filter.Page > 1,
	}, nil
}
