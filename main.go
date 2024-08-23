package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	"github.com/sebasegovia01/base-template-go-gin/middleware"
	"github.com/sebasegovia01/base-template-go-gin/routes"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Inicializar base de datos
	db, err := config.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Init go gin
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery()) // recovery from panic, keep server running

	// middlewares
	r.Use(middleware.ResponseWrapperMiddleware())

	// Configurar manejador para rutas no encontradas
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
	})

	// Configurar rutas
	routes.SetupRoutes(r, db)

	// Iniciar servidor
	serverAddress := cfg.ServerAddress
	if serverAddress == "" {
		serverAddress = ":8080" // Puerto por defecto si no se especifica
	}

	environment := cfg.Environment
	if environment == "" {
		environment = enums.Dev // env por defecto
	}

	log.Printf("Server starting on port %s, environment is %s", serverAddress, environment)
	r.Run(serverAddress)
}
