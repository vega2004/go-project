package repository

import (
	"database/sql"
	"fmt"
	"time"
	"tu-proyecto/model"
)

type PerfilRepository interface {
	GetPerfil(userID int) (*model.Perfil, error)
	UpdatePerfil(perfil *model.Perfil) error
	UpdateFoto(userID int, rutaFoto string) error
}

type perfilRepository struct {
	db *sql.DB
}

func NewPerfilRepository(db *sql.DB) PerfilRepository {
	return &perfilRepository{db: db}
}

func (r *perfilRepository) GetPerfil(userID int) (*model.Perfil, error) {
	fmt.Printf("🔍 Buscando perfil para userID: %d\n", userID)

	perfil := &model.Perfil{}
	query := `SELECT user_id, foto, bio, ubicacion, sitio_web, updated_at 
              FROM perfiles WHERE user_id = $1`

	fmt.Printf("📝 Query: %s\n", query)

	err := r.db.QueryRow(query, userID).Scan(
		&perfil.UserID,
		&perfil.Foto,
		&perfil.Bio,
		&perfil.Ubicacion,
		&perfil.SitioWeb,
		&perfil.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		fmt.Println("⚠️ No se encontró perfil, creando uno por defecto")
		return r.createDefaultPerfil(userID)
	}

	if err != nil {
		fmt.Printf("❌ Error en QueryRow: %v\n", err)
		return nil, err
	}

	fmt.Printf("✅ Perfil encontrado: %+v\n", perfil)
	return perfil, nil
}
func (r *perfilRepository) createDefaultPerfil(userID int) (*model.Perfil, error) {
	query := `INSERT INTO perfiles (user_id, foto, updated_at) 
	          VALUES ($1, $2, $3) RETURNING user_id, foto, updated_at`

	perfil := &model.Perfil{
		UserID:    userID,
		Foto:      "/static/uploads/perfil/default-avatar.png",
		UpdatedAt: time.Now(),
	}

	err := r.db.QueryRow(query,
		perfil.UserID,
		perfil.Foto,
		perfil.UpdatedAt,
	).Scan(&perfil.UserID, &perfil.Foto, &perfil.UpdatedAt)

	return perfil, err
}

func (r *perfilRepository) UpdatePerfil(perfil *model.Perfil) error {
	query := `UPDATE perfiles SET bio=$1, ubicacion=$2, sitio_web=$3, updated_at=$4 
	          WHERE user_id=$5`

	_, err := r.db.Exec(query,
		perfil.Bio,
		perfil.Ubicacion,
		perfil.SitioWeb,
		time.Now(),
		perfil.UserID,
	)
	return err
}

func (r *perfilRepository) UpdateFoto(userID int, rutaFoto string) error {
	query := `UPDATE perfiles SET foto=$1, updated_at=$2 WHERE user_id=$3`
	_, err := r.db.Exec(query, rutaFoto, time.Now(), userID)
	return err
}
