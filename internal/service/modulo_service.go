package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"tu-proyecto/internal/models"
	"tu-proyecto/internal/repository"
)

type ModuloService interface {
	Create(modulo *models.Modulo) error
	Update(modulo *models.Modulo) error
	Delete(id int) error
	GetByID(id int) (*models.Modulo, error)
	GetAll(filter *models.ModuloFilter) (*models.ModuloPaginatedResponse, error)
}

type moduloService struct {
	repo repository.ModuloRepository
}

func NewModuloService(repo repository.ModuloRepository) ModuloService {
	log.Println("[DEBUG] 🔧 ModuloService inicializado")
	return &moduloService{repo: repo}
}

// Create - Crear un nuevo módulo
// Create - Crear un nuevo módulo
func (s *moduloService) Create(modulo *models.Modulo) error {
	log.Println("[DEBUG] 🔧 Service.Create - INICIANDO")
	log.Printf("[DEBUG] 🔧 Datos recibidos: Nombre='%s', Descripcion='%s', Activo=%v",
		modulo.Nombre, modulo.Descripcion, modulo.Activo)

	modulo.Nombre = strings.TrimSpace(modulo.Nombre)
	modulo.Descripcion = strings.TrimSpace(modulo.Descripcion)

	if modulo.Nombre == "" {
		return errors.New("el nombre del módulo es requerido")
	}
	if len(modulo.Nombre) < 3 {
		return errors.New("el nombre debe tener al menos 3 caracteres")
	}
	if len(modulo.Nombre) > 100 {
		return errors.New("el nombre no puede exceder 100 caracteres")
	}

	exists, err := s.repo.ExistsByNombre(modulo.Nombre, 0)
	if err != nil {
		return fmt.Errorf("error verificando nombre: %w", err)
	}
	if exists {
		return fmt.Errorf("ya existe un módulo con el nombre '%s'", modulo.Nombre)
	}

	// ✅ ELIMINAR ESTA LÍNEA - Ya viene del handler correctamente
	// modulo.Activo = true

	log.Printf("[DEBUG] 🔧 Valor de Activo que se enviará a BD: %v", modulo.Activo)

	if err := s.repo.Create(modulo); err != nil {
		log.Printf("[ERROR] 🔧 Error en repo.Create: %v", err)
		return err
	}

	log.Printf("[DEBUG] 🔧 ✅ Módulo creado exitosamente - ID asignado: %d", modulo.ID)
	return nil
}

// Update - Actualizar un módulo existente
func (s *moduloService) Update(modulo *models.Modulo) error {
	log.Println("[DEBUG] 🔧 Service.Update - INICIANDO")
	log.Printf("[DEBUG] 🔧 Datos recibidos: ID=%d, Nombre='%s', Descripcion='%s', Activo=%v",
		modulo.ID, modulo.Nombre, modulo.Descripcion, modulo.Activo)

	// Verificar que existe
	log.Printf("[DEBUG] 🔧 Verificando existencia de módulo ID=%d...", modulo.ID)
	existing, err := s.repo.FindByID(modulo.ID)
	if err != nil {
		log.Printf("[ERROR] 🔧 Módulo ID=%d no encontrado: %v", modulo.ID, err)
		return fmt.Errorf("módulo no encontrado: %w", err)
	}
	log.Printf("[DEBUG] 🔧 ✅ Módulo existente encontrado: Nombre='%s'", existing.Nombre)

	// Limpiar campos
	modulo.Nombre = strings.TrimSpace(modulo.Nombre)
	modulo.Descripcion = strings.TrimSpace(modulo.Descripcion)
	log.Printf("[DEBUG] 🔧 Después de limpiar: Nombre='%s', Descripcion='%s'",
		modulo.Nombre, modulo.Descripcion)

	// Validaciones
	log.Println("[DEBUG] 🔧 Ejecutando validaciones...")
	if modulo.Nombre == "" {
		log.Printf("[WARN] 🔧 ⚠️ Validación fallida: nombre vacío")
		return errors.New("el nombre del módulo es requerido")
	}
	if len(modulo.Nombre) < 3 {
		log.Printf("[WARN] 🔧 ⚠️ Validación fallida: nombre muy corto")
		return errors.New("el nombre debe tener al menos 3 caracteres")
	}
	if len(modulo.Nombre) > 100 {
		log.Printf("[WARN] 🔧 ⚠️ Validación fallida: nombre muy largo")
		return errors.New("el nombre no puede exceder 100 caracteres")
	}
	log.Println("[DEBUG] 🔧 ✅ Validaciones pasadas")

	// Verificar nombre único (excluyendo el ID actual)
	log.Printf("[DEBUG] 🔧 Verificando nombre único '%s' (excluyendo ID=%d)...", modulo.Nombre, modulo.ID)
	exists, err := s.repo.ExistsByNombre(modulo.Nombre, modulo.ID)
	if err != nil {
		log.Printf("[ERROR] 🔧 Error verificando nombre único: %v", err)
		return fmt.Errorf("error verificando nombre: %w", err)
	}
	if exists {
		log.Printf("[WARN] 🔧 ⚠️ Ya existe otro módulo con el nombre '%s'", modulo.Nombre)
		return fmt.Errorf("ya existe un módulo con el nombre '%s'", modulo.Nombre)
	}
	log.Println("[DEBUG] 🔧 ✅ Nombre único verificado")

	// Mantener fecha de creación original
	modulo.CreatedAt = existing.CreatedAt
	log.Printf("[DEBUG] 🔧 Manteniendo CreatedAt original: %v", modulo.CreatedAt)

	log.Printf("[DEBUG] 🔧 Datos finales a actualizar: ID=%d, Nombre='%s', Activo=%v",
		modulo.ID, modulo.Nombre, modulo.Activo)

	// Llamar al repositorio
	log.Println("[DEBUG] 🔧 Llamando a repo.Update()...")
	if err := s.repo.Update(modulo); err != nil {
		log.Printf("[ERROR] 🔧 Error en repo.Update: %v", err)
		return err
	}

	log.Printf("[DEBUG] 🔧 ✅ Módulo ID=%d actualizado exitosamente", modulo.ID)
	return nil
}

