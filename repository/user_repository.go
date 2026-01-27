package repository

import (
	"database/sql"
	"log"
	"tu-proyecto/model"
)

type UserRepository interface {
	Create(user *model.User) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	query := `INSERT INTO users (name, email, phone, created_at) 
              VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.db.QueryRow(query,
		user.Name,
		user.Email,
		user.Phone,
		user.CreatedAt).Scan(&user.ID)

	if err != nil {
		log.Printf("Error al insertar usuario: %v", err)
		return err
	}

	log.Printf("Usuario insertado exitosamente. ID: %d", user.ID)
	return nil
}
