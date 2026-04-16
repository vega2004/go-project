package service

import (
	"errors"
	"fmt"
	"log" // ← Agregar import
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
	log.Println("[DEBUG] PerfilService.Create - Iniciando")
	log.Printf("[DEBUG] PerfilService.Create - Nombre original: '%s'", perfil.Nombre)
	log.Printf("[DEBUG] PerfilService.Create - Descripción original: '%s'", perfil.Descripcion)

	perfil.Nombre = strings.TrimSpace(perfil.Nombre)
	perfil.Descripcion = strings.TrimSpace(perfil.Descripcion)

	log.Printf("[DEBUG] PerfilService.Create - Nombre después de trim: '%s'", perfil.Nombre)
	log.Printf("[DEBUG] PerfilService.Create - Descripción después de trim: '%s'", perfil.Descripcion)

	if perfil.Nombre == "" {
		log.Printf("[DEBUG] PerfilService.Create - Error: nombre vacío")
		return errors.New("el nombre del perfil es requerido")
	}
	if len(perfil.Nombre) < 3 {
		log.Printf("[DEBUG] PerfilService.Create - Error: nombre muy corto (%d caracteres)", len(perfil.Nombre))
		return errors.New("el nombre debe tener al menos 3 caracteres")
	}
	if len(perfil.Nombre) > 50 {
		log.Printf("[DEBUG] PerfilService.Create - Error: nombre muy largo (%d caracteres)", len(perfil.Nombre))
		return errors.New("el nombre no puede exceder 50 caracteres")
	}

	log.Printf("[DEBUG] PerfilService.Create - Llamando a repositorio.Create para nombre: '%s'", perfil.Nombre)
	err := s.perfilRepo.Create(perfil)
	if err != nil {
		log.Printf("[ERROR] PerfilService.Create - Error del repositorio: %v", err)
		return err
	}

	log.Printf("[DEBUG] PerfilService.Create - Perfil creado exitosamente con ID: %d", perfil.ID)
	return nil
}

// Update - Actualiza un perfil/rol existente
func (s *perfilService) Update(perfil *models.Perfil) error {
	log.Println("[DEBUG] PerfilService.Update - Iniciando")
	log.Printf("[DEBUG] PerfilService.Update - ID: %d, Nombre original: '%s'", perfil.ID, perfil.Nombre)

	// Verificar que existe
	_, err := s.perfilRepo.FindByID(perfil.ID)
	if err != nil {
		log.Printf("[ERROR] PerfilService.Update - Perfil no encontrado: %v", err)
		return fmt.Errorf("perfil no encontrado: %w", err)
	}

	perfil.Nombre = strings.TrimSpace(perfil.Nombre)
	perfil.Descripcion = strings.TrimSpace(perfil.Descripcion)

	log.Printf("[DEBUG] PerfilService.Update - Nombre después de trim: '%s'", perfil.Nombre)

	if perfil.Nombre == "" {
		return errors.New("el nombre del perfil es requerido")
	}
	if len(perfil.Nombre) < 3 {
		return errors.New("el nombre debe tener al menos 3 caracteres")
	}
	if len(perfil.Nombre) > 50 {
		return errors.New("el nombre no puede exceder 50 caracteres")
	}

	log.Printf("[DEBUG] PerfilService.Update - Llamando a repositorio.Update")
	return s.perfilRepo.Update(perfil)
}

// Delete - Elimina un perfil/rol (con verificación de uso)
func (s *perfilService) Delete(id int) error {
	log.Printf("[DEBUG] PerfilService.Delete - Iniciando para ID: %d", id)

	// Verificar que existe
	_, err := s.perfilRepo.FindByID(id)
	if err != nil {
		log.Printf("[ERROR] PerfilService.Delete - Perfil no encontrado: %v", err)
		return fmt.Errorf("perfil no encontrado: %w", err)
	}

	// No permitir eliminar perfil de Administrador
	if id == 1 {
		log.Printf("[DEBUG] PerfilService.Delete - Intento de eliminar Administrador bloqueado")
		return errors.New("no se puede eliminar el perfil de Administrador")
	}

	// Verificar si está en uso
	inUse, err := s.perfilRepo.IsInUse(id)
	if err != nil {
		log.Printf("[ERROR] PerfilService.Delete - Error verificando uso: %v", err)
		return fmt.Errorf("error verificando uso del perfil: %w", err)
	}
	if inUse {
		log.Printf("[DEBUG] PerfilService.Delete - Perfil en uso, no se puede eliminar")
		return errors.New("no se puede eliminar el perfil porque tiene usuarios asignados")
	}

	log.Printf("[DEBUG] PerfilService.Delete - Llamando a repositorio.Delete")
	return s.perfilRepo.Delete(id)
}

// GetByID - Obtiene un perfil/rol por ID
func (s *perfilService) GetByID(id int) (*models.Perfil, error) {
	log.Printf("[DEBUG] PerfilService.GetByID - Buscando ID: %d", id)
	if id <= 0 {
		return nil, errors.New("ID inválido")
	}
	return s.perfilRepo.FindByID(id)
}

// GetAll - Lista perfiles/roles con paginación
func (s *perfilService) GetAll(filter *models.PerfilFilter) (*models.PerfilPaginatedResponse, error) {
	log.Printf("[DEBUG] PerfilService.GetAll - Filtro: Nombre='%s', Page=%d, PageSize=%d",
		filter.Nombre, filter.Page, filter.PageSize)

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
	log.Printf("[DEBUG] PerfilService.GetPerfil - Buscando perfil para userID: %d", userID)
	if userID <= 0 {
		return nil, errors.New("ID de usuario inválido")
	}
	return s.perfilRepo.GetPerfilByUserID(userID)
}

