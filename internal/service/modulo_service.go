package service

import (
	"errors"
	"fmt"
	"regexp"
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
	GetByCategoria(categoria string) ([]models.Modulo, error)
	GetCategoriasDisponibles() ([]string, error)
	Reordenar(categoria string, ordenes []int) error
}

type moduloService struct {
	repo repository.ModuloRepository
}

func NewModuloService(repo repository.ModuloRepository) ModuloService {
	return &moduloService{repo: repo}
}

// GetCategoriasDisponibles - Obtiene categorías desde el repositorio (CORREGIDO)
func (s *moduloService) GetCategoriasDisponibles() ([]string, error) {
	categorias, err := s.repo.GetCategoriasDistinct()
	if err != nil {
		// Fallback a categorías por defecto si hay error de conexión
		return []string{"seguridad", "principal1", "principal2"}, nil
	}
	return categorias, nil
}

// validateRuta - Validación de formato de ruta
func validateRuta(ruta string) error {
	if !strings.HasPrefix(ruta, "/") {
		return errors.New("la ruta debe comenzar con '/'")
	}
	if strings.Contains(ruta, "//") {
		return errors.New("la ruta no puede tener '//' consecutivos")
	}
	if strings.HasSuffix(ruta, "/") && ruta != "/" {
		return errors.New("la ruta no puede terminar con '/'")
	}

	validPath := regexp.MustCompile(`^/[a-zA-Z0-9/_.-]+$`)
	if !validPath.MatchString(ruta) {
		return errors.New("la ruta solo puede contener letras, números, /, _, ., y -")
	}
	return nil
}

func (s *moduloService) Create(modulo *models.Modulo) error {
	modulo.Nombre = strings.TrimSpace(modulo.Nombre)
	modulo.NombreMostrar = strings.TrimSpace(modulo.NombreMostrar)
	modulo.Ruta = strings.TrimSpace(modulo.Ruta)

	if modulo.Nombre == "" {
		return errors.New("el nombre del módulo es requerido")
	}
	if modulo.NombreMostrar == "" {
		return errors.New("el nombre para mostrar es requerido")
	}
	if modulo.Ruta == "" {
		return errors.New("la ruta del módulo es requerida")
	}
	if len(modulo.Nombre) < 3 {
		return errors.New("el nombre debe tener al menos 3 caracteres")
	}
	if len(modulo.Nombre) > 100 {
		return errors.New("el nombre no puede exceder 100 caracteres")
	}

	if err := validateRuta(modulo.Ruta); err != nil {
		return err
	}

	categorias, err := s.GetCategoriasDisponibles()
	if err != nil {
		return fmt.Errorf("error obteniendo categorías: %w", err)
	}

	categoriaValida := false
	for _, cat := range categorias {
		if cat == modulo.Categoria {
			categoriaValida = true
			break
		}
	}
	if !categoriaValida {
		return fmt.Errorf("categoría no válida. Use: %s", strings.Join(categorias, ", "))
	}

	if modulo.Orden < 0 {
		return errors.New("el orden no puede ser negativo")
	}

	exists, err := s.repo.ExistsByRuta(modulo.Ruta, 0)
	if err != nil {
		return fmt.Errorf("error validando ruta: %w", err)
	}
	if exists {
		return errors.New("ya existe un módulo con esta ruta")
	}

	if modulo.Orden == 0 {
		maxOrden, err := s.repo.GetMaxOrdenByCategoria(modulo.Categoria)
		if err != nil {
			return fmt.Errorf("error obteniendo orden: %w", err)
		}
		modulo.Orden = maxOrden + 1
	}

	return s.repo.Create(modulo)
}

