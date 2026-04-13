package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"tu-proyecto/internal/models"

	"github.com/lib/pq"
)

// ============================================
// INTERFAZ PRINCIPAL (definida UNA SOLA VEZ)
// ============================================
type ModuloRepository interface {
	// CRUD Básico
	Create(modulo *models.Modulo) error
	Update(modulo *models.Modulo) error
	Delete(id int) error
	FindByID(id int) (*models.Modulo, error)
	FindAll(filter *models.ModuloFilter) (*models.ModuloPaginatedResponse, error)

	// Consultas específicas
	GetByCategoria(categoria string) ([]models.Modulo, error)
	HasAssociatedPermissions(id int) (bool, error)
	GetMaxOrdenByCategoria(categoria string) (int, error)
	ExistsByRuta(ruta string, excludeID int) (bool, error)
	UpdateOrder(id, orden int) error          // ← Agregado para Reordenar
	GetCategoriasDistinct() ([]string, error) // ← NUEVO MÉTODO
}

// ============================================
// IMPLEMENTACIÓN
// ============================================
type moduloRepository struct {
	db *sql.DB
}

func NewModuloRepository(db *sql.DB) ModuloRepository {
	return &moduloRepository{db: db}
}

// Create - Crear nuevo módulo
func (r *moduloRepository) Create(modulo *models.Modulo) error {
	query := `INSERT INTO modulos (nombre, nombre_mostrar, ruta, icono, categoria, orden, activo) 
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`

	err := r.db.QueryRow(query,
		modulo.Nombre,
		modulo.NombreMostrar,
		modulo.Ruta,
		modulo.Icono,
		modulo.Categoria,
		modulo.Orden,
		modulo.Activo,
	).Scan(&modulo.ID, &modulo.CreatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("ya existe un módulo con el nombre '%s'", modulo.Nombre)
		}
		return fmt.Errorf("error al crear módulo: %w", err)
	}
	return nil
}

// Update - Actualizar módulo existente
func (r *moduloRepository) Update(modulo *models.Modulo) error {
	query := `UPDATE modulos SET 
                nombre=$1, 
                nombre_mostrar=$2, 
                ruta=$3, 
                icono=$4, 
                categoria=$5, 
                orden=$6, 
                activo=$7 
              WHERE id=$8`

	result, err := r.db.Exec(query,
		modulo.Nombre,
		modulo.NombreMostrar,
		modulo.Ruta,
		modulo.Icono,
		modulo.Categoria,
		modulo.Orden,
		modulo.Activo,
		modulo.ID,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("ya existe un módulo con el nombre '%s'", modulo.Nombre)
		}
		return fmt.Errorf("error al actualizar módulo: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("módulo con ID %d no encontrado", modulo.ID)
	}
	return nil
}