// UpdatePerfil - Actualiza los datos del perfil de usuario
func (s *perfilService) UpdatePerfil(userID int, perfil *models.PerfilUsuario) error {
	log.Printf("[DEBUG] PerfilService.UpdatePerfil - Actualizando perfil para userID: %d", userID)

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

	log.Printf("[DEBUG] PerfilService.UpdatePerfil - Bio: '%s', Direccion: '%s'", perfil.Bio, perfil.Direccion)

	return s.perfilRepo.UpdatePerfil(userID, perfil)
}

// UpdateFoto - Actualiza la foto de perfil del usuario
func (s *perfilService) UpdateFoto(userID int, file multipart.File, header *multipart.FileHeader) (string, error) {
	log.Printf("[DEBUG] PerfilService.UpdateFoto - Subiendo foto para userID: %d", userID)
	log.Printf("[DEBUG] PerfilService.UpdateFoto - Archivo: %s, Tamaño: %d bytes", header.Filename, header.Size)

	if userID <= 0 {
		return "", errors.New("ID de usuario inválido")
	}
	if file == nil || header == nil {
		return "", errors.New("archivo de imagen inválido")
	}

	// Validar y guardar imagen
	log.Printf("[DEBUG] PerfilService.UpdateFoto - Llamando a SaveProfileImage")
	uploadedFile, err := utils.SaveProfileImage(file, header)
	if err != nil {
		log.Printf("[ERROR] PerfilService.UpdateFoto - Error guardando imagen: %v", err)
		return "", fmt.Errorf("error guardando imagen: %w", err)
	}

	log.Printf("[DEBUG] PerfilService.UpdateFoto - Imagen guardada en: %s", uploadedFile.Path)

	// Obtener perfil actual para eliminar foto anterior
	perfilActual, _ := s.perfilRepo.GetPerfilByUserID(userID)
	if perfilActual != nil && perfilActual.FotoPath != "" && perfilActual.FotoPath != "/static/uploads/perfil/default-avatar.png" {
		log.Printf("[DEBUG] PerfilService.UpdateFoto - Eliminando foto anterior: %s", perfilActual.FotoPath)
		utils.DeleteFile(perfilActual.FotoPath)
	}

	// Actualizar en BD
	log.Printf("[DEBUG] PerfilService.UpdateFoto - Actualizando BD con nueva foto")
	err = s.perfilRepo.UpdateFoto(userID, uploadedFile.Path, uploadedFile.OriginalName, uploadedFile.MimeType, int(uploadedFile.Size))
	if err != nil {
		log.Printf("[ERROR] PerfilService.UpdateFoto - Error actualizando BD: %v", err)
		return "", fmt.Errorf("error actualizando foto en BD: %w", err)
	}

	log.Printf("[DEBUG] PerfilService.UpdateFoto - Foto actualizada exitosamente")
	return uploadedFile.Path, nil
}

// ChangePassword - Cambia la contraseña del usuario
func (s *perfilService) ChangePassword(userID int, form *models.CambioPassword) error {
	log.Printf("[DEBUG] PerfilService.ChangePassword - Cambiando contraseña para userID: %d", userID)

	if userID <= 0 {
		return errors.New("ID de usuario inválido")
	}
	if form == nil {
		return errors.New("datos de cambio de contraseña inválidos")
	}

	// Validar que las contraseñas coincidan
	if form.Nueva != form.Confirmar {
		log.Printf("[DEBUG] PerfilService.ChangePassword - Las contraseñas no coinciden")
		return errors.New("las contraseñas no coinciden")
	}

	// Validar longitud mínima
	if len(form.Nueva) < 6 {
		log.Printf("[DEBUG] PerfilService.ChangePassword - Contraseña muy corta: %d caracteres", len(form.Nueva))
		return errors.New("la nueva contraseña debe tener al menos 6 caracteres")
	}

	// Obtener usuario actual
	user, err := s.authRepo.FindByID(userID)
	if err != nil {
		log.Printf("[ERROR] PerfilService.ChangePassword - Usuario no encontrado: %v", err)
		return errors.New("usuario no encontrado")
	}

	// Validar contraseña actual
	if !s.authRepo.ValidatePassword(user.Password, form.Actual) {
		log.Printf("[DEBUG] PerfilService.ChangePassword - Contraseña actual incorrecta")
		return errors.New("contraseña actual incorrecta")
	}

	// Actualizar contraseña
	log.Printf("[DEBUG] PerfilService.ChangePassword - Actualizando contraseña en BD")
	return s.authRepo.UpdatePassword(userID, form.Nueva)
}

// DeleteFoto - Elimina la foto de perfil (restaura la default)
func (s *perfilService) DeleteFoto(userID int) error {
	log.Printf("[DEBUG] PerfilService.DeleteFoto - Eliminando foto para userID: %d", userID)

	if userID <= 0 {
		return errors.New("ID de usuario inválido")
	}

	// Obtener perfil actual
	perfil, err := s.perfilRepo.GetPerfilByUserID(userID)
	if err != nil {
		log.Printf("[ERROR] PerfilService.DeleteFoto - Perfil no encontrado: %v", err)
		return errors.New("perfil no encontrado")
	}

	// Eliminar archivo si no es el default
	if perfil.FotoPath != "" && perfil.FotoPath != "/static/uploads/perfil/default-avatar.png" {
		log.Printf("[DEBUG] PerfilService.DeleteFoto - Eliminando archivo: %s", perfil.FotoPath)
		utils.DeleteFile(perfil.FotoPath)
	}

	// Restaurar foto por defecto
	log.Printf("[DEBUG] PerfilService.DeleteFoto - Restaurando foto por defecto")
	return s.perfilRepo.UpdateFoto(userID, "/static/uploads/perfil/default-avatar.png", "", "", 0)
}
