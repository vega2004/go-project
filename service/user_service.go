package service

import (
	"errors"
	"strings"
	"time"
	"tu-proyecto/model"
	"tu-proyecto/repository"
	"tu-proyecto/utils"
)

type UserService interface {
	CreateUser(form *model.UserForm) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

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

	// Validar nombre (sin caracteres especiales)
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

func (s *userService) CreateUser(form *model.UserForm) error {
	// Validar primero
	if err := s.validateForm(form); err != nil {
		return err
	}

	// Crear usuario
	user := &model.User{
		Name:      form.Name,
		Email:     form.Email,
		Phone:     form.Phone,
		CreatedAt: time.Now(),
	}

	// Insertar en base de datos
	err := s.repo.Create(user)
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
