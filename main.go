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
	"github.com/sebasegovia01/base-template-go-gin/repositories"
	"github.com/sebasegovia01/base-template-go-gin/routes"
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

// Función envoltorio para NewAutomatedTellerMachineClient
func NewAutomatedTellerMachineInterface(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error) {
	return repositories.NewAutomatedTellerMachineClient(cfg, repositories.RealAutomatedTellerMachineCreator{})
}

// Función envoltorio para NewPresentialChannelClient
func NewPresentialChannelInterface(cfg *config.Config) (repositories.PresentialChannelInterface, error) {
	return repositories.NewPresentialChannelClient(cfg, repositories.RealPresentialChannelCreator{})
}

func setupServer(
	cfg *config.Config,
	newAutomatedTellerMachineClient func(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error),
	newPresentialChannelClient func(cfg *config.Config) (repositories.PresentialChannelInterface, error),
) (EngineRunner, error) {

	// Crear el cliente Datastore para ATM
	clientATM, err := newAutomatedTellerMachineClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Datastore client for ATM: %w", err)
	}

	// Crear el cliente Datastore para Presential Channel
	clientPresentialChannel, err := newPresentialChannelClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Datastore client for Presential Channel: %w", err)
	}

	// Crear el repositorio de ATM
	atmRepository := repositories.NewDatastoreATMRepository(clientATM, cfg.DataStoreDBName, cfg.DataStoreNamespace, cfg.DatastoreAutomaticTellerMachineKind)

	// Crear el repositorio de Presential Channel
	presentialChannelRepository := repositories.NewDatastorePresentialChannelRepository(clientPresentialChannel, cfg.DataStoreDBName, cfg.DataStoreNamespace, cfg.DatastorePresentialChannelKind)

	// Inicializar controladores
	atmController := controllers.NewAutomatedTellerMachineController(atmRepository)
	presentialChannelController := controllers.NewPresentialChannelController(presentialChannelRepository)

	// Configurar rutas
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.ResponseWrapperMiddleware())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
	})

	// Añadir rutas
	routes.SetupRoutes(r, atmController, presentialChannelController)

	// Adaptar *gin.Engine a EngineRunner
	return adaptGinToEngineRunner(r), nil
}

func run(
	loadConfigFunc func() (*config.Config, error),
	setupServerFunc func(
		cfg *config.Config,
		newAutomatedTellerMachineClient func(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error),
		newPresentialChannelClient func(cfg *config.Config) (repositories.PresentialChannelInterface, error),
	) (EngineRunner, error),
) error {

	// Cargar configuración
	cfg, err := loadConfigFunc()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	// Configurar dirección del servidor por defecto si no está definida
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = ":8080"
	}

	// Inicializar servidor usando setupServerFunc
	r, err := setupServerFunc(cfg, NewAutomatedTellerMachineInterface, NewPresentialChannelInterface)
	if err != nil {
		fatalfFunc("%v", err) // Usa fatalfFunc en lugar de log.Fatalf
		return err
	}

	// Configurar el entorno actual
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
