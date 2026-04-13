package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/repository"
	"tu-proyecto/internal/utils"
)

// ============================================
// INTERFAZ PÚBLICA - COMPLETA
// ============================================
type PerfilService interface {
	// CRUD para perfiles/roles
	Create(perfil *models.Perfil) error
	Update(perfil *models.Perfil) error
	Delete(id int) error
	GetByID(id int) (*models.Perfil, error)
	GetAll(filter *models.PerfilFilter) (*models.PerfilPaginatedResponse, error)

	// Métodos para perfil de usuario
	GetPerfil(userID int) (*models.PerfilUsuario, error)
	UpdatePerfil(userID int, perfil *models.PerfilUsuario) error
	UpdateFoto(userID int, file multipart.File, header *multipart.FileHeader) (string, error)
	ChangePassword(userID int, form *models.CambioPassword) error
	DeleteFoto(userID int) error
}

// ============================================
// ESTRUCTURA PRIVADA - COMPLETA
// ============================================
type perfilService struct {
	perfilRepo repository.PerfilRepository
	authRepo   repository.AuthRepository
	userRepo   repository.UserRepository
}

// ============================================
// CONSTRUCTOR - ACTUALIZADO
// ============================================
func NewPerfilService(perfilRepo repository.PerfilRepository, authRepo repository.AuthRepository, userRepo repository.UserRepository) PerfilService {
	return &perfilService{
		perfilRepo: perfilRepo,
		authRepo:   authRepo,
		userRepo:   userRepo,
	}
}

// ============================================
// CRUD PARA PERFILES/ROLES
// ============================================

// Create - Crea un nuevo perfil/rol
func (s *perfilService) Create(perfil *models.Perfil) error {
	perfil.Nombre = strings.TrimSpace(perfil.Nombre)
	perfil.Descripcion = strings.TrimSpace(perfil.Descripcion)

	if perfil.Nombre == "" {
		return errors.New("el nombre del perfil es requerido")
	}
	if len(perfil.Nombre) < 3 {
		return errors.New("el nombre debe tener al menos 3 caracteres")
	}
	if len(perfil.Nombre) > 50 {
		return errors.New("el nombre no puede exceder 50 caracteres")
	}
	return s.perfilRepo.Create(perfil)
}

// Update - Actualiza un perfil/rol existente
func (s *perfilService) Update(perfil *models.Perfil) error {
	// Verificar que existe
	_, err := s.perfilRepo.FindByID(perfil.ID)
	if err != nil {
		return fmt.Errorf("perfil no encontrado: %w", err)
	}

	perfil.Nombre = strings.TrimSpace(perfil.Nombre)
	perfil.Descripcion = strings.TrimSpace(perfil.Descripcion)

	if perfil.Nombre == "" {
		return errors.New("el nombre del perfil es requerido")
	}
	if len(perfil.Nombre) < 3 {
		return errors.New("el nombre debe tener al menos 3 caracteres")
	}
	if len(perfil.Nombre) > 50 {
		return errors.New("el nombre no puede exceder 50 caracteres")
	}
	return s.perfilRepo.Update(perfil)
}

// Delete - Elimina un perfil/rol (con verificación de uso)
func (s *perfilService) Delete(id int) error {
	// Verificar que existe
	_, err := s.perfilRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("perfil no encontrado: %w", err)
	}

	// No permitir eliminar perfil de Administrador
	if id == 1 {
		return errors.New("no se puede eliminar el perfil de Administrador")
	}

	// Verificar si está en uso
	inUse, err := s.perfilRepo.IsInUse(id)
	if err != nil {
		return fmt.Errorf("error verificando uso del perfil: %w", err)
	}
	if inUse {
		return errors.New("no se puede eliminar el perfil porque tiene usuarios asignados")
	}
	return s.perfilRepo.Delete(id)
}

// GetByID - Obtiene un perfil/rol por ID
func (s *perfilService) GetByID(id int) (*models.Perfil, error) {
	if id <= 0 {
		return nil, errors.New("ID inválido")
	}
	return s.perfilRepo.FindByID(id)
}

