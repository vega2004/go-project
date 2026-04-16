package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/repository"
	"tu-proyecto/internal/utils"
)

// ============================================
// INTERFAZ PÚBLICA
// ============================================
type UserService interface {
	Create(req *models.UserCreateRequest) error
	Update(req *models.UserUpdateRequest, auditorID int) error
	Delete(id int, auditorID int) error
	GetByID(id int) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll(filter *models.UserFilter) (*models.UserPaginatedResponse, error)
	UpdatePassword(id int, newPassword string) error
	UpdateStatus(id int, activo bool, auditorID int) error
	GetUserWithProfile(id int) (*models.UserWithProfile, error)
	ExistsByID(id int) (bool, error)
}

// ============================================
// ESTRUCTURA PRIVADA
// ============================================
type userService struct {
	userRepo repository.UserRepository
}

// ============================================
// CONSTRUCTOR
// ============================================
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// ============================================
// VALIDACIONES PRIVADAS
// ============================================

// validateUserData - Validaciones comunes para nombre, email, teléfono
func (s *userService) validateUserData(name, email, phone string) error {
	if !utils.ValidateName(name) {
		return errors.New("nombre inválido: solo letras y espacios (2-50 caracteres)")
	}
	if !utils.ValidateEmail(email) {
		return errors.New("email inválido")
	}
	if phone != "" && !utils.ValidatePhone(phone) {
		return errors.New("teléfono inválido (8-15 dígitos)")
	}
	return nil
}

// validatePassword - Validación de contraseña
func (s *userService) validatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("la contraseña debe tener al menos 6 caracteres")
	}
	return nil
}

// ============================================
// MÉTODOS PÚBLICOS
// ============================================

// Create - Crear nuevo usuario
func (s *userService) Create(req *models.UserCreateRequest) error {
	// Validar datos
	if err := s.validateUserData(req.Name, req.Email, req.Phone); err != nil {
		return err
	}

	if err := s.validatePassword(req.Password); err != nil {
		return err
	}

	// Validar que las contraseñas coincidan
	if req.Password != req.ConfirmPassword {
		return errors.New("las contraseñas no coinciden")
	}

	if req.PerfilID <= 0 {
		return errors.New("debe seleccionar un perfil válido")
	}

	// Verificar email único
	unique, err := s.userRepo.IsEmailUnique(req.Email, 0)
	if err != nil {
		return fmt.Errorf("error verificando email: %w", err)
	}
	if !unique {
		return errors.New("el email ya está registrado")
	}

	// Crear usuario
	user := &models.User{
		Name:     strings.TrimSpace(req.Name),
		Email:    strings.TrimSpace(req.Email),
		Phone:    strings.TrimSpace(req.Phone),
		Password: req.Password,
		PerfilID: req.PerfilID,
		Activo:   req.Activo,
	}

	log.Printf("[AUDIT] Creando nuevo usuario: %s (email: %s, perfil_id: %d)", user.Name, user.Email, user.PerfilID)
	return s.userRepo.Create(user)
}

// Update - Actualizar usuario existente
func (s *userService) Update(req *models.UserUpdateRequest, auditorID int) error {
	// Validar datos
	if err := s.validateUserData(req.Name, req.Email, req.Phone); err != nil {
		return err
	}

	if req.PerfilID <= 0 {
		return errors.New("debe seleccionar un perfil válido")
	}

	// Verificar que el usuario existe
	existing, err := s.userRepo.FindByID(req.ID)
	if err != nil {
		return fmt.Errorf("usuario no encontrado: %w", err)
	}

	// Verificar email único (excluyendo el usuario actual)
	unique, err := s.userRepo.IsEmailUnique(req.Email, req.ID)
	if err != nil {
		return fmt.Errorf("error verificando email: %w", err)
	}
	if !unique {
		return errors.New("el email ya está registrado por otro usuario")
	}

	// Si se proporcionó nueva contraseña, validarla y actualizarla
	if req.Password != "" {
		if err := s.validatePassword(req.Password); err != nil {
			return err
		}
		if err := s.userRepo.UpdatePassword(req.ID, req.Password); err != nil {
			return fmt.Errorf("error actualizando contraseña: %w", err)
		}
		log.Printf("[AUDIT] Usuario %d actualizó la contraseña del usuario %d", auditorID, req.ID)
	}

	// Actualizar datos del usuario
	user := &models.User{
		ID:       req.ID,
		Name:     strings.TrimSpace(req.Name),
		Email:    strings.TrimSpace(req.Email),
		Phone:    strings.TrimSpace(req.Phone),
		PerfilID: req.PerfilID,
		Activo:   req.Activo,
	}

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// Log de auditoría
	log.Printf("[AUDIT] Usuario %d actualizó usuario ID=%d (nombre: %s, email: %s, perfil: %d->%d)",
		auditorID, existing.ID, req.Name, req.Email, existing.PerfilID, req.PerfilID)

	return nil
}

