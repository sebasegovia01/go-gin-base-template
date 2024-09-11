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

// Función envoltorio para crear el servicio HTTP
func NewHTTPService(cfg *config.Config) services.HTTPServiceInterface {
	return services.NewHTTPService()
}

func setupServer(
	cfg *config.Config,
	newHTTPService func(cfg *config.Config) services.HTTPServiceInterface,
) (EngineRunner, error) {

	// Crear el servicio HTTP
	httpService := newHTTPService(cfg)

	// Inicializar el controlador HTTP
	httpController := controllers.NewHTTPController(httpService, cfg)

	// Configurar rutas
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.ResponseWrapperMiddleware())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
	})

	// Añadir rutas
	routes.SetupRoutes(r, httpController)

	// Adaptar *gin.Engine a EngineRunner
	return adaptGinToEngineRunner(r), nil
}

func run(
	loadConfigFunc func() (*config.Config, error),
	setupServerFunc func(
		cfg *config.Config,
		newHTTPService func(cfg *config.Config) services.HTTPServiceInterface,
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
	r, err := setupServerFunc(cfg, NewHTTPService)
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