// GetAll - Lista perfiles/roles con paginación
func (s *perfilService) GetAll(filter *models.PerfilFilter) (*models.PerfilPaginatedResponse, error) {
	if filter == nil {
		filter = &models.PerfilFilter{Page: 1, PageSize: 10}
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	return s.perfilRepo.FindAll(filter)
}

// ============================================
// MÉTODOS PARA PERFIL DE USUARIO
// ============================================

// GetPerfil - Obtiene el perfil de usuario (foto, bio, dirección)
func (s *perfilService) GetPerfil(userID int) (*models.PerfilUsuario, error) {
	if userID <= 0 {
		return nil, errors.New("ID de usuario inválido")
	}
	return s.perfilRepo.GetPerfilByUserID(userID)
}

// UpdatePerfil - Actualiza los datos del perfil de usuario
func (s *perfilService) UpdatePerfil(userID int, perfil *models.PerfilUsuario) error {
	if userID <= 0 {
		return errors.New("ID de usuario inválido")
	}
	if perfil == nil {
		return errors.New("datos de perfil inválidos")
	}

	// Limpiar campos
	perfil.Bio = strings.TrimSpace(perfil.Bio)
	perfil.Direccion = strings.TrimSpace(perfil.Direccion)
	perfil.TelefonoAlterno = strings.TrimSpace(perfil.TelefonoAlterno)

	return s.perfilRepo.UpdatePerfil(userID, perfil)
}

// UpdateFoto - Actualiza la foto de perfil del usuario
func (s *perfilService) UpdateFoto(userID int, file multipart.File, header *multipart.FileHeader) (string, error) {
	if userID <= 0 {
		return "", errors.New("ID de usuario inválido")
	}
	if file == nil || header == nil {
		return "", errors.New("archivo de imagen inválido")
	}

	// Validar y guardar imagen
	uploadedFile, err := utils.SaveProfileImage(file, header)
	if err != nil {
		return "", fmt.Errorf("error guardando imagen: %w", err)
	}

	// Obtener perfil actual para eliminar foto anterior
	perfilActual, _ := s.perfilRepo.GetPerfilByUserID(userID)
	if perfilActual != nil && perfilActual.FotoPath != "" && perfilActual.FotoPath != "/static/uploads/perfil/default-avatar.png" {
		utils.DeleteFile(perfilActual.FotoPath)
	}

	// Actualizar en BD
	err = s.perfilRepo.UpdateFoto(userID, uploadedFile.Path, uploadedFile.OriginalName, uploadedFile.MimeType, int(uploadedFile.Size))
	if err != nil {
		return "", fmt.Errorf("error actualizando foto en BD: %w", err)
	}

	return uploadedFile.Path, nil
}

// ChangePassword - Cambia la contraseña del usuario
func (s *perfilService) ChangePassword(userID int, form *models.CambioPassword) error {
	if userID <= 0 {
		return errors.New("ID de usuario inválido")
	}
	if form == nil {
		return errors.New("datos de cambio de contraseña inválidos")
	}

	// Validar que las contraseñas coincidan
	if form.Nueva != form.Confirmar {
		return errors.New("las contraseñas no coinciden")
	}

	// Validar longitud mínima
	if len(form.Nueva) < 6 {
		return errors.New("la nueva contraseña debe tener al menos 6 caracteres")
	}

	// Obtener usuario actual
	user, err := s.authRepo.FindByID(userID)
	if err != nil {
		return errors.New("usuario no encontrado")
	}

	// Validar contraseña actual
	if !s.authRepo.ValidatePassword(user.Password, form.Actual) {
		return errors.New("contraseña actual incorrecta")
	}

	// Actualizar contraseña
	return s.authRepo.UpdatePassword(userID, form.Nueva)
}

// DeleteFoto - Elimina la foto de perfil (restaura la default)
func (s *perfilService) DeleteFoto(userID int) error {
	if userID <= 0 {
		return errors.New("ID de usuario inválido")
	}

	// Obtener perfil actual
	perfil, err := s.perfilRepo.GetPerfilByUserID(userID)
	if err != nil {
		return errors.New("perfil no encontrado")
	}

	// Eliminar archivo si no es el default
	if perfil.FotoPath != "" && perfil.FotoPath != "/static/uploads/perfil/default-avatar.png" {
		utils.DeleteFile(perfil.FotoPath)
	}

	// Restaurar foto por defecto
	return s.perfilRepo.UpdateFoto(userID, "/static/uploads/perfil/default-avatar.png", "", "", 0)
}
