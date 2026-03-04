package repository

import (
	"database/sql"
	"tu-proyecto/model"
)

type ImagenRepository interface {
	Create(imagen *model.Imagen) error
	Delete(id int) error
	FindAll() ([]model.Imagen, error)
	UpdateOrder(id, orden int) error
	GetForCarrusel() (*model.CarruselResponse, error)
}

type imagenRepository struct {
	db *sql.DB
}

func NewImagenRepository(db *sql.DB) ImagenRepository {
	return &imagenRepository{db: db}
}

func (r *imagenRepository) Create(imagen *model.Imagen) error {
	// Obtener el máximo orden actual
	var maxOrden int
	err := r.db.QueryRow("SELECT COALESCE(MAX(orden), 0) FROM imagenes").Scan(&maxOrden)
	if err != nil {
		return err
	}

	imagen.Orden = maxOrden + 1

	query := `INSERT INTO imagenes (nombre, ruta, orden, activo, user_id, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err = r.db.QueryRow(query,
		imagen.Nombre,
		imagen.Ruta,
		imagen.Orden,
		imagen.Activo,
		imagen.UserID,
		imagen.CreatedAt).Scan(&imagen.ID)

	return err
}

func (r *imagenRepository) Delete(id int) error {
	query := `DELETE FROM imagenes WHERE id=$1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *imagenRepository) FindAll() ([]model.Imagen, error) {
	query := `SELECT id, nombre, ruta, orden, activo, user_id, created_at 
	          FROM imagenes ORDER BY orden ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var imagenes []model.Imagen
	for rows.Next() {
		var i model.Imagen
		err := rows.Scan(&i.ID, &i.Nombre, &i.Ruta, &i.Orden, &i.Activo, &i.UserID, &i.CreatedAt)
		if err != nil {
			return nil, err
		}
		imagenes = append(imagenes, i)
	}

	return imagenes, nil
}

func (r *imagenRepository) UpdateOrder(id, orden int) error {
	query := `UPDATE imagenes SET orden=$1 WHERE id=$2`
	_, err := r.db.Exec(query, orden, id)
	return err
}

func (r *imagenRepository) GetForCarrusel() (*model.CarruselResponse, error) {
	query := `SELECT id, nombre, ruta, orden FROM imagenes WHERE activo = true ORDER BY orden ASC LIMIT 10`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var imagenes []model.Imagen
	for rows.Next() {
		var i model.Imagen
		err := rows.Scan(&i.ID, &i.Nombre, &i.Ruta, &i.Orden)
		if err != nil {
			return nil, err
		}
		imagenes = append(imagenes, i)
	}

	return &model.CarruselResponse{
		Imagenes: imagenes,
		Total:    len(imagenes),
	}, nil
}
