package services

import (
	"database/sql"
	"fmt"
	"inventory/internal/models"
	"log"
)

type OrderService struct {
	db          *sql.DB
	productRepo ProductRepository
}

func NewOrderService(db *sql.DB, productRepo ProductRepository) *OrderService {
	return &OrderService{db: db, productRepo: productRepo}
}

func (s *OrderService) Create(o models.Order) (models.Order, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to start transaction: %w", err)
	}

	o.State = models.Pending

	// Check product availability and update stock
	p, err := s.productRepo.FindByID(o.ProductID)
	if err != nil {
		tx.Rollback()
		return models.Order{}, err
	}

	if o.Quantity > p.StockQuantity {
		tx.Rollback()
		return models.Order{}, fmt.Errorf("order quantity exceeds product stock quantity")
	}

	// Insert the order
	query := `INSERT INTO orders (product_id, quantity, state) VALUES (?, ?, ?)`
	result, err := tx.Exec(query, o.ProductID, o.Quantity, o.State)
	if err != nil {
		tx.Rollback()
		log.Printf("Error creating order: %v", err)
		return models.Order{}, fmt.Errorf("error creating order: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		log.Printf("Error getting last insert ID: %v", err)
		return models.Order{}, fmt.Errorf("error getting last insert ID: %w", err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return models.Order{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	o.ID = uint(id)
	return o, nil
}

func (s *OrderService) Update(o models.Order) (models.Order, error) {
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return models.Order{}, fmt.Errorf("failed to start transaction: %w", err)
	}

	existingOrder, err := s.FindByID(o.ID)
	if err != nil {
		tx.Rollback()
		return models.Order{}, err
	}

	if existingOrder.ProductID != o.ProductID {
		tx.Rollback()
		return models.Order{}, fmt.Errorf("cannot change product ID of an existing order")
	}

	deltaStockQuantity := stockQuantityChange(existingOrder, o)

	if deltaStockQuantity != 0 {
		product, err := s.productRepo.FindByID(o.ProductID)
		if err != nil {
			tx.Rollback()
			return models.Order{}, fmt.Errorf("failed to fetch product: %w", err)
		}

		newStock := int(product.StockQuantity) + deltaStockQuantity
		if newStock < 0 {
			tx.Rollback()
			return models.Order{}, fmt.Errorf("insufficient stock for product")
		}

		product.StockQuantity = uint(newStock)
		_, err = tx.Exec("UPDATE products SET stock_quantity = ? WHERE id = ?", product.StockQuantity, product.ID)
		if err != nil {
			tx.Rollback()
			return models.Order{}, fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	_, err = tx.Exec("UPDATE orders SET quantity = ?, state = ? WHERE id = ?", o.Quantity, o.State, o.ID)
	if err != nil {
		tx.Rollback()
		return models.Order{}, fmt.Errorf("error updating order: %w", err)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return models.Order{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return o, nil
}

func (s *OrderService) Delete(o models.Order) (models.Order, error) {
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return models.Order{}, fmt.Errorf("failed to start transaction: %w", err)
	}

	query := `DELETE FROM orders WHERE id=?`
	result, err := tx.Exec(query, o.ID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error deleting order: %v", err)
		return models.Order{}, fmt.Errorf("error deleting order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		log.Printf("Error checking rows affected after deleting order: %v", err)
		return models.Order{}, fmt.Errorf("error checking rows affected after deleting order: %w", err)
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return models.Order{}, fmt.Errorf("no order found with ID %d", o.ID)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return models.Order{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return o, nil
}

func (s *OrderService) FindByID(id uint) (models.Order, error) {
	query := `SELECT id, product_id, quantity, state FROM orders WHERE id = ?`
	var order models.Order
	row := s.db.QueryRow(query, id)
	err := row.Scan(&order.ID, &order.ProductID, &order.Quantity, &order.State)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No order found with ID: %v", id)
			return models.Order{}, fmt.Errorf("no order found with ID %d", id)
		}
		log.Printf("Error finding order: %v", err)
		return models.Order{}, fmt.Errorf("error finding order: %w", err)
	}

	return order, nil
}

func (s *OrderService) List() ([]models.Order, error) {
	orders := []models.Order{}
	rows, err := s.db.Query("SELECT id, product_id, quantity, state FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.ProductID, &o.Quantity, &o.State); err != nil {
			log.Println("Failed to scan row:", err)
			continue
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func stockQuantityChange(prevOrder, curOrder models.Order) int {
	return int(prevOrder.TakenStock()) - int(curOrder.TakenStock())
}