// Delete - Eliminar un módulo
func (s *moduloService) Delete(id int) error {
	log.Println("[DEBUG] 🔧 Service.Delete - INICIANDO")
	log.Printf("[DEBUG] 🔧 Eliminando módulo ID: %d", id)

	// Verificar que existe
	log.Printf("[DEBUG] 🔧 Verificando existencia de módulo ID=%d...", id)
	modulo, err := s.repo.FindByID(id)
	if err != nil {
		log.Printf("[ERROR] 🔧 Módulo ID=%d no encontrado: %v", id, err)
		return fmt.Errorf("módulo no encontrado: %w", err)
	}
	log.Printf("[DEBUG] 🔧 ✅ Módulo encontrado: '%s'", modulo.Nombre)

	// Verificar si tiene permisos asociados
	log.Printf("[DEBUG] 🔧 Verificando permisos asociados...")
	hasPerms, err := s.repo.HasAssociatedPermissions(id)
	if err != nil {
		log.Printf("[ERROR] 🔧 Error verificando permisos asociados: %v", err)
		return fmt.Errorf("error verificando permisos asociados: %w", err)
	}
	if hasPerms {
		log.Printf("[WARN] 🔧 ⚠️ Módulo ID=%d tiene permisos asignados, no se puede eliminar", id)
		return errors.New("no se puede eliminar el módulo porque tiene permisos asignados")
	}
	log.Println("[DEBUG] 🔧 ✅ Sin permisos asociados")

	// Llamar al repositorio
	log.Println("[DEBUG] 🔧 Llamando a repo.Delete()...")
	if err := s.repo.Delete(id); err != nil {
		log.Printf("[ERROR] 🔧 Error en repo.Delete: %v", err)
		return err
	}

	log.Printf("[DEBUG] 🔧 ✅ Módulo ID=%d eliminado exitosamente", id)
	return nil
}

// GetByID - Obtener un módulo por ID
func (s *moduloService) GetByID(id int) (*models.Modulo, error) {
	log.Println("[DEBUG] 🔧 Service.GetByID - INICIANDO")
	log.Printf("[DEBUG] 🔧 Buscando módulo ID: %d", id)

	if id <= 0 {
		log.Printf("[WARN] 🔧 ⚠️ ID inválido: %d", id)
		return nil, errors.New("ID inválido")
	}

	modulo, err := s.repo.FindByID(id)
	if err != nil {
		log.Printf("[ERROR] 🔧 Error al buscar módulo ID=%d: %v", id, err)
		return nil, err
	}

	log.Printf("[DEBUG] 🔧 ✅ Módulo encontrado: ID=%d, Nombre='%s'", modulo.ID, modulo.Nombre)
	return modulo, nil
}

// GetAll - Listar módulos con paginación y filtros
func (s *moduloService) GetAll(filter *models.ModuloFilter) (*models.ModuloPaginatedResponse, error) {
	log.Println("[DEBUG] 🔧 Service.GetAll - INICIANDO")

	if filter == nil {
		filter = &models.ModuloFilter{Page: 1, PageSize: 10}
		log.Println("[DEBUG] 🔧 Filter nil, usando valores por defecto")
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 10
		log.Printf("[DEBUG] 🔧 PageSize ajustado a: %d", filter.PageSize)
	}
	if filter.Page < 1 {
		filter.Page = 1
		log.Printf("[DEBUG] 🔧 Page ajustado a: %d", filter.Page)
	}

	log.Printf("[DEBUG] 🔧 Filtros aplicados: Nombre='%s', Page=%d, PageSize=%d",
		filter.Nombre, filter.Page, filter.PageSize)

	log.Println("[DEBUG] 🔧 Llamando a repo.FindAll()...")
	result, err := s.repo.FindAll(filter)
	if err != nil {
		log.Printf("[ERROR] 🔧 Error en repo.FindAll: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] 🔧 ✅ Listado completado: %d módulos encontrados (Total: %d, Página %d/%d)",
		len(result.Data), result.Total, result.Page, result.TotalPages)

	return result, nil
}
