package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"tu-proyecto/internal/models"

	"github.com/lib/pq"
)

// ============================================
// INTERFAZ PRINCIPAL
// ============================================
type ModuloRepository interface {
	// CRUD Básico
	Create(modulo *models.Modulo) error
	Update(modulo *models.Modulo) error
	Delete(id int) error
	FindByID(id int) (*models.Modulo, error)
	FindAll(filter *models.ModuloFilter) (*models.ModuloPaginatedResponse, error)

	// Métodos simplificados
	HasAssociatedPermissions(id int) (bool, error)
	ExistsByNombre(nombre string, excludeID int) (bool, error)
}

// ============================================
// IMPLEMENTACIÓN
// ============================================
type moduloRepository struct {
	db *sql.DB
}

func NewModuloRepository(db *sql.DB) ModuloRepository {
	log.Println("[DEBUG] 📦 ModuloRepository inicializado")
	return &moduloRepository{db: db}
}

// Create - Crear nuevo módulo
func (r *moduloRepository) Create(modulo *models.Modulo) error {
	log.Println("[DEBUG] 📦 Repository.Create - INICIANDO")
	log.Printf("[DEBUG] 📦 Datos recibidos: Nombre='%s', Descripcion='%s', Activo=%v",
		modulo.Nombre, modulo.Descripcion, modulo.Activo)

	query := `INSERT INTO modulos (nombre, descripcion, activo) 
              VALUES ($1, $2, $3) RETURNING id, created_at`

	log.Printf("[DEBUG] 📦 Ejecutando query: %s", query)
	log.Printf("[DEBUG] 📦 Parámetros: $1='%s', $2='%s', $3=%v",
		modulo.Nombre, modulo.Descripcion, modulo.Activo)

	err := r.db.QueryRow(query,
		modulo.Nombre,
		modulo.Descripcion,
		modulo.Activo,
	).Scan(&modulo.ID, &modulo.CreatedAt)

	if err != nil {
		log.Printf("[ERROR] 📦 Error en INSERT: %v", err)
		if pqErr, ok := err.(*pq.Error); ok {
			log.Printf("[ERROR] 📦 Código PostgreSQL: %s, Mensaje: %s", pqErr.Code, pqErr.Message)
			if pqErr.Code == "23505" {
				log.Printf("[WARN] 📦 Violación de UNIQUE constraint - Ya existe módulo con nombre '%s'", modulo.Nombre)
				return fmt.Errorf("ya existe un módulo con el nombre '%s'", modulo.Nombre)
			}
		}
		return fmt.Errorf("error al crear módulo: %w", err)
	}

	log.Printf("[DEBUG] 📦 ✅ Módulo creado exitosamente en BD")
	log.Printf("[DEBUG] 📦 ID asignado: %d, CreatedAt: %v", modulo.ID, modulo.CreatedAt)
	return nil
}

