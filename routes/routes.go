package routes

import (
	"github.com/sebasegovia01/base-template-go-gin/controllers"
	"github.com/sebasegovia01/base-template-go-gin/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, phoneChannelController *controllers.DataPhoneChannelsController) {

	api := r.Group("/service-channels/v1/api")

	// Health check route
	healthController := controllers.NewHealthController()
	api.GET("/health", healthController.HealthCheck)

	// Customer route
	{
		customers := api.Group("/phone/channels")
		{
			customers.POST("/retrieve", phoneChannelController.HandlePushMessage)
		}
	}
}

// use for headers traceability - add in paths.
func WithTraceability(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		middleware.TraceabilityMiddleware()(c)
		if !c.IsAborted() {
			handler(c)
		}
	}
}
