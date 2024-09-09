package main

import (
	"fmt"
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

func setupServer(
	cfg *config.Config,
	newStorageService func(cfg *config.Config) (*services.StorageService, error),
	newPubSubService func(cfg *config.Config) (*services.PubSubService, error),
	newPubSubPublishService func(cfg *config.Config) (*services.PubSubPublishService, error),
) (*gin.Engine, error) {
	// Inicializar servicios
	storageService, err := newStorageService(cfg)
	if err != nil {
		if storageService != nil {
			storageService.Close()
		}
		return nil, fmt.Errorf("error initializing storage service: %w", err)
	}

	pubSubService, err := newPubSubService(cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing PubSub service: %w", err)
	}

	pubSubPublishService, err := newPubSubPublishService(cfg)
	if err != nil {
		if pubSubPublishService != nil {
			pubSubPublishService.Close()
		}
		return nil, fmt.Errorf("error initializing PubSub publish service: %w", err)
	}

	// Inicializar controladores
	dataCustomerController := controllers.NewDataCustomerController(pubSubService, storageService, pubSubPublishService)

	// Configurar rutas
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.ResponseWrapperMiddleware())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
	})

	routes.SetupRoutes(r, dataCustomerController)

	return r, nil
}

var loadConfig = config.Load

func main() {

	// Load configs
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// setup server
	r, err := setupServer(cfg, services.NewStorageService, services.NewPubSubService, services.NewPubSubPublishService)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// init server
	serverAddress := cfg.ServerAddress
	if serverAddress == "" {
		serverAddress = ":8080"
	}

	// set current environment
	environment := cfg.Environment
	if environment == "" {
		environment = enums.Dev
	}

	log.Printf("Server starting on port %s, environment is %s", serverAddress, environment)
	r.Run(serverAddress)
}
