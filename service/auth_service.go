package service

import (
	"errors"
	"time"
	"tu-proyecto/model"
	"tu-proyecto/repository"
	"tu-proyecto/utils"
)

type AuthService interface {
	// Registro de nuevo usuario (siempre role_id=2 por defecto)
	Register(form *model.RegisterForm) error

	// Login de usuario (obtiene rol y permisos)
	Login(email, password, recaptchaToken string) (*model.UserAuth, error)

	// Obtener rol por ID
	GetRolByID(roleID int) (string, error)
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authService{repo: repo}
}

// Register - Registra un nuevo usuario con role_id=2 (user)
func (s *authService) Register(form *model.RegisterForm) error {
	// 1. Validar reCAPTCHA
	isValid, err := utils.ValidateRecaptcha(form.RecaptchaToken)
	if err != nil || !isValid {
		return errors.New("verificación reCAPTCHA fallida")
	}

	// 2. Validar formato de campos
	if err := s.validateRegisterForm(form); err != nil {
		return err
	}

	// 3. Verificar si email ya existe
	existing, _ := s.repo.FindByEmail(form.Email)
	if existing != nil {
		return errors.New("el email ya está registrado")
	}

	// 4. Crear usuario con role_id=2 (user por defecto)
	user := &model.UserAuth{
		Name:      form.Name,
		Email:     form.Email,
		Phone:     form.Phone,
		Password:  form.Password,
		RoleID:    2, // ← SIEMPRE user al registrarse
		CreatedAt: time.Now(),
	}

	// 5. Guardar en base de datos
	return s.repo.CreateUser(user)
}

// Login - Autentica un usuario y devuelve sus datos incluyendo rol
func (s *authService) Login(email, password, recaptchaToken string) (*model.UserAuth, error) {
	// 1. Validar reCAPTCHA
	isValid, err := utils.ValidateRecaptcha(recaptchaToken)
	if err != nil || !isValid {
		return nil, errors.New("verificación reCAPTCHA fallida")
	}

	// 2. Buscar usuario por email
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("email o contraseña incorrectos")
	}

	// 3. Validar contraseña
	if !s.repo.ValidatePassword(user.Password, password) {
		return nil, errors.New("email o contraseña incorrectos")
	}

	// 4. Obtener nombre del rol (para mostrar en UI)
	rolNombre, err := s.GetRolByID(user.RoleID)
	if err == nil {
		// Podrías agregar esto si tuvieras campo RolNombre en UserAuth
		// user.RolNombre = rolNombre
		_ = rolNombre // Ignoramos por ahora, lo manejamos en session
	}

	return user, nil
}

// GetRolByID - Obtiene el nombre del rol por su ID
func (s *authService) GetRolByID(roleID int) (string, error) {
	switch roleID {
	case 1:
		return "admin", nil
	case 2:
		return "user", nil
	case 3:
		return "editor", nil
	default:
		return "user", nil // Por defecto
	}
}

// validateRegisterForm - Validaciones adicionales del formulario
func (s *authService) validateRegisterForm(form *model.RegisterForm) error {
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

	// Validar contraseña (mínimo 6 caracteres)
	if len(form.Password) < 6 {
		return errors.New("la contraseña debe tener al menos 6 caracteres")
	}

	return nil
}

// RegisterByAdmin - Registro realizado por administrador (puede asignar rol)
func (s *authService) RegisterByAdmin(form *model.RegisterForm, roleID int) error {
	// 1. Validar reCAPTCHA (opcional para admin)
	if form.RecaptchaToken != "" {
		isValid, err := utils.ValidateRecaptcha(form.RecaptchaToken)
		if err != nil || !isValid {
			return errors.New("verificación reCAPTCHA fallida")
		}
	}

	// 2. Validar formato de campos
	if err := s.validateRegisterForm(form); err != nil {
		return err
	}

	// 3. Verificar si email ya existe
	existing, _ := s.repo.FindByEmail(form.Email)
	if existing != nil {
		return errors.New("el email ya está registrado")
	}

	// 4. Validar que roleID sea válido (1,2,3)
	if roleID < 1 || roleID > 3 {
		roleID = 2 // user por defecto
	}

	// 5. Crear usuario con rol asignado por admin
	user := &model.UserAuth{
		Name:      form.Name,
		Email:     form.Email,
		Phone:     form.Phone,
		Password:  form.Password,
		RoleID:    roleID,
		CreatedAt: time.Now(),
	}

	// 6. Guardar en base de datos
	return s.repo.CreateUser(user)
}

// UpdateUserRole - Actualizar rol de un usuario (solo admin)
func (s *authService) UpdateUserRole(userID, newRoleID int) error {
	// Validar que newRoleID sea válido
	if newRoleID < 1 || newRoleID > 3 {
		return errors.New("rol inválido")
	}

	// Aquí iría la lógica para actualizar el rol en BD
	// Necesitarías un método en el repositorio: UpdateUserRole
	return s.repo.UpdateUserRole(userID, newRoleID)
}

// ChangePassword - Cambiar contraseña (para perfil de usuario)
func (s *authService) ChangePassword(userID int, oldPassword, newPassword string) error {
	// 1. Obtener usuario
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errors.New("usuario no encontrado")
	}

	// 2. Validar contraseña actual
	if !s.repo.ValidatePassword(user.Password, oldPassword) {
		return errors.New("contraseña actual incorrecta")
	}

	// 3. Validar nueva contraseña
	if len(newPassword) < 6 {
		return errors.New("la nueva contraseña debe tener al menos 6 caracteres")
	}

	// 4. Actualizar contraseña
	return s.repo.UpdatePassword(userID, newPassword)
}
