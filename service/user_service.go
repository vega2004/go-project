package service

import (
	"errors"
	"strings"
	"time"
	"tu-proyecto/model"
	"tu-proyecto/repository"
	"tu-proyecto/utils"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	// Métodos existentes
	CreateUser(form *model.UserForm) error

	// NUEVOS MÉTODOS PARA ADMIN
	GetAllUsers(filter *model.UserFilter) (*model.UserPaginatedResponse, error)
	GetUserByID(id int) (*model.UserAuth, error)
	CreateUserByAdmin(form *model.RegisterForm, roleID int) error
	UpdateUser(id int, form *model.RegisterForm, roleID int) error
	DeleteUser(id int) error
}

type userService struct {
	userRepo repository.UserRepository // Repositorio original (User)
	authRepo repository.AuthRepository // Repositorio de autenticación (UserAuth)
}

// NewUserService - Constructor que recibe ambos repositorios
func NewUserService(userRepo repository.UserRepository, authRepo repository.AuthRepository) UserService {
	return &userService{
		userRepo: userRepo,
		authRepo: authRepo,
	}
}

// ============================================
// MÉTODOS EXISTENTES (para registro público)
// ============================================

// validateForm - Validaciones para el formulario público
func (s *userService) validateForm(form *model.UserForm) error {
	// Validar que no estén vacíos
	if strings.TrimSpace(form.Name) == "" {
		return errors.New("el nombre es requerido")
	}

	if strings.TrimSpace(form.Email) == "" {
		return errors.New("el email es requerido")
	}

	if strings.TrimSpace(form.Phone) == "" {
		return errors.New("el teléfono es requerido")
	}

	// Sanitizar entradas
	form.Name = utils.SanitizeInput(form.Name)
	form.Email = utils.SanitizeInput(form.Email)
	form.Phone = utils.SanitizeInput(form.Phone)

	// Validar nombre
	if !utils.ValidateName(form.Name) {
		return errors.New("nombre inválido: solo letras y espacios permitidos (2-50 caracteres)")
	}

	// Validar email
	if !utils.ValidateEmail(form.Email) {
		return errors.New("email inválido: formato incorrecto")
	}

	// Validar teléfono
	if !utils.ValidatePhone(form.Phone) {
		return errors.New("teléfono inválido: solo números, +, - y espacios (8-15 caracteres)")
	}

	// Validar reCAPTCHA
	isValid, err := utils.ValidateRecaptcha(form.RecaptchaToken)
	if err != nil {
		return errors.New("error validando reCAPTCHA. Intente nuevamente")
	}
	if !isValid {
		return errors.New("verificación reCAPTCHA fallida. Por favor, marque la casilla 'No soy un robot'")
	}

	return nil
}

// CreateUser - Registro público de usuario (siempre role_id=2)
func (s *userService) CreateUser(form *model.UserForm) error {
	// Validar primero
	if err := s.validateForm(form); err != nil {
		return err
	}

	// Crear usuario (model.User no tiene role_id)
	user := &model.User{
		Name:      form.Name,
		Email:     form.Email,
		Phone:     form.Phone,
		CreatedAt: time.Now(),
	}

	// Insertar en base de datos usando userRepo
	err := s.userRepo.Create(user)
	if err != nil {
		// Verificar si es error de duplicado
		if strings.Contains(err.Error(), "duplicate") ||
			strings.Contains(err.Error(), "unique") {
			return errors.New("el email ya está registrado")
		}
		return errors.New("error al guardar en la base de datos")
	}

	return nil
}

// ============================================
// NUEVOS MÉTODOS PARA ADMIN (USAN authRepo)
// ============================================

// GetAllUsers - Obtiene todos los usuarios con paginación y filtros
func (s *userService) GetAllUsers(filter *model.UserFilter) (*model.UserPaginatedResponse, error) {
	return s.authRepo.GetAllUsers(filter)
}

// GetUserByID - Obtiene un usuario por su ID
func (s *userService) GetUserByID(id int) (*model.UserAuth, error) {
	return s.authRepo.FindByID(id)
}

// validateAdminForm - Validaciones para formularios de admin
func (s *userService) validateAdminForm(form *model.RegisterForm) error {
	// Validar nombre
	if !utils.ValidateName(form.Name) {
		return errors.New("nombre inválido: solo letras y espacios (2-50 caracteres)")
	}

	// Validar email
	if !utils.ValidateEmail(form.Email) {
		return errors.New("email inválido: formato incorrecto")
	}

	// Validar teléfono
	if !utils.ValidatePhone(form.Phone) {
		return errors.New("teléfono inválido: solo números, +, - y espacios (8-15 dígitos)")
	}

	// Validar contraseña (opcional en edición)
	if form.Password != "" && len(form.Password) < 6 {
		return errors.New("la contraseña debe tener al menos 6 caracteres")
	}

	return nil
}

// CreateUserByAdmin - Crea un usuario con rol específico (solo admin)
func (s *userService) CreateUserByAdmin(form *model.RegisterForm, roleID int) error {
	// Validar formato de campos
	if err := s.validateAdminForm(form); err != nil {
		return err
	}

	// Verificar si email ya existe
	existing, _ := s.authRepo.FindByEmail(form.Email)
	if existing != nil {
		return errors.New("el email ya está registrado")
	}

	// Validar que roleID sea válido (1,2,3)
	if roleID < 1 || roleID > 3 {
		roleID = 2 // user por defecto
	}

	// Hashear contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("error al procesar la contraseña")
	}

	// Crear usuario con rol asignado
	user := &model.UserAuth{
		Name:      form.Name,
		Email:     form.Email,
		Phone:     form.Phone,
		Password:  string(hashedPassword),
		RoleID:    roleID,
		CreatedAt: time.Now(),
	}

	// Guardar en base de datos usando authRepo
	return s.authRepo.CreateUser(user)
}

// UpdateUser - Actualiza un usuario existente
func (s *userService) UpdateUser(id int, form *model.RegisterForm, roleID int) error {
	// Validar formato de campos
	if err := s.validateAdminForm(form); err != nil {
		return err
	}

	// Verificar si el usuario existe
	existing, err := s.authRepo.FindByID(id)
	if err != nil {
		return errors.New("usuario no encontrado")
	}

	// Validar que roleID sea válido
	if roleID < 1 || roleID > 3 {
		roleID = existing.RoleID // mantener el existente
	}

	// Preparar datos actualizados
	existing.Name = form.Name
	existing.Email = form.Email
	existing.Phone = form.Phone
	existing.RoleID = roleID

	// Si se proporcionó nueva contraseña, actualizarla
	if form.Password != "" {
		if len(form.Password) < 6 {
			return errors.New("la contraseña debe tener al menos 6 caracteres")
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.New("error al procesar la contraseña")
		}
		existing.Password = string(hashedPassword)
	}

	// Guardar cambios usando authRepo
	return s.authRepo.UpdateUser(existing)
}

// DeleteUser - Elimina un usuario
func (s *userService) DeleteUser(id int) error {
	// Verificar que no sea el último admin (opcional - mejora futura)
	// Por ahora, solo eliminar
	return s.authRepo.DeleteUser(id)
}
