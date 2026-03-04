package service

import (
	"errors"
	"strings"
	"time"
	"tu-proyecto/model"
	"tu-proyecto/repository"
	"tu-proyecto/utils"
)

type CrudService interface {
	Create(persona *model.Persona, userID int) error
	Update(persona *model.Persona) error
	Delete(id int) error
	FindByID(id int) (*model.Persona, error)
	FindAll(filter *model.PersonaFilter) (*model.PaginatedResponse, error)
}

type crudService struct {
	repo repository.CrudRepository
}

func NewCrudService(repo repository.CrudRepository) CrudService {
	return &crudService{repo: repo}
}

func (s *crudService) Create(persona *model.Persona, userID int) error {
	// Validar datos
	if strings.TrimSpace(persona.Nombre) == "" {
		return errors.New("el nombre es requerido")
	}

	// Sanitizar
	persona.Nombre = utils.SanitizeInput(persona.Nombre)

	// Validar nombre
	if !utils.ValidateName(persona.Nombre) {
		return errors.New("nombre inválido: solo letras y espacios (2-100 caracteres)")
	}

	// Validar estado civil
	validStates := map[string]bool{
		"soltero": true, "casado": true, "divorciado": true, "viudo": true,
	}
	if !validStates[persona.EstadoCivil] {
		return errors.New("estado civil inválido")
	}

	persona.UserID = userID
	persona.CreatedAt = time.Now()
	persona.UpdatedAt = time.Now()

	return s.repo.Create(persona)
}

func (s *crudService) Update(persona *model.Persona) error {
	// Validar datos
	if strings.TrimSpace(persona.Nombre) == "" {
		return errors.New("el nombre es requerido")
	}

	persona.Nombre = utils.SanitizeInput(persona.Nombre)

	if !utils.ValidateName(persona.Nombre) {
		return errors.New("nombre inválido: solo letras y espacios")
	}

	validStates := map[string]bool{
		"soltero": true, "casado": true, "divorciado": true, "viudo": true,
	}
	if !validStates[persona.EstadoCivil] {
		return errors.New("estado civil inválido")
	}

	persona.UpdatedAt = time.Now()
	return s.repo.Update(persona)
}

func (s *crudService) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *crudService) FindByID(id int) (*model.Persona, error) {
	return s.repo.FindByID(id)
}

func (s *crudService) FindAll(filter *model.PersonaFilter) (*model.PaginatedResponse, error) {
	if filter.PageSize <= 0 || filter.PageSize > 50 {
		filter.PageSize = 5
	}
	return s.repo.FindAll(filter)
}