// Delete - Eliminar usuario
func (s *userService) Delete(id int, auditorID int) error {
	if id <= 0 {
		return errors.New("ID inválido")
	}

	// Verificar que el usuario existe
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("usuario no encontrado: %w", err)
	}

	// No permitir eliminar el propio usuario
	if id == auditorID {
		return errors.New("no puedes eliminar tu propio usuario")
	}

	// Log de auditoría
	log.Printf("[AUDIT] Usuario %d eliminó usuario ID=%d (%s, email: %s)",
		auditorID, user.ID, user.Name, user.Email)

	return s.userRepo.Delete(id)
}

// GetByID - Obtener usuario por ID
func (s *userService) GetByID(id int) (*models.User, error) {
	if id <= 0 {
		return nil, errors.New("ID inválido")
	}
	return s.userRepo.FindByID(id)
}

// GetByEmail - Obtener usuario por email
func (s *userService) GetByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email requerido")
	}
	return s.userRepo.FindByEmail(email)
}

// GetAll - Listar usuarios con filtros y paginación
func (s *userService) GetAll(filter *models.UserFilter) (*models.UserPaginatedResponse, error) {
	if filter == nil {
		filter = &models.UserFilter{}
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	return s.userRepo.FindAll(filter)
}

// UpdatePassword - Actualizar contraseña de usuario
func (s *userService) UpdatePassword(id int, newPassword string) error {
	if id <= 0 {
		return errors.New("ID inválido")
	}
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(id, newPassword)
}

// UpdateStatus - Actualizar estado activo/inactivo del usuario
func (s *userService) UpdateStatus(id int, activo bool, auditorID int) error {
	if id <= 0 {
		return errors.New("ID inválido")
	}

	// No permitir desactivar el propio usuario
	if id == auditorID && !activo {
		return errors.New("no puedes desactivar tu propio usuario")
	}

	// Obtener usuario para verificar si es administrador
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("usuario no encontrado: %w", err)
	}

	// Si se está desactivando un administrador, verificar que no sea el último
	if !activo && user.PerfilID == 1 {
		adminCount, err := s.userRepo.CountActiveAdmins()
		if err != nil {
			return fmt.Errorf("error verificando administradores: %w", err)
		}
		if adminCount <= 1 {
			return errors.New("no puedes desactivar al último administrador del sistema")
		}
	}

	log.Printf("[AUDIT] Usuario %d cambió estado de usuario %d a activo=%v", auditorID, id, activo)
	return s.userRepo.UpdateStatus(id, activo)
}

// GetUserWithProfile - Obtener usuario con datos completos del perfil
func (s *userService) GetUserWithProfile(id int) (*models.UserWithProfile, error) {
	if id <= 0 {
		return nil, errors.New("ID inválido")
	}
	return s.userRepo.GetUserWithProfile(id)
}

// ExistsByID - Verificar si existe un usuario por ID
func (s *userService) ExistsByID(id int) (bool, error) {
	if id <= 0 {
		return false, errors.New("ID inválido")
	}
	return s.userRepo.ExistsByID(id)
}
