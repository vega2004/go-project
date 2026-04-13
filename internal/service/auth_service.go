package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/repository"
	"tu-proyecto/internal/utils"
)

// ============================================
// INTERFAZ - ACTUALIZADA
// ============================================
type AuthService interface {
	Register(form *models.RegisterForm, ipAddress string) error             // ← AÑADIR ipAddress
	Login(email, password, recaptchaToken string) (*models.UserAuth, error) // ← SIN ipAddress
	GetUserByID(id int) (*models.UserAuth, error)
	GetRolName(roleID int) string
	RecordFailedAttempt(email string) error
	ResetFailedAttempts(email string) error
	IsBlocked(email string) (bool, error)
}

// ============================================
// IMPLEMENTACIÓN
// ============================================
type authService struct {
	repo            repository.AuthRepository
	recaptchaSecret string
	maxAttempts     int
	blockDuration   time.Duration
}

func NewAuthService(repo repository.AuthRepository, recaptchaSecret string) AuthService {
	return &authService{
		repo:            repo,
		recaptchaSecret: recaptchaSecret,
		maxAttempts:     5,
		blockDuration:   15 * time.Minute,
	}
}

// Register - Registro de nuevo usuario (ACTUALIZADO con ipAddress)
func (s *authService) Register(form *models.RegisterForm, ipAddress string) error {
	// Validar reCAPTCHA
	isValid, err := utils.ValidateRecaptcha(form.RecaptchaToken, s.recaptchaSecret, 0.5, "register")
	if err != nil || !isValid {
		return errors.New("verificación reCAPTCHA fallida")
	}

	// Validar formato
	if err := s.validateRegisterForm(form); err != nil {
		return err
	}

	// Verificar si el email ya existe
	existing, _ := s.repo.FindByEmail(form.Email)
	if existing != nil {
		return errors.New("el email ya está registrado")
	}

	// Crear usuario con role_id=2 (usuario normal)
	user := &models.UserAuth{
		Name:      strings.TrimSpace(form.Name),
		Email:     strings.TrimSpace(form.Email),
		Phone:     strings.TrimSpace(form.Phone),
		Password:  form.Password,
		RoleID:    2,
		Activo:    true,
		CreatedAt: time.Now(),
	}

	log.Printf("[AUDIT] Registro de nuevo usuario desde IP %s: %s", ipAddress, form.Email)

	return s.repo.CreateUser(user)
}

// Login - Inicio de sesión (SIN ipAddress)
func (s *authService) Login(email, password, recaptchaToken string) (*models.UserAuth, error) {
	// Verificar si está bloqueado
	blocked, err := s.IsBlocked(email)
	if err != nil {
		return nil, fmt.Errorf("error verificando bloqueo: %w", err)
	}
	if blocked {
		return nil, errors.New("demasiados intentos fallidos. Espere 15 minutos")
	}

	// Validar reCAPTCHA
	isValid, err := utils.ValidateRecaptcha(recaptchaToken, s.recaptchaSecret, 0.5, "login")
	if err != nil || !isValid {
		s.RecordFailedAttempt(email)
		return nil, errors.New("verificación reCAPTCHA fallida")
	}

	// Buscar usuario
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		s.RecordFailedAttempt(email)
		return nil, errors.New("email o contraseña incorrectos")
	}

	// Verificar si está activo
	if !user.Activo {
		return nil, errors.New("usuario desactivado, contacte al administrador")
	}

	// Validar contraseña
	if !s.repo.ValidatePassword(user.Password, password) {
		s.RecordFailedAttempt(email)
		return nil, errors.New("email o contraseña incorrectos")
	}

	// Resetear intentos fallidos
	s.ResetFailedAttempts(email)

	return user, nil
}

// RecordFailedAttempt - Registra intento fallido
func (s *authService) RecordFailedAttempt(email string) error {
	return s.repo.RecordFailedAttempt(email, "")
}

// ResetFailedAttempts - Resetea intentos fallidos
func (s *authService) ResetFailedAttempts(email string) error {
	return s.repo.ResetFailedAttempts(email)
}

// IsBlocked - Verifica si está bloqueado
func (s *authService) IsBlocked(email string) (bool, error) {
	return s.repo.IsBlocked(email, s.maxAttempts, s.blockDuration)
}

// GetUserByID - Obtiene usuario por ID
func (s *authService) GetUserByID(id int) (*models.UserAuth, error) {
	if id <= 0 {
		return nil, errors.New("ID inválido")
	}
	return s.repo.FindByID(id)
}

// GetRolName - Obtiene nombre del rol
func (s *authService) GetRolName(roleID int) string {
	switch roleID {
	case 1:
		return "administrador"
	case 2:
		return "usuario"
	case 3:
		return "editor"
	default:
		return "usuario"
	}
}

// validateRegisterForm - Validaciones del formulario de registro
func (s *authService) validateRegisterForm(form *models.RegisterForm) error {
	if !utils.ValidateName(form.Name) {
		return errors.New("nombre inválido: solo letras y espacios (2-50 caracteres)")
	}
	if !utils.ValidateEmail(form.Email) {
		return errors.New("email inválido")
	}
	if !utils.ValidatePhone(form.Phone) {
		return errors.New("teléfono inválido (8-15 dígitos)")
	}
	if len(form.Password) < 6 {
		return errors.New("la contraseña debe tener al menos 6 caracteres")
	}
	return nil
}
