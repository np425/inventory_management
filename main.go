package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"inventory/internal/routes"
	"inventory/internal/services"
	"log"
	"net/http"
	"os"
	"path/filepath"

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

	err = setupTemplates(router)
	if err != nil {
		log.Fatal(err)
	}

	router.Static("/static", "./public/static")

	setupRoutes(router, productService, orderService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Println("INFO: No PORT environment variable detected, defaulting to 8080")
	}

	err = router.Run(":" + port)
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

func setupTemplates(router *gin.Engine) error {
	// Set up json conversion for templates
	funcMap := template.FuncMap{
		"json": toJSON,
	}

	tmpl := template.New("").Funcs(funcMap)

	templateFiles, err := filepath.Glob("./public/templates/*.html")
	if err != nil {
		return err
	}

	for _, templateFile := range templateFiles {
		_, err = tmpl.ParseFiles(templateFile)
		if err != nil {
			return err
		}
	}

	router.SetHTMLTemplate(tmpl) // Set the parsed templates with functions to the router
	return nil
}

func setupRoutes(router *gin.Engine, p services.ProductServicable, o services.OrderServicable) {
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/products")
	})

	router.GET("/products", func(c *gin.Context) {
		routes.GetProducts(c, p)
	})

	router.POST("/products", func(c *gin.Context) {
		routes.PostProducts(c, p)
	})

	router.DELETE("/products", func(c *gin.Context) {
		routes.DeleteProducts(c, p)
	})

	router.GET("/orders/:id", func(c *gin.Context) {
		routes.GetOrders(c, o)
	})

	router.POST("/orders/:id", func(c *gin.Context) {
		routes.PostOrders(c, o)
	})

	router.DELETE("/orders/:id", func(c *gin.Context) {
		routes.DeleteOrders(c, o)
	})
}

func toJSON(v interface{}) template.JS {
	a, _ := json.Marshal(v)
	return template.JS(a)
}
