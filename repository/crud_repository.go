package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"tu-proyecto/model"
)

type CrudRepository interface {
	Create(persona *model.Persona) error
	Update(persona *model.Persona) error
	Delete(id int) error
	FindByID(id int) (*model.Persona, error)
	FindAll(filter *model.PersonaFilter) (*model.PaginatedResponse, error)
}

type crudRepository struct {
	db *sql.DB
}

func NewCrudRepository(db *sql.DB) CrudRepository {
	return &crudRepository{db: db}
}

func (r *crudRepository) Create(persona *model.Persona) error {
	query := `INSERT INTO personas (nombre, estado_civil, user_id, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $4) RETURNING id`

	err := r.db.QueryRow(query,
		persona.Nombre,
		persona.EstadoCivil,
		persona.UserID,
		persona.CreatedAt).Scan(&persona.ID)

	return err
}

func (r *crudRepository) Update(persona *model.Persona) error {
	query := `UPDATE personas SET nombre=$1, estado_civil=$2, updated_at=$3 WHERE id=$4`

	_, err := r.db.Exec(query,
		persona.Nombre,
		persona.EstadoCivil,
		persona.UpdatedAt,
		persona.ID)

	return err
}

func (r *crudRepository) Delete(id int) error {
	query := `DELETE FROM personas WHERE id=$1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *crudRepository) FindByID(id int) (*model.Persona, error) {
	persona := &model.Persona{}
	query := `SELECT id, nombre, estado_civil, user_id, created_at, updated_at 
	          FROM personas WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&persona.ID,
		&persona.Nombre,
		&persona.EstadoCivil,
		&persona.UserID,
		&persona.CreatedAt,
		&persona.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return persona, nil
}

func (r *crudRepository) FindAll(filter *model.PersonaFilter) (*model.PaginatedResponse, error) {
	// Configurar paginación
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 5
	}
	offset := (filter.Page - 1) * filter.PageSize

	// Construir query con filtros
	where := []string{}
	args := []interface{}{}
	argPos := 1

	if filter.Nombre != "" {
		where = append(where, fmt.Sprintf("nombre ILIKE $%d", argPos))
		args = append(args, "%"+filter.Nombre+"%")
		argPos++
	}

	if filter.EstadoCivil != "" {
		where = append(where, fmt.Sprintf("estado_civil = $%d", argPos))
		args = append(args, filter.EstadoCivil)
		argPos++
	}

	// Query para contar total
	countQuery := "SELECT COUNT(*) FROM personas"
	if len(where) > 0 {
		countQuery += " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Query para datos paginados
	dataQuery := "SELECT id, nombre, estado_civil, user_id, created_at, updated_at FROM personas"
	if len(where) > 0 {
		dataQuery += " WHERE " + strings.Join(where, " AND ")
	}
	dataQuery += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var personas []model.Persona
	for rows.Next() {
		var p model.Persona
		err := rows.Scan(&p.ID, &p.Nombre, &p.EstadoCivil, &p.UserID, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		personas = append(personas, p)
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize

	return &model.PaginatedResponse{
		Data:       personas,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
		HasNext:    filter.Page < totalPages,
		HasPrev:    filter.Page > 1,
	}, nil
}