func (s *moduloService) Update(modulo *models.Modulo) error {
	existing, err := s.repo.FindByID(modulo.ID)
	if err != nil {
		return fmt.Errorf("módulo no encontrado: %w", err)
	}

	modulo.Nombre = strings.TrimSpace(modulo.Nombre)
	modulo.NombreMostrar = strings.TrimSpace(modulo.NombreMostrar)
	modulo.Ruta = strings.TrimSpace(modulo.Ruta)

	if modulo.Nombre == "" {
		return errors.New("el nombre del módulo es requerido")
	}
	if modulo.NombreMostrar == "" {
		return errors.New("el nombre para mostrar es requerido")
	}
	if modulo.Ruta == "" {
		return errors.New("la ruta del módulo es requerida")
	}
	if len(modulo.Nombre) < 3 {
		return errors.New("el nombre debe tener al menos 3 caracteres")
	}

	if err := validateRuta(modulo.Ruta); err != nil {
		return err
	}

	categorias, err := s.GetCategoriasDisponibles()
	if err != nil {
		return fmt.Errorf("error obteniendo categorías: %w", err)
	}

	categoriaValida := false
	for _, cat := range categorias {
		if cat == modulo.Categoria {
			categoriaValida = true
			break
		}
	}
	if !categoriaValida {
		return fmt.Errorf("categoría no válida. Use: %s", strings.Join(categorias, ", "))
	}

	if modulo.Orden < 0 {
		return errors.New("el orden no puede ser negativo")
	}

	exists, err := s.repo.ExistsByRuta(modulo.Ruta, modulo.ID)
	if err != nil {
		return fmt.Errorf("error validando ruta: %w", err)
	}
	if exists {
		return errors.New("ya existe un módulo con esta ruta")
	}

	modulo.CreatedAt = existing.CreatedAt
	return s.repo.Update(modulo)
}

func (s *moduloService) Delete(id int) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("módulo no encontrado: %w", err)
	}

	hasPerms, err := s.repo.HasAssociatedPermissions(id)
	if err != nil {
		return fmt.Errorf("error verificando permisos asociados: %w", err)
	}
	if hasPerms {
		return errors.New("no se puede eliminar el módulo porque tiene permisos asignados")
	}

	return s.repo.Delete(id)
}

func (s *moduloService) Reordenar(categoria string, ordenes []int) error {
	if categoria == "" {
		return errors.New("la categoría es requerida")
	}

	if len(ordenes) == 0 {
		return errors.New("no hay órdenes para actualizar")
	}

	modulos, err := s.repo.GetByCategoria(categoria)
	if err != nil {
		return fmt.Errorf("error obteniendo módulos: %w", err)
	}

	if len(modulos) != len(ordenes) {
		return fmt.Errorf("el número de órdenes (%d) no coincide con los módulos de la categoría (%d)",
			len(ordenes), len(modulos))
	}

	validIDs := make(map[int]bool)
	for _, m := range modulos {
		validIDs[m.ID] = true
	}

	for _, moduloID := range ordenes {
		if !validIDs[moduloID] {
			return fmt.Errorf("el módulo ID %d no pertenece a la categoría '%s' o no existe", moduloID, categoria)
		}
	}

	for i, moduloID := range ordenes {
		if err := s.repo.UpdateOrder(moduloID, i+1); err != nil {
			return fmt.Errorf("error actualizando orden para módulo %d: %w", moduloID, err)
		}
	}

	return nil
}

func (s *moduloService) GetByID(id int) (*models.Modulo, error) {
	if id <= 0 {
		return nil, errors.New("ID inválido")
	}
	return s.repo.FindByID(id)
}

func (s *moduloService) GetAll(filter *models.ModuloFilter) (*models.ModuloPaginatedResponse, error) {
	if filter == nil {
		filter = &models.ModuloFilter{Page: 1, PageSize: 10}
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 10
	}
	return s.repo.FindAll(filter)
}

func (s *moduloService) GetByCategoria(categoria string) ([]models.Modulo, error) {
	if categoria == "" {
		return nil, errors.New("la categoría es requerida")
	}
	return s.repo.GetByCategoria(categoria)
}
