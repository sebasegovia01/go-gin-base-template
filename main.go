package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/config"
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

	// Inicializar router
	r := gin.Default()

	// Configurar rutas
	routes.SetupRoutes(r, db)

	// Iniciar servidor
	serverAddress := cfg.ServerAddress
	if serverAddress == "" {
		serverAddress = ":8080" // Puerto por defecto si no se especifica
	}
	log.Printf("Server starting on %s", serverAddress)
	r.Run(serverAddress)
}
