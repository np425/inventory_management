package services

import "inventory/internal/models"

type OrderServicable interface {
	Create(order *models.Order) error
	Update(order *models.Order) error
	Delete(id uint) error
	FindByID(id uint) (models.Order, error)
	FindByProductID(productID uint) ([]models.Order, error)
}
