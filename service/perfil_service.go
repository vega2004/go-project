package service

import (
	"errors"
	"mime/multipart"
	"time"
	"tu-proyecto/model"
	"tu-proyecto/repository"
	"tu-proyecto/utils"
)

type PerfilService interface {
	// Métodos básicos
	GetPerfil(userID int) (*model.Perfil, error)
	UpdatePerfil(userID int, perfil *model.Perfil) error

	// Métodos para foto
	UpdateFoto(userID int, file multipart.File, header *multipart.FileHeader) (string, error)
	DeleteFoto(userID int) error

	// Método para contraseña
	ChangePassword(userID int, form *model.CambioPassword) error
}

type perfilService struct {
	perfilRepo repository.PerfilRepository
	authRepo   repository.AuthRepository
}

func NewPerfilService(perfilRepo repository.PerfilRepository, authRepo repository.AuthRepository) PerfilService {
	return &perfilService{
		perfilRepo: perfilRepo,
		authRepo:   authRepo,
	}
}

// ============================================
// MÉTODOS DE PERFIL
// ============================================

// GetPerfil - Obtiene el perfil de un usuario
func (s *perfilService) GetPerfil(userID int) (*model.Perfil, error) {
	return s.perfilRepo.GetPerfil(userID)
}

// UpdatePerfil - Actualiza datos del perfil
func (s *perfilService) UpdatePerfil(userID int, perfil *model.Perfil) error {
	perfil.UserID = userID
	perfil.UpdatedAt = time.Now()
	return s.perfilRepo.UpdatePerfil(perfil)
}

// ============================================
// MÉTODOS PARA FOTO DE PERFIL
// ============================================

// UpdateFoto - Actualiza la foto de perfil
func (s *perfilService) UpdateFoto(userID int, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validar y guardar imagen usando utils.Upload
	uploadedFile, err := utils.SaveProfileImage(file, header)
	if err != nil {
		return "", err
	}

	// Obtener perfil actual para eliminar foto anterior
	perfil, _ := s.perfilRepo.GetPerfil(userID)
	if perfil != nil && perfil.Foto != "" && perfil.Foto != "/static/uploads/perfil/default-avatar.png" {
		// Eliminar archivo anterior
		utils.DeleteFile(perfil.Foto)
	}

	// Actualizar en BD
	err = s.perfilRepo.UpdateFoto(userID, uploadedFile.Path)
	if err != nil {
		return "", err
	}

	return uploadedFile.Path, nil
}

// DeleteFoto - Elimina la foto de perfil (restaura default)
func (s *perfilService) DeleteFoto(userID int) error {
	// Obtener perfil actual
	perfil, err := s.perfilRepo.GetPerfil(userID)
	if err != nil {
		return errors.New("perfil no encontrado")
	}

	// Eliminar archivo si no es el default
	if perfil.Foto != "" && perfil.Foto != "/static/uploads/perfil/default-avatar.png" {
		if err := utils.DeleteFile(perfil.Foto); err != nil {
			return err
		}
	}

	// Restaurar foto por defecto en BD
	defaultFoto := "/static/uploads/perfil/default-avatar.png"
	return s.perfilRepo.UpdateFoto(userID, defaultFoto)
}

// ============================================
// MÉTODO PARA CAMBIO DE CONTRASEÑA
// ============================================

// ChangePassword - Cambia la contraseña del usuario
func (s *perfilService) ChangePassword(userID int, form *model.CambioPassword) error {
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
