package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/controllers"
	"github.com/sebasegovia01/base-template-go-gin/middleware"
)

func SetupRoutes(r *gin.Engine, automatedTellerMachineController *controllers.AutomatedTellerMachineController,
	presentialChannelsController *controllers.PresentialChannelController) {

	api := r.Group("/service-channels/v1/api")

	// Health check route
	healthController := controllers.NewHealthController()
	api.GET("/health", healthController.HealthCheck)

	// automated teller machines routes
	{
		atms := api.Group("/automated-teller-machines")
		{
			atms.GET("/", WithTraceability(automatedTellerMachineController.GetATMs))

			atms.GET("/:id", WithTraceability(automatedTellerMachineController.GetATMs))
		}
	}

	// presential channels routes
	{
		presentialChannels := api.Group("/presentialchannels")
		{
			presentialChannels.GET("/", WithTraceability(presentialChannelsController.GetPresentialChannels))

			presentialChannels.GET("/:id", WithTraceability(presentialChannelsController.GetPresentialChannels))
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
