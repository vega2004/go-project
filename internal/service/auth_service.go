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
// INTERFAZ
// ============================================
type AuthService interface {
	Register(form *models.RegisterForm, ipAddress string) error
	Login(email, password, recaptchaToken string) (*models.UserAuth, error)
	GetUserByID(id int) (*models.UserAuth, error)
	GetPerfilNombre(perfilID int) string
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

// ============================================
// REGISTER - Registro de nuevo usuario
// ============================================
// ============================================
// REGISTER - Registro de nuevo usuario
// ============================================
func (s *authService) Register(form *models.RegisterForm, ipAddress string) error {
	// Validar reCAPTCHA - Para v2 (casilla), usar minScore=0 y action vacío
	log.Printf("[RECAPTCHA] Validando token...")
	isValid, err := utils.ValidateRecaptcha(form.RecaptchaToken, s.recaptchaSecret, 0, "") // ← Cambiado
	if err != nil || !isValid {
		log.Printf("[RECAPTCHA] Error: %v, isValid: %v", err, isValid)
		return errors.New("verificación reCAPTCHA fallida")
	}

	// Validar formato
	if err := s.validateRegisterForm(form); err != nil {
		return err
	}

	// Validar que las contraseñas coincidan
	if form.Password != form.ConfirmPassword {
		return errors.New("las contraseñas no coinciden")
	}

	// Verificar si el email ya existe
	existing, _ := s.repo.FindByEmail(form.Email)
	if existing != nil {
		return errors.New("el email ya está registrado")
	}

	// Obtener ID del perfil "usuario" (por defecto)
	perfilID := s.getDefaultPerfilID()

	// Crear usuario con perfil por defecto
	user := &models.UserAuth{
		Name:      strings.TrimSpace(form.Name),
		Email:     strings.TrimSpace(form.Email),
		Phone:     strings.TrimSpace(form.Phone),
		Password:  form.Password,
		PerfilID:  perfilID, // ← Cambiado de RoleID a PerfilID
		Activo:    true,
		CreatedAt: time.Now(),
	}

	log.Printf("[AUDIT] Registro de nuevo usuario desde IP %s: %s (perfil_id: %d)", ipAddress, form.Email, perfilID)

	return s.repo.CreateUser(user)
}

// ============================================
// LOGIN - Inicio de sesión
// ============================================
func (s *authService) Login(email, password, recaptchaToken string) (*models.UserAuth, error) {
	// Verificar si está bloqueado
	blocked, err := s.IsBlocked(email)
	if err != nil {
		return nil, fmt.Errorf("error verificando bloqueo: %w", err)
	}
	if blocked {
		return nil, errors.New("demasiados intentos fallidos. Espere 15 minutos")
	}

	// Validar reCAPTCHA - Para v2 (casilla), usar minScore=0 y action vacío
	isValid, err := utils.ValidateRecaptcha(recaptchaToken, s.recaptchaSecret, 0, "")
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

	log.Printf("[AUDIT] Usuario %d (%s) inició sesión exitosamente", user.ID, user.Email)

	return user, nil
}

// ============================================
// RECORD FAILED ATTEMPT
// ============================================
func (s *authService) RecordFailedAttempt(email string) error {
	return s.repo.RecordFailedAttempt(email, "")
}

// ============================================
// RESET FAILED ATTEMPTS
// ============================================
func (s *authService) ResetFailedAttempts(email string) error {
	return s.repo.ResetFailedAttempts(email)
}

// ============================================
// IS BLOCKED
// ============================================
func (s *authService) IsBlocked(email string) (bool, error) {
	return s.repo.IsBlocked(email, s.maxAttempts, s.blockDuration)
}

// ============================================
// GET USER BY ID
// ============================================
func (s *authService) GetUserByID(id int) (*models.UserAuth, error) {
	if id <= 0 {
		return nil, errors.New("ID inválido")
	}
	return s.repo.FindByID(id)
}

// ============================================
// GET PERFIL NOMBRE
// ============================================
func (s *authService) GetPerfilNombre(perfilID int) string {
	switch perfilID {
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

// ============================================
// GET DEFAULT PERFIL ID
// ============================================
func (s *authService) getDefaultPerfilID() int {
	// Buscar el perfil "usuario" en la BD
	// Por ahora retornamos 2, pero idealmente deberías consultar
	return 2
}

// ============================================
// VALIDATE REGISTER FORM
// ============================================
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
