package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/controllers"
	"github.com/sebasegovia01/base-template-go-gin/middleware"
)

func SetupRoutes(r *gin.Engine, httpController *controllers.HTTPController) {

	api := r.Group("/service-channels/v1/api")

	// Health check route
	healthController := controllers.NewHealthController()
	api.GET("/health", healthController.HealthCheck)

	// Rutas para los cajeros automáticos (Automated Teller Machines)
	{
		atms := api.Group("/automated-teller-machines")
		{
			atms.GET("/:id", WithTraceability(httpController.GetAutomatedTellerMachine))
		}
	}

	// Rutas para los canales presenciales (Presential Channels)
	{
		channels := api.Group("/presentialchannels")
		{
			channels.GET("/:id", WithTraceability(httpController.GetPresentialChannel))
		}
	}
}

// use for headers traceability - add in paths for required.
func WithTraceability(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		middleware.TraceabilityMiddleware()(c)
		if !c.IsAborted() {
			handler(c)
		}
	}
}
