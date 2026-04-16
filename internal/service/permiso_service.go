package service

import (
	"fmt"
	"log"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/repository"
)

type PermisoService interface {
	// Obtener permisos de un perfil
	GetPermissionsByPerfil(perfilID int) (*models.PermisosPorPerfilResponse, error)

	// Guardar permisos para un perfil
	SavePermissions(perfilID int, permisos []models.PermisoItemRequest, auditorUserID int) error

	// Verificar permiso de usuario
	UserHasPermission(userID int, moduloNombre string, permiso string) (bool, error)
	GetUserPermissions(userID int) (map[string]map[string]bool, error)
	IsAdmin(userID int) (bool, error)

	// Nuevos métodos para auditoría
	GetUserRole(userID int) (int, string, error)
	GetUserPermissionsDetails(userID int) ([]models.ModuloConPermisos, error)
}

type permisoService struct {
	permisoRepo repository.PermisoRepository
	perfilRepo  repository.PerfilRepository
	moduloRepo  repository.ModuloRepository
}

func NewPermisoService(
	permisoRepo repository.PermisoRepository,
	perfilRepo repository.PerfilRepository,
	moduloRepo repository.ModuloRepository,
) PermisoService {
	return &permisoService{
		permisoRepo: permisoRepo,
		perfilRepo:  perfilRepo,
		moduloRepo:  moduloRepo,
	}
}

// GetPermissionsByPerfil - Obtiene todos los módulos con sus permisos para un perfil
func (s *permisoService) GetPermissionsByPerfil(perfilID int) (*models.PermisosPorPerfilResponse, error) {
	// Validar ID
	if perfilID <= 0 {
		return nil, fmt.Errorf("ID de perfil inválido: %d", perfilID)
	}

	// Verificar que el perfil existe
	perfil, err := s.perfilRepo.FindByID(perfilID)
	if err != nil {
		return nil, fmt.Errorf("perfil no encontrado: %w", err)
	}

	// Obtener permisos
	permisos, err := s.permisoRepo.GetPermissionsByPerfil(perfilID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo permisos: %w", err)
	}

	return &models.PermisosPorPerfilResponse{
		Perfil:   *perfil,
		Permisos: permisos,
		Total:    len(permisos),
	}, nil
}

// SavePermissions - Guarda los permisos para un perfil (CON AUDITORÍA)
func (s *permisoService) SavePermissions(perfilID int, permisos []models.PermisoItemRequest, auditorUserID int) error {
	// Validar ID
	if perfilID <= 0 {
		return fmt.Errorf("ID de perfil inválido: %d", perfilID)
	}

	// Verificar que el perfil existe
	perfil, err := s.perfilRepo.FindByID(perfilID)
	if err != nil {
		return fmt.Errorf("perfil no encontrado: %w", err)
	}

	// PREVENCIÓN: Evitar que un administrador se quite permisos a sí mismo
	// Si el perfil que se está editando es "Administrador" (ID=1)
	if perfilID == 1 {
		// Verificar quién está haciendo el cambio
		auditorRole, _, err := s.GetUserRole(auditorUserID)
		if err != nil {
			return fmt.Errorf("error verificando permisos del auditor: %w", err)
		}

		// Si el auditor NO es el perfil de administrador, bloquear
		if auditorRole != 1 {
			return fmt.Errorf("no se pueden modificar los permisos del perfil Administrador")
		}
	}

	// Verificar que los módulos existen
	for _, p := range permisos {
		_, err := s.moduloRepo.FindByID(p.ModuloID)
		if err != nil {
			return fmt.Errorf("módulo ID %d no encontrado", p.ModuloID)
		}
	}

	// Guardar permisos
	err = s.permisoRepo.AssignPermissions(perfilID, permisos)
	if err != nil {
		return fmt.Errorf("error guardando permisos: %w", err)
	}

	// AUDITORÍA: Registrar cambio de permisos
	log.Printf("[AUDIT] Usuario ID=%d modificó permisos del perfil ID=%d (%s). Total módulos: %d",
		auditorUserID, perfilID, perfil.Nombre, len(permisos))

	return nil
}

// UserHasPermission - Verifica si un usuario tiene un permiso específico
// UserHasPermission - Verifica si un usuario tiene un permiso específico
func (s *permisoService) UserHasPermission(userID int, moduloNombre string, permiso string) (bool, error) {
	if userID <= 0 {
		return false, fmt.Errorf("ID de usuario inválido")
	}

	// ✅ VERIFICAR SI ES ADMINISTRADOR PRIMERO
	isAdmin, err := s.permisoRepo.IsAdmin(userID)
	if err == nil && isAdmin {
		log.Printf("[DEBUG] UserHasPermission - Usuario %d es ADMIN, concediendo permiso '%s' en '%s'",
			userID, permiso, moduloNombre)
		return true, nil
	}

	// Si no es admin, verificar en BD
	return s.permisoRepo.UserHasPermission(userID, moduloNombre, permiso)
}

// GetUserPermissions - Obtiene todos los permisos de un usuario (para frontend)
// GetUserPermissions - Obtiene todos los permisos de un usuario (para frontend)
func (s *permisoService) GetUserPermissions(userID int) (map[string]map[string]bool, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("ID de usuario inválido")
	}

	// ✅ Si es Administrador, devolver todos los permisos en true
	isAdmin, err := s.permisoRepo.IsAdmin(userID)
	if err == nil && isAdmin {
		log.Printf("[DEBUG] GetUserPermissions - Usuario %d es ADMIN, devolviendo todos los permisos", userID)

		// Obtener todos los módulos
		filter := &models.ModuloFilter{Page: 1, PageSize: 100}
		modulos, err := s.moduloRepo.FindAll(filter)
		if err != nil {
			return nil, err
		}

		// Crear mapa con todos los permisos en true
		result := make(map[string]map[string]bool)
		for _, m := range modulos.Data {
			result[m.Nombre] = map[string]bool{
				"ver":      true,
				"crear":    true,
				"editar":   true,
				"eliminar": true,
				"detalle":  true,
			}
		}
		return result, nil
	}

	return s.permisoRepo.GetUserPermissions(userID)
}

// IsAdmin - Verifica si un usuario es administrador
func (s *permisoService) IsAdmin(userID int) (bool, error) {
	if userID <= 0 {
		return false, fmt.Errorf("ID de usuario inválido")
	}
	return s.permisoRepo.IsAdmin(userID)
}

// GetUserRole - Obtiene el rol de un usuario
func (s *permisoService) GetUserRole(userID int) (int, string, error) {
	if userID <= 0 {
		return 0, "", fmt.Errorf("ID de usuario inválido")
	}
	return s.permisoRepo.GetUserRole(userID)
}

// GetUserPermissionsDetails - Obtiene detalles completos de permisos de un usuario
func (s *permisoService) GetUserPermissionsDetails(userID int) ([]models.ModuloConPermisos, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("ID de usuario inválido")
	}
	return s.permisoRepo.GetUserPermissionsByUserID(userID)
}
