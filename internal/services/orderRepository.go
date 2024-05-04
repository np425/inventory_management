package services

import "inventory/internal/models"

type OrderRepository interface {
	Create(order models.Order) (models.Order, error)
	Update(order models.Order) (models.Order, error)
	Delete(order models.Order) (models.Order, error)
	FindByID(id uint) (models.Order, error)
	List() ([]models.Order, error)
}
