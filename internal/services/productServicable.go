package services

import "inventory/internal/models"

type ProductServicable interface {
	Create(product *models.Product) error
	Update(product *models.Product) error
	Delete(id uint) error
	FindByID(id uint) (models.Product, error)
	List() ([]models.Product, error)
}