// Update - Actualizar módulo existente
func (r *moduloRepository) Update(modulo *models.Modulo) error {
	log.Println("[DEBUG] 📦 Repository.Update - INICIANDO")
	log.Printf("[DEBUG] 📦 Datos recibidos: ID=%d, Nombre='%s', Descripcion='%s', Activo=%v",
		modulo.ID, modulo.Nombre, modulo.Descripcion, modulo.Activo)

	query := `UPDATE modulos SET 
                nombre = $1, 
                descripcion = $2, 
                activo = $3 
              WHERE id = $4`

	log.Printf("[DEBUG] 📦 Ejecutando query: %s", query)
	log.Printf("[DEBUG] 📦 Parámetros: $1='%s', $2='%s', $3=%v, $4=%d",
		modulo.Nombre, modulo.Descripcion, modulo.Activo, modulo.ID)

	result, err := r.db.Exec(query,
		modulo.Nombre,
		modulo.Descripcion,
		modulo.Activo,
		modulo.ID,
	)

	if err != nil {
		log.Printf("[ERROR] 📦 Error en UPDATE: %v", err)
		if pqErr, ok := err.(*pq.Error); ok {
			log.Printf("[ERROR] 📦 Código PostgreSQL: %s, Mensaje: %s", pqErr.Code, pqErr.Message)
			if pqErr.Code == "23505" {
				log.Printf("[WARN] 📦 Violación de UNIQUE constraint - Ya existe módulo con nombre '%s'", modulo.Nombre)
				return fmt.Errorf("ya existe un módulo con el nombre '%s'", modulo.Nombre)
			}
		}
		return fmt.Errorf("error al actualizar módulo: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("[DEBUG] 📦 Filas afectadas: %d", rowsAffected)

	if rowsAffected == 0 {
		log.Printf("[WARN] 📦 No se encontró módulo con ID %d para actualizar", modulo.ID)
		return fmt.Errorf("módulo con ID %d no encontrado", modulo.ID)
	}

	log.Printf("[DEBUG] 📦 ✅ Módulo actualizado exitosamente")
	return nil
}

// Delete - Eliminar módulo
func (r *moduloRepository) Delete(id int) error {
	log.Println("[DEBUG] 📦 Repository.Delete - INICIANDO")
	log.Printf("[DEBUG] 📦 Eliminando módulo ID: %d", id)

	query := `DELETE FROM modulos WHERE id = $1`
	log.Printf("[DEBUG] 📦 Ejecutando query: %s con ID=%d", query, id)

	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Printf("[ERROR] 📦 Error en DELETE: %v", err)
		return fmt.Errorf("error al eliminar módulo: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("[DEBUG] 📦 Filas afectadas: %d", rowsAffected)

	if rowsAffected == 0 {
		log.Printf("[WARN] 📦 No se encontró módulo con ID %d para eliminar", id)
		return fmt.Errorf("módulo con ID %d no encontrado", id)
	}

	log.Printf("[DEBUG] 📦 ✅ Módulo ID=%d eliminado exitosamente", id)
	return nil
}

// FindByID - Buscar módulo por ID
func (r *moduloRepository) FindByID(id int) (*models.Modulo, error) {
	log.Println("[DEBUG] 📦 Repository.FindByID - INICIANDO")
	log.Printf("[DEBUG] 📦 Buscando módulo ID: %d", id)

	modulo := &models.Modulo{}
	query := `SELECT id, nombre, descripcion, activo, created_at 
              FROM modulos WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&modulo.ID,
		&modulo.Nombre,
		&modulo.Descripcion,
		&modulo.Activo,
		&modulo.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] 📦 Módulo con ID %d no encontrado", id)
			return nil, fmt.Errorf("módulo con ID %d no encontrado", id)
		}
		log.Printf("[ERROR] 📦 Error al buscar módulo: %v", err)
		return nil, fmt.Errorf("error al buscar módulo: %w", err)
	}

	log.Printf("[DEBUG] 📦 ✅ Módulo encontrado: ID=%d, Nombre='%s', Activo=%v",
		modulo.ID, modulo.Nombre, modulo.Activo)
	return modulo, nil
}

// FindAll - Listar módulos con paginación y filtros
func (r *moduloRepository) FindAll(filter *models.ModuloFilter) (*models.ModuloPaginatedResponse, error) {
	log.Println("[DEBUG] 📦 Repository.FindAll - INICIANDO")
	log.Printf("[DEBUG] 📦 Filtros: Nombre='%s', Page=%d, PageSize=%d",
		filter.Nombre, filter.Page, filter.PageSize)

	// Validar paginación
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	offset := (filter.Page - 1) * filter.PageSize

	log.Printf("[DEBUG] 📦 Paginación ajustada: Page=%d, PageSize=%d, Offset=%d",
		filter.Page, filter.PageSize, offset)

	// Construir filtros WHERE dinámicos
	where := []string{}
	args := []interface{}{}
	argPos := 1

	if filter.Nombre != "" {
		where = append(where, fmt.Sprintf("nombre ILIKE $%d", argPos))
		args = append(args, "%"+filter.Nombre+"%")
		argPos++
		log.Printf("[DEBUG] 📦 Filtro WHERE agregado: nombre ILIKE '%%%s%%'", filter.Nombre)
	}

	// Query para contar total
	countQuery := "SELECT COUNT(*) FROM modulos"
	if len(where) > 0 {
		countQuery += " WHERE " + strings.Join(where, " AND ")
	}

	log.Printf("[DEBUG] 📦 Count Query: %s", countQuery)
	log.Printf("[DEBUG] 📦 Count Args: %v", args)

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		log.Printf("[ERROR] 📦 Error al contar módulos: %v", err)
		return nil, fmt.Errorf("error al contar módulos: %w", err)
	}
	log.Printf("[DEBUG] 📦 Total módulos encontrados: %d", total)

	// Query para datos paginados
	dataQuery := `SELECT id, nombre, descripcion, activo, created_at 
                  FROM modulos`
	if len(where) > 0 {
		dataQuery += " WHERE " + strings.Join(where, " AND ")
	}
	dataQuery += fmt.Sprintf(" ORDER BY id ASC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.PageSize, offset)

	log.Printf("[DEBUG] 📦 Data Query: %s", dataQuery)
	log.Printf("[DEBUG] 📦 Data Args: %v", args)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		log.Printf("[ERROR] 📦 Error al listar módulos: %v", err)
		return nil, fmt.Errorf("error al listar módulos: %w", err)
	}
	defer rows.Close()

	var modulos []models.Modulo
	for rows.Next() {
		var m models.Modulo
		err := rows.Scan(
			&m.ID,
			&m.Nombre,
			&m.Descripcion,
			&m.Activo,
			&m.CreatedAt,
		)
		if err != nil {
			log.Printf("[ERROR] 📦 Error al escanear módulo: %v", err)
			return nil, fmt.Errorf("error al escanear módulo: %w", err)
		}
		modulos = append(modulos, m)
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize

	log.Printf("[DEBUG] 📦 ✅ Listado completado: %d módulos retornados, TotalPages=%d",
		len(modulos), totalPages)

	return &models.ModuloPaginatedResponse{
		Data:       modulos,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ============================================
// MÉTODOS ADICIONALES
// ============================================

// HasAssociatedPermissions - Verifica si el módulo tiene permisos asignados
// HasAssociatedPermissions - Verifica si el módulo tiene permisos ACTIVOS asignados
func (r *moduloRepository) HasAssociatedPermissions(id int) (bool, error) {
	log.Println("[DEBUG] 📦 Repository.HasAssociatedPermissions - INICIANDO")
	log.Printf("[DEBUG] 📦 Verificando permisos ACTIVOS asociados para módulo ID: %d", id)

	var count int
	query := `
		SELECT COUNT(*) FROM permisos 
		WHERE modulo_id = $1 
		AND (puede_ver = true OR puede_crear = true OR puede_editar = true 
		     OR puede_eliminar = true OR puede_detalle = true)
	`
	err := r.db.QueryRow(query, id).Scan(&count)
	if err != nil {
		log.Printf("[ERROR] 📦 Error verificando permisos asociados: %v", err)
		return false, fmt.Errorf("error verificando permisos asociados: %w", err)
	}

	log.Printf("[DEBUG] 📦 Permisos ACTIVOS encontrados: %d", count)
	return count > 0, nil
}

// ExistsByNombre - Verifica si ya existe un módulo con el mismo nombre
func (r *moduloRepository) ExistsByNombre(nombre string, excludeID int) (bool, error) {
	log.Println("[DEBUG] 📦 Repository.ExistsByNombre - INICIANDO")
	log.Printf("[DEBUG] 📦 Verificando nombre: '%s', Excluir ID: %d", nombre, excludeID)

	if nombre == "" {
		log.Printf("[WARN] 📦 Nombre vacío proporcionado")
		return false, fmt.Errorf("el nombre no puede estar vacío")
	}

	var count int
	var query string
	var err error

	if excludeID > 0 {
		query = `SELECT COUNT(*) FROM modulos WHERE nombre = $1 AND id != $2`
		err = r.db.QueryRow(query, nombre, excludeID).Scan(&count)
		log.Printf("[DEBUG] 📦 Query: %s [nombre='%s', excludeID=%d]", query, nombre, excludeID)
	} else {
		query = `SELECT COUNT(*) FROM modulos WHERE nombre = $1`
		err = r.db.QueryRow(query, nombre).Scan(&count)
		log.Printf("[DEBUG] 📦 Query: %s [nombre='%s']", query, nombre)
	}

	if err != nil {
		log.Printf("[ERROR] 📦 Error verificando nombre único: %v", err)
		return false, fmt.Errorf("error verificando nombre único: %w", err)
	}

	exists := count > 0
	log.Printf("[DEBUG] 📦 Resultado: existe=%v (count=%d)", exists, count)
	return exists, nil
}
