package repository

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
	"tu-proyecto/model"

	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	// Métodos básicos
	CreateUser(user *model.UserAuth) error
	FindByEmail(email string) (*model.UserAuth, error)
	FindByID(id int) (*model.UserAuth, error)
	ValidatePassword(hashedPassword, password string) bool

	// Métodos de actualización
	UpdateUser(user *model.UserAuth) error
	UpdateUserRole(userID, roleID int) error
	UpdatePassword(userID int, newPassword string) error
	DeleteUser(id int) error

	// Método para roles
	GetRolByID(roleID int) (string, error)

	// NUEVO: Método para listar usuarios con paginación
	GetAllUsers(filter *model.UserFilter) (*model.UserPaginatedResponse, error)
}

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

// ============================================
// MÉTODOS DE CREACIÓN
// ============================================

// CreateUser - Inserta un nuevo usuario
func (r *authRepository) CreateUser(user *model.UserAuth) error {
	query := `INSERT INTO users (name, email, phone, password, role_id, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = r.db.QueryRow(query,
		user.Name,
		user.Email,
		user.Phone,
		string(hashedPassword),
		user.RoleID,
		user.CreatedAt).Scan(&user.ID)

	return err
}

// ============================================
// MÉTODOS DE BÚSQUEDA
// ============================================

// FindByEmail - Busca usuario por email
func (r *authRepository) FindByEmail(email string) (*model.UserAuth, error) {
	user := &model.UserAuth{}
	query := `SELECT id, name, email, phone, password, role_id, created_at 
	          FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.RoleID,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// FindByID - Busca usuario por ID
func (r *authRepository) FindByID(id int) (*model.UserAuth, error) {
	user := &model.UserAuth{}
	query := `SELECT id, name, email, phone, password, role_id, created_at 
	          FROM users WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.RoleID,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// ============================================
// NUEVO: MÉTODO DE LISTADO CON PAGINACIÓN
// ============================================

// GetAllUsers - Obtiene todos los usuarios con paginación y filtros
// GetAllUsers - Obtiene todos los usuarios con paginación y filtros
func (r *authRepository) GetAllUsers(filter *model.UserFilter) (*model.UserPaginatedResponse, error) {
	// Configurar paginación
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 5
	}
	offset := (filter.Page - 1) * filter.PageSize

	// Construir query con filtros
	where := []string{}
	args := []interface{}{}
	argPos := 1

	if filter.Email != "" {
		where = append(where, "email ILIKE $"+strconv.Itoa(argPos))
		args = append(args, "%"+filter.Email+"%")
		argPos++
	}

	if filter.Name != "" {
		where = append(where, "name ILIKE $"+strconv.Itoa(argPos))
		args = append(args, "%"+filter.Name+"%")
		argPos++
	}

	// Query para contar total
	countQuery := "SELECT COUNT(*) FROM users"
	if len(where) > 0 {
		countQuery += " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Query para datos paginados
	dataQuery := "SELECT id, name, email, phone, role_id, created_at FROM users"
	if len(where) > 0 {
		dataQuery += " WHERE " + strings.Join(where, " AND ")
	}
	dataQuery += " ORDER BY id DESC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.UserAdminResponse
	for rows.Next() {
		var u model.UserAdminResponse
		var createdAt time.Time
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.RoleID, &createdAt)
		if err != nil {
			return nil, err
		}
		u.CreatedAt = createdAt.Format("2006-01-02 15:04:05")

		// Obtener nombre del rol
		roleName, _ := r.GetRolByID(u.RoleID)
		u.RoleName = roleName

		users = append(users, u)
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize

	return &model.UserPaginatedResponse{
		Data:       users,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
		HasNext:    filter.Page < totalPages,
		HasPrev:    filter.Page > 1,
	}, nil
}

// ============================================
// MÉTODOS DE ACTUALIZACIÓN
// ============================================

// UpdateUser - Actualiza datos de usuario (sin contraseña)
func (r *authRepository) UpdateUser(user *model.UserAuth) error {
	query := `UPDATE users SET name=$1, email=$2, phone=$3, role_id=$4 WHERE id=$5`

	result, err := r.db.Exec(query,
		user.Name,
		user.Email,
		user.Phone,
		user.RoleID,
		user.ID)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateUserRole - Actualiza el rol de un usuario
func (r *authRepository) UpdateUserRole(userID, roleID int) error {
	query := `UPDATE users SET role_id = $1 WHERE id = $2`

	result, err := r.db.Exec(query, roleID, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UpdatePassword - Actualiza la contraseña de un usuario
func (r *authRepository) UpdatePassword(userID int, newPassword string) error {
	// Hashear la nueva contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `UPDATE users SET password = $1 WHERE id = $2`

	result, err := r.db.Exec(query, string(hashedPassword), userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ============================================
// MÉTODOS DE ELIMINACIÓN
// ============================================

// DeleteUser - Elimina un usuario
func (r *authRepository) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id=$1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ============================================
// MÉTODOS DE ROLES
// ============================================

// GetRolByID - Obtiene el nombre del rol por su ID
func (r *authRepository) GetRolByID(roleID int) (string, error) {
	var nombre string
	query := `SELECT nombre FROM roles WHERE id = $1`
	err := r.db.QueryRow(query, roleID).Scan(&nombre)
	return nombre, err
}

// ============================================
// MÉTODOS DE VALIDACIÓN
// ============================================

// ValidatePassword - Compara contraseña con hash
func (r *authRepository) ValidatePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
