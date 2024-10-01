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

// Variables globales para apuntar a las funciones reales, reemplazables en pruebas
var loadConfigFunc = config.Load
var setupServerFunc = setupServer
var fatalfFunc = log.Fatalf
var logPrintf = log.Printf

// Definir la interfaz EngineRunner para abstraer el servidor
type EngineRunner interface {
	Run(addr ...string) error
}

// Adaptar gin.Engine para que implemente la interfaz EngineRunner
func adaptGinToEngineRunner(engine *gin.Engine) EngineRunner {
	return engineRunnerAdapter{engine}
}

// engineRunnerAdapter es un adaptador para convertir *gin.Engine en EngineRunner
type engineRunnerAdapter struct {
	*gin.Engine
}

// Función envoltorio para NewStorageService
func NewStorageServiceInterface(cfg *config.Config) (services.StorageServiceInterface, error) {
	return services.NewStorageService(cfg)
}

// Función envoltorio para NewPubSubService
func NewPubSubServiceInterface(cfg *config.Config) (services.PubSubServiceInterface, error) {
	return services.NewPubSubService(cfg)
}

var newPubSubPublishServiceFunc = services.NewPubSubPublishService

func NewPubSubPublishServiceInterface(cfg *config.Config) (services.PubSubPublishServiceInterface, error) {
	return newPubSubPublishServiceFunc(cfg)
}

// Modificar setupServer para devolver EngineRunner
func setupServer(
	cfg *config.Config,
	newStorageService func(cfg *config.Config) (services.StorageServiceInterface, error),
	newPubSubService func(cfg *config.Config) (services.PubSubServiceInterface, error),
	newPubSubPublishService func(cfg *config.Config) (services.PubSubPublishServiceInterface, error),
) (EngineRunner, error) {
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
	phoneChannelController := controllers.NewDataPhoneChannelsController(pubSubService, storageService, pubSubPublishService)

	// Configurar rutas
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.ResponseWrapperMiddleware())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
	})

	routes.SetupRoutes(r, phoneChannelController)

	// Adaptar *gin.Engine a EngineRunner
	return adaptGinToEngineRunner(r), nil
}

// Modificar run para usar EngineRunner
func run(
	loadConfigFunc func() (*config.Config, error),
	setupServerFunc func(
		cfg *config.Config,
		newStorageService func(cfg *config.Config) (services.StorageServiceInterface, error),
		newPubSubService func(cfg *config.Config) (services.PubSubServiceInterface, error),
		newPubSubPublishService func(cfg *config.Config) (services.PubSubPublishServiceInterface, error),
	) (EngineRunner, error),
) error {

	// Load configs
	cfg, err := loadConfigFunc()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)

	}

	// Set default server address if not provided
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = ":8080"
	}

	// setup server usando las funciones envoltorio
	r, err := setupServerFunc(cfg, NewStorageServiceInterface, NewPubSubServiceInterface, NewPubSubPublishServiceInterface)
	if err != nil {
		fatalfFunc("%v", err) // Usa fatalfFunc en lugar de log.Fatalf
		return err
	}

	// set current environment
	// set current environment
	if cfg.Environment == "" {
		cfg.Environment = enums.Dev
	}

	logPrintf("Server starting on port %s, environment is %s", cfg.ServerAddress, cfg.Environment)
	return r.Run(cfg.ServerAddress)
}

func main() {
	if err := run(loadConfigFunc, setupServerFunc); err != nil {
		fatalfFunc("%v", err)
	}
}
