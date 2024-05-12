package services

import (
	"database/sql"
	"fmt"
	"inventory/internal/models"
)

type OrderService struct {
	db             *sql.DB
	productService ProductServicable
}

func NewOrderService(db *sql.DB, productService ProductServicable) *OrderService {
	return &OrderService{db: db, productService: productService}
}

func (s *OrderService) Create(o *models.Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	p, err := s.productService.FindByID(o.ProductID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to fetch product: %w", err)
	}

	if o.Quantity > p.StockQuantity {
		tx.Rollback()
		return fmt.Errorf("order quantity exceeds product stock quantity")
	}

	// Insert the order
	query := `INSERT INTO orders (product_id, quantity, state) VALUES (?, ?, ?)`
	result, err := tx.Exec(query, o.ProductID, o.Quantity, o.State)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error creating order: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	o.ID = uint(id)

	// Enforce pending state for new order
	o.State = models.Pending

	return nil
}

func (s *OrderService) Update(o *models.Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	existingOrder, err := s.FindByID(o.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if existingOrder.ProductID != o.ProductID {
		tx.Rollback()
		return fmt.Errorf("cannot change product ID of an existing order")
	}

	_, err = tx.Exec("UPDATE orders SET quantity = ?, state = ? WHERE id = ?", o.Quantity, o.State, o.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating order: %w", err)
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	deltaStockQuantity := stockQuantityChange(existingOrder, *o)
	if deltaStockQuantity != 0 {
		product, err := s.productService.FindByID(o.ProductID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to fetch product: %w", err)
		}

		newStock := int(product.StockQuantity) + deltaStockQuantity
		if newStock < 0 {
			tx.Rollback()
			return fmt.Errorf("insufficient stock for product")
		}

		_, err = tx.Exec("UPDATE products SET stock_quantity = ? WHERE id = ?", newStock, product.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update product stock: %w", err)
		}

		product.StockQuantity = uint(newStock)
	}

	return nil
}

func (s *OrderService) Delete(id uint) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	query := `DELETE FROM orders WHERE id=?`
	result, err := tx.Exec(query, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error checking rows affected after deleting order: %w", err)
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("no order found with ID %d", id)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *OrderService) FindByID(id uint) (models.Order, error) {
	query := `SELECT id, product_id, quantity, state FROM orders WHERE id = ?`

	var order models.Order
	row := s.db.QueryRow(query, id)

	err := row.Scan(&order.ID, &order.ProductID, &order.Quantity, &order.State)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Order{}, fmt.Errorf("no order found with ID %d", id)
		}
		return models.Order{}, fmt.Errorf("error finding order: %w", err)
	}

	return order, nil
}

func (s *OrderService) FindByProductID(productID uint) ([]models.Order, error) {
	_, err := s.productService.FindByID(productID)
	if err != nil {
		return nil, err
	}

	orders := []models.Order{}
	rows, err := s.db.Query("SELECT id, product_id, quantity, state FROM orders WHERE product_id = ?", productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.ProductID, &o.Quantity, &o.State); err != nil {
			continue
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func stockQuantityChange(prevOrder, curOrder models.Order) int {
	return int(prevOrder.TakenStock()) - int(curOrder.TakenStock())
}
