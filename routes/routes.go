package routes

import (
	"github.com/sebasegovia01/base-template-go-gin/controllers"
	"github.com/sebasegovia01/base-template-go-gin/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, dataCustomerController *controllers.DataCustomerController) {

	api := r.Group("/customer-data-retrieval/v1/api")

	// Health check route
	healthController := controllers.NewHealthController()
	api.GET("/health", healthController.HealthCheck)

	// Customer route
	{
		customers := api.Group("/customers")
		{
			customers.POST("/retrieve", dataCustomerController.HandlePushMessage)
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
