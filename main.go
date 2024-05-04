package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"inventory/internal/services"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := "./db/app.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	err = initializeDB(db)
	if err != nil {
		log.Fatal(err)
	}

	productService := services.NewProductService(db)
	orderService := services.NewOrderService(db, productService)

	router := gin.Default()

	router.SetHTMLTemplate(template.Must(template.ParseGlob("./public/templates/*.html")))
	router.Static("/static", "./public/static")

	setupRoutes(router, productService, orderService)

	err = router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server %v", err)
	}
}

func initializeDB(db *sql.DB) error {
	// Read SQL file using os.ReadFile
	sqlFile, err := os.ReadFile("./db/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}

	// Execute SQL script
	if _, err := db.Exec(string(sqlFile)); err != nil {
		return fmt.Errorf("error executing SQL script: %v", err)
	}

	return nil
}

func setupRoutes(router *gin.Engine, p services.ProductRepository, o services.OrderRepository) {
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", nil)
	})

	router.GET("/about", func(c *gin.Context) {
		log.Println("Rendering the about page")
		c.HTML(http.StatusOK, "about.html", nil)
	})

}
