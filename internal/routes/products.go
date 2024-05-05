package routes

import (
	"inventory/internal/models"
	"inventory/internal/services"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetProducts(c *gin.Context, p services.ProductRepository) {
	products, err := p.List()
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.HTML(http.StatusOK, "products.html", gin.H{
		"Products": products,
	})
}

func PostProducts(c *gin.Context, p services.ProductRepository) {
	defer c.Redirect(http.StatusFound, "/products")

	name := c.PostForm("name")
	stockQuantity, err := strconv.ParseUint(c.PostForm("stock_quantity"), 10, 32)
	if err != nil {
		c.Error(err)
		return
	}

	id, err := strconv.Atoi(c.PostForm("id"))
	if err != nil {
		id = -1
	}

	if id <= 0 {
		product := models.Product{Name: name, StockQuantity: uint(stockQuantity)}
		_, err = p.Create(product)
		if err != nil {
			c.Error(err)
		}
	} else {
		product := models.Product{ID: uint(id), Name: name, StockQuantity: uint(stockQuantity)}
		_, err = p.Update(product)
		if err != nil {
			c.Error(err)
		}
	}
}

func DeleteProducts(c *gin.Context, p services.ProductRepository) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Error(err)
		return
	}

	id, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		c.Error(err)
		return
	}

	product := models.Product{ID: uint(id)}
	_, err = p.Delete(product)
	if err != nil {
		c.Error(err)
		return
	}
}
