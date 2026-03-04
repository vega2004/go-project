package service

import (
	"errors"
	"time"
	"tu-proyecto/model"
	"tu-proyecto/repository"
)

type ImagenService interface {
	Save(imagen *model.Imagen, userID int) error
	Delete(id int) error
	GetAll() ([]model.Imagen, error)
	GetForCarrusel() (*model.CarruselResponse, error)
	Reorder(ordenes []int) error
}

type imagenService struct {
	repo repository.ImagenRepository
}

func NewImagenService(repo repository.ImagenRepository) ImagenService {
	return &imagenService{repo: repo}
}

func (s *imagenService) Save(imagen *model.Imagen, userID int) error {
	imagen.UserID = userID
	imagen.Activo = true
	imagen.CreatedAt = time.Now()
	return s.repo.Create(imagen)
}

func (s *imagenService) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *imagenService) GetAll() ([]model.Imagen, error) {
	return s.repo.FindAll()
}

func (s *imagenService) GetForCarrusel() (*model.CarruselResponse, error) {
	return s.repo.GetForCarrusel()
}

func (s *imagenService) Reorder(ordenes []int) error {
	if len(ordenes) == 0 {
		return errors.New("no hay órdenes para actualizar")
	}
	// Implementar reordenamiento
	for i, id := range ordenes {
		if err := s.repo.UpdateOrder(id, i+1); err != nil {
			return err
		}
	}
	return nil
}
