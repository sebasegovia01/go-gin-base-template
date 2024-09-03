package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/controllers"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	"github.com/sebasegovia01/base-template-go-gin/middleware"
	"github.com/sebasegovia01/base-template-go-gin/routes"
	"github.com/sebasegovia01/base-template-go-gin/services"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

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

	// Inicializar servicios
	storageService, err := services.NewStorageService(cfg)
	if err != nil {
		log.Fatalf("Error initializing storage service: %v", err)
	}
	defer storageService.Close()

	pubSubService, err := services.NewPubSubService(cfg)
	if err != nil {
		log.Fatalf("Error initializing PubSub service: %v", err)
	}

	pubSubPublishService, err := services.NewPubSubPublishService(cfg)
	if err != nil {
		log.Fatalf("Error initializing PubSub publish service: %v", err)
	}
	defer pubSubPublishService.Close()

	// Inicializar controladores
	dataCustomerController := controllers.NewDataCustomerController(pubSubService, storageService, pubSubPublishService)

	// Configurar rutas
	routes.SetupRoutes(r, dataCustomerController)

	// Routes log
	for _, route := range r.Routes() {
		log.Printf("Route: %s %s", route.Method, route.Path)
	}

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
