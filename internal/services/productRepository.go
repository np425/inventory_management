package services

import "inventory/internal/models"

type ProductRepository interface {
	Create(product models.Product) (models.Product, error)
	Update(product models.Product) (models.Product, error)
	Delete(product models.Product) (models.Product, error)
	FindByID(id uint) (models.Product, error)
	List() ([]models.Product, error)
}
