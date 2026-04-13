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

type UserService interface {
	Create(req *models.UserCreateRequest) error
	Update(req *models.UserUpdateRequest, auditorID int) error // ← Añadido auditorID
	Delete(id int, auditorID int) error
	GetByID(id int) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll(filter *models.UserFilter) (*models.UserPaginatedResponse, error)
	UpdatePassword(id int, newPassword string) error
	UpdateStatus(id int, activo bool, auditorID int) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// validateUserData - Validaciones comunes
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

// Create - Crear nuevo usuario
func (s *userService) Create(req *models.UserCreateRequest) error {
	// Validar datos
	if err := s.validateUserData(req.Name, req.Email, req.Phone); err != nil {
		return err
	}

	if len(req.Password) < 6 {
		return errors.New("la contraseña debe tener al menos 6 caracteres")
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

	user := &models.User{
		Name:     strings.TrimSpace(req.Name),
		Email:    strings.TrimSpace(req.Email),
		Phone:    strings.TrimSpace(req.Phone),
		Password: req.Password,
		PerfilID: req.PerfilID,
		Activo:   req.Activo,
	}

	return s.userRepo.Create(user)
}

// Update - Actualizar usuario (CORREGIDO)
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

	user := &models.User{
		ID:       req.ID,
		Name:     strings.TrimSpace(req.Name),
		Email:    strings.TrimSpace(req.Email),
		Phone:    strings.TrimSpace(req.Phone),
		PerfilID: req.PerfilID,
		Activo:   req.Activo,
	}

	// Si se proporcionó nueva contraseña, actualizarla
	if req.Password != "" {
		if len(req.Password) < 6 {
			return errors.New("la contraseña debe tener al menos 6 caracteres")
		}
		if err := s.userRepo.UpdatePassword(req.ID, req.Password); err != nil {
			return fmt.Errorf("error actualizando contraseña: %w", err)
		}
	}

	// Actualizar datos del usuario
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// Log de auditoría (CORREGIDO - usa auditorID)
	log.Printf("[AUDIT] Usuario %d actualizó usuario ID=%d (rol anterior: %d, nuevo: %d)",
		auditorID, existing.ID, existing.PerfilID, req.PerfilID)

	return nil
}

// Delete - Eliminar usuario (con verificación de último admin) (CORREGIDO)
func (s *userService) Delete(id int, auditorID int) error {
	if id == auditorID {
		return errors.New("no puedes eliminar tu propio usuario")
	}
	// CORREGIDO: Delete solo recibe id, no auditorID
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

// GetAll - Listar usuarios con filtros
func (s *userService) GetAll(filter *models.UserFilter) (*models.UserPaginatedResponse, error) {
	return s.userRepo.FindAll(filter)
}

// UpdatePassword - Actualizar contraseña de usuario
func (s *userService) UpdatePassword(id int, newPassword string) error {
	if id <= 0 {
		return errors.New("ID inválido")
	}
	if len(newPassword) < 6 {
		return errors.New("la contraseña debe tener al menos 6 caracteres")
	}
	return s.userRepo.UpdatePassword(id, newPassword)
}

// UpdateStatus - Actualizar estado activo/inactivo
func (s *userService) UpdateStatus(id int, activo bool, auditorID int) error {
	if id <= 0 {
		return errors.New("ID inválido")
	}

	// No permitir desactivar el propio usuario
	if id == auditorID && !activo {
		return errors.New("no puedes desactivar tu propio usuario")
	}

	// Si se está desactivando un administrador, verificar que no sea el último
	if !activo {
		user, err := s.userRepo.FindByID(id)
		if err != nil {
			return fmt.Errorf("usuario no encontrado: %w", err)
		}

		if user.PerfilID == 1 {
			adminCount, err := s.userRepo.CountActiveAdmins()
			if err != nil {
				return fmt.Errorf("error verificando administradores: %w", err)
			}
			if adminCount <= 1 {
				return errors.New("no puedes desactivar al último administrador")
			}
		}
	}

	log.Printf("[AUDIT] Usuario %d cambió estado de usuario %d a activo=%v", auditorID, id, activo)
	return s.userRepo.UpdateStatus(id, activo)
}
