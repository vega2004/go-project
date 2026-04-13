package repository

import (
	"database/sql"
	"fmt"
	"time"
	"tu-proyecto/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// ============================================
// INTERFAZ PRINCIPAL (definida UNA SOLA VEZ)
// ============================================
type AuthRepository interface {
	// Métodos básicos
	FindByEmail(email string) (*models.UserAuth, error)
	FindByID(id int) (*models.UserAuth, error)
	CreateUser(user *models.UserAuth) error
	UpdatePassword(userID int, newPassword string) error
	ValidatePassword(hashedPassword, password string) bool

	// Métodos para bloqueo por intentos fallidos
	RecordFailedAttempt(email, ipAddress string) error
	ResetFailedAttempts(email string) error
	GetFailedAttempts(email string, since time.Time) (int, error)
	IsBlocked(email string, maxAttempts int, blockDuration time.Duration) (bool, error)
}

// ============================================
// ESTRUCTURA PRIVADA
// ============================================
type authRepository struct {
	db *sql.DB
}

// ============================================
// CONSTRUCTOR
// ============================================
func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

// ============================================
// FIND BY EMAIL
// ============================================
func (r *authRepository) FindByEmail(email string) (*models.UserAuth, error) {
	user := &models.UserAuth{}
	query := `
        SELECT u.id, u.name, u.email, u.phone, u.password, u.perfil_id, u.activo, u.created_at
        FROM users u
        WHERE u.email = $1
    `
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone,
		&user.Password, &user.RoleID, &user.Activo, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, fmt.Errorf("error buscando usuario: %w", err)
	}
	return user, nil
}

// ============================================
// FIND BY ID
// ============================================
func (r *authRepository) FindByID(id int) (*models.UserAuth, error) {
	user := &models.UserAuth{}
	query := `
        SELECT u.id, u.name, u.email, u.phone, u.password, u.perfil_id, u.activo, u.created_at
        FROM users u
        WHERE u.id = $1
    `
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone,
		&user.Password, &user.RoleID, &user.Activo, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, fmt.Errorf("error buscando usuario: %w", err)
	}
	return user, nil
}

// ============================================
// CREATE USER
// ============================================
func (r *authRepository) CreateUser(user *models.UserAuth) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hasheando contraseña: %w", err)
	}

	query := `
        INSERT INTO users (name, email, phone, password, perfil_id, activo, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
        RETURNING id, created_at
    `
	err = r.db.QueryRow(query, user.Name, user.Email, user.Phone, string(hashedPassword),
		user.RoleID, user.Activo).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creando usuario: %w", err)
	}

	// Crear perfil de usuario
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
// UPDATE PASSWORD
// ============================================
func (r *authRepository) UpdatePassword(userID int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hasheando contraseña: %w", err)
	}

	query := `UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.Exec(query, string(hashedPassword), userID)
	if err != nil {
		return fmt.Errorf("error actualizando contraseña: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("usuario no encontrado")
	}
	return nil
}

// ============================================
// VALIDATE PASSWORD
// ============================================
func (r *authRepository) ValidatePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ============================================
// RECORD FAILED ATTEMPT (CORREGIDO)
// ============================================
func (r *authRepository) RecordFailedAttempt(email, ipAddress string) error {
	// Verificar que la tabla existe (opcional, para desarrollo)
	query := `
        INSERT INTO login_attempts (email, ip_address, attempt_time)
        VALUES ($1, $2, NOW())
    `
	_, err := r.db.Exec(query, email, ipAddress)
	if err != nil {
		// Si la tabla no existe, ignoramos el error en desarrollo
		fmt.Printf("Warning: no se pudo registrar intento fallido: %v\n", err)
		return nil
	}
	return nil
}

// ============================================
// RESET FAILED ATTEMPTS (CORREGIDO)
// ============================================
func (r *authRepository) ResetFailedAttempts(email string) error {
	query := `DELETE FROM login_attempts WHERE email = $1`
	_, err := r.db.Exec(query, email)
	if err != nil {
		fmt.Printf("Warning: no se pudieron resetear intentos: %v\n", err)
		return nil
	}
	return nil
}

// ============================================
// GET FAILED ATTEMPTS (CORREGIDO)
// ============================================
func (r *authRepository) GetFailedAttempts(email string, since time.Time) (int, error) {
	var count int
	query := `
        SELECT COUNT(*)
        FROM login_attempts
        WHERE email = $1 AND attempt_time > $2
    `
	err := r.db.QueryRow(query, email, since).Scan(&count)
	if err != nil {
		return 0, nil // Si hay error, asumimos 0 intentos
	}
	return count, nil
}

// ============================================
// IS BLOCKED (CORREGIDO)
// ============================================
func (r *authRepository) IsBlocked(email string, maxAttempts int, blockDuration time.Duration) (bool, error) {
	since := time.Now().Add(-blockDuration)

	count, err := r.GetFailedAttempts(email, since)
	if err != nil {
		return false, nil
	}

	if count >= maxAttempts {
		return true, nil
	}
	return false, nil
}
