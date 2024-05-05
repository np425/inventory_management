package models

type Product struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	StockQuantity uint   `json:"stock_quantity"`
}
