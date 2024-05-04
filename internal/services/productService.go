package services

import (
	"database/sql"
	"fmt"
	"inventory/internal/models"
	"log"
)

type ProductService struct {
	db *sql.DB
}

func NewProductService(db *sql.DB) *ProductService {
	return &ProductService{db: db}
}

func (s *ProductService) Create(p models.Product) (models.Product, error) {
	query := `INSERT INTO products (name, stock_quantity) VALUES (?, ?)`

	result, err := s.db.Exec(query, p.Name, p.StockQuantity)
	if err != nil {
		log.Printf("Error creating product: %v", err)
		return models.Product{}, fmt.Errorf("error creating product: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		return models.Product{}, fmt.Errorf("error getting last insert ID: %w", err)
	}

	p.ID = uint(id)
	return p, nil
}

func (s *ProductService) Update(p models.Product) (models.Product, error) {
	query := `UPDATE products SET name=?, stock_quantity=? WHERE id=?`

	_, err := s.db.Exec(query, p.Name, p.StockQuantity, p.ID)
	if err != nil {
		log.Printf("Error updating product: %v", err)
		return models.Product{}, fmt.Errorf("error updating product: %w", err)
	}

	return p, nil
}

func (s *ProductService) Delete(p models.Product) (models.Product, error) {
	query := `DELETE FROM products WHERE id=?`

	result, err := s.db.Exec(query, p.ID)
	if err != nil {
		log.Printf("Error deleting product: %v", err)
		return models.Product{}, fmt.Errorf("error deleting product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error checking rows affected after deleting product: %v", err)
		return models.Product{}, fmt.Errorf("error checking rows affected after deleting product: %w", err)
	}

	if rowsAffected == 0 {
		return models.Product{}, fmt.Errorf("no product found with ID %d", p.ID)
	}

	return p, nil
}

func (s *ProductService) FindByID(id uint) (models.Product, error) {
	query := `SELECT id, name, stock_quantity FROM products WHERE id = ?`

	var product models.Product

	row := s.db.QueryRow(query, id)
	err := row.Scan(&product.ID, &product.Name, &product.StockQuantity)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No product found with ID: %v", id)
			return models.Product{}, fmt.Errorf("no product found with ID %d", id)
		}
		log.Printf("Error finding product: %v", err)
		return models.Product{}, fmt.Errorf("error finding product: %w", err)
	}

	return product, nil
}

func (s *ProductService) List() ([]models.Product, error) {
	products := []models.Product{}
	rows, err := s.db.Query("SELECT id, name, stock_quantity FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.StockQuantity); err != nil {
			log.Println("Failed to scan row:", err)
			continue
		}
		products = append(products, p)
	}

	return products, nil
}