// Delete - Eliminar módulo
func (r *moduloRepository) Delete(id int) error {
	query := `DELETE FROM modulos WHERE id=$1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar módulo: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("módulo con ID %d no encontrado", id)
	}
	return nil
}

// FindByID - Buscar módulo por ID
func (r *moduloRepository) FindByID(id int) (*models.Modulo, error) {
	modulo := &models.Modulo{}
	query := `SELECT id, nombre, nombre_mostrar, ruta, icono, categoria, orden, activo, created_at 
              FROM modulos WHERE id=$1`

	err := r.db.QueryRow(query, id).Scan(
		&modulo.ID,
		&modulo.Nombre,
		&modulo.NombreMostrar,
		&modulo.Ruta,
		&modulo.Icono,
		&modulo.Categoria,
		&modulo.Orden,
		&modulo.Activo,
		&modulo.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("módulo con ID %d no encontrado", id)
		}
		return nil, fmt.Errorf("error al buscar módulo: %w", err)
	}
	return modulo, nil
}

// FindAll - Listar módulos con paginación y filtros
func (r *moduloRepository) FindAll(filter *models.ModuloFilter) (*models.ModuloPaginatedResponse, error) {
	// Validar paginación
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	// Límite máximo de página (seguridad)
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	offset := (filter.Page - 1) * filter.PageSize

	// Construir filtros WHERE dinámicos
	where := []string{}
	args := []interface{}{}
	argPos := 1

	if filter.Nombre != "" {
		where = append(where, fmt.Sprintf("nombre ILIKE $%d", argPos))
		args = append(args, "%"+filter.Nombre+"%")
		argPos++
	}

	// Query para contar total
	countQuery := "SELECT COUNT(*) FROM modulos"
	if len(where) > 0 {
		countQuery += " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("error al contar módulos: %w", err)
	}

	// Query para datos paginados
	dataQuery := `SELECT id, nombre, nombre_mostrar, ruta, icono, categoria, orden, activo, created_at 
                  FROM modulos`
	if len(where) > 0 {
		dataQuery += " WHERE " + strings.Join(where, " AND ")
	}
	dataQuery += fmt.Sprintf(" ORDER BY orden ASC, id ASC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("error al listar módulos: %w", err)
	}
	defer rows.Close()

	var modulos []models.Modulo
	for rows.Next() {
		var m models.Modulo
		err := rows.Scan(
			&m.ID,
			&m.Nombre,
			&m.NombreMostrar,
			&m.Ruta,
			&m.Icono,
			&m.Categoria,
			&m.Orden,
			&m.Activo,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear módulo: %w", err)
		}
		modulos = append(modulos, m)
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize

	return &models.ModuloPaginatedResponse{
		Data:       modulos,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetByCategoria - Obtener módulos por categoría (solo activos)
func (r *moduloRepository) GetByCategoria(categoria string) ([]models.Modulo, error) {
	// Validar que la categoría no esté vacía
	if categoria == "" {
		return nil, fmt.Errorf("la categoría no puede estar vacía")
	}

	query := `SELECT id, nombre, nombre_mostrar, ruta, icono, categoria, orden, activo, created_at 
              FROM modulos 
              WHERE categoria = $1 AND activo = true 
              ORDER BY orden ASC`

	rows, err := r.db.Query(query, categoria)
	if err != nil {
		return nil, fmt.Errorf("error al obtener módulos por categoría: %w", err)
	}
	defer rows.Close()

	var modulos []models.Modulo
	for rows.Next() {
		var m models.Modulo
		err := rows.Scan(
			&m.ID,
			&m.Nombre,
			&m.NombreMostrar,
			&m.Ruta,
			&m.Icono,
			&m.Categoria,
			&m.Orden,
			&m.Activo,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear módulo: %w", err)
		}
		modulos = append(modulos, m)
	}

	return modulos, nil
}

// ============================================
// MÉTODOS ADICIONALES PARA VALIDACIONES
// ============================================

// HasAssociatedPermissions - Verifica si el módulo tiene permisos asignados
func (r *moduloRepository) HasAssociatedPermissions(id int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM permisos WHERE modulo_id = $1`
	err := r.db.QueryRow(query, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error verificando permisos asociados: %w", err)
	}
	return count > 0, nil
}

// GetMaxOrdenByCategoria - Obtiene el orden máximo para una categoría
func (r *moduloRepository) GetMaxOrdenByCategoria(categoria string) (int, error) {
	if categoria == "" {
		return 0, fmt.Errorf("la categoría no puede estar vacía")
	}

	var maxOrden int
	query := `SELECT COALESCE(MAX(orden), 0) FROM modulos WHERE categoria = $1`
	err := r.db.QueryRow(query, categoria).Scan(&maxOrden)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo máximo orden: %w", err)
	}
	return maxOrden, nil
}

// ExistsByRuta - Verifica si ya existe un módulo con la misma ruta
func (r *moduloRepository) ExistsByRuta(ruta string, excludeID int) (bool, error) {
	if ruta == "" {
		return false, fmt.Errorf("la ruta no puede estar vacía")
	}

	var count int
	var query string
	var err error

	if excludeID > 0 {
		// Excluir un ID específico (para updates)
		query = `SELECT COUNT(*) FROM modulos WHERE ruta = $1 AND id != $2`
		err = r.db.QueryRow(query, ruta, excludeID).Scan(&count)
	} else {
		// Sin exclusión (para creates)
		query = `SELECT COUNT(*) FROM modulos WHERE ruta = $1`
		err = r.db.QueryRow(query, ruta).Scan(&count)
	}

	if err != nil {
		return false, fmt.Errorf("error verificando ruta única: %w", err)
	}
	return count > 0, nil
}

// UpdateOrder - Actualiza el orden de un módulo específico
func (r *moduloRepository) UpdateOrder(id, orden int) error {
	if id <= 0 {
		return fmt.Errorf("ID inválido: %d", id)
	}
	if orden < 0 {
		return fmt.Errorf("orden inválido: %d", orden)
	}

	query := `UPDATE modulos SET orden = $1 WHERE id = $2`
	result, err := r.db.Exec(query, orden, id)
	if err != nil {
		return fmt.Errorf("error actualizando orden: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("módulo con ID %d no encontrado", id)
	}
	return nil
}

// Implementar el método
func (r *moduloRepository) GetCategoriasDistinct() ([]string, error) {
	query := `SELECT DISTINCT categoria FROM modulos ORDER BY categoria`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo categorías: %w", err)
	}
	defer rows.Close()

	var categorias []string
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			return nil, fmt.Errorf("error escaneando categoría: %w", err)
		}
		categorias = append(categorias, cat)
	}

	if len(categorias) == 0 {
		// Fallback a categorías por defecto
		return []string{"seguridad", "principal1", "principal2"}, nil
	}
	return categorias, nil
}
