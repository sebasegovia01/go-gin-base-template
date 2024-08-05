package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	"github.com/sebasegovia01/base-template-go-gin/repositories"
	"github.com/sebasegovia01/base-template-go-gin/routes"
	"github.com/sebasegovia01/base-template-go-gin/services"
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

	// Inicializar repositorio ATM
	atmRepo := repositories.NewATMRepository(db)

	// Inicializar servicio ATM
	atmService := services.NewATMService(atmRepo)

	// Inicializar servicio Storage
	storageService, err := services.NewStorageService(cfg, atmService)
	if err != nil {
		log.Fatalf("Error initializing Storage service: %v", err)
	}
	defer storageService.Close()

	// Inicializar servicio PubSub
	pubsubService, err := services.NewPubSubService(cfg, storageService)
	if err != nil {
		log.Printf("Error initializing PubSub service: %v", err)
		// Cerrar la conexión de la base de datos antes de salir
		db.Close()
		os.Exit(1)
	}
	defer pubsubService.Close()

	// Iniciar la recepción de mensajes en una goroutine
	go func() {
		if err := pubsubService.ReceiveMessages(); err != nil {
			log.Printf("Error in PubSub message receiving: %v", err)
			// Aquí podrías implementar una lógica para reintentar o notificar
		}
	}()
	// Inicializar router
	r := gin.Default()

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
