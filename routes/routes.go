package routes

import (
	"database/sql"

	"github.com/sebasegovia01/base-template-go-gin/controllers"
	"github.com/sebasegovia01/base-template-go-gin/middleware"
	"github.com/sebasegovia01/base-template-go-gin/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, db *sql.DB) {

	api := r.Group("/service-channels/v1/api")

	// Health check route
	healthController := controllers.NewHealthController()
	api.GET("/health", healthController.HealthCheck)

	alloyDbService := services.NewAlloyDB(db)
	atmController := controllers.NewATMController(alloyDbService)
	{
		atms := api.Group("/atms")
		{
			atms.POST("/", WithTraceability(atmController.Create))
			atms.GET("/", WithTraceability(atmController.GetAll))
			atms.GET("/:id", WithTraceability(atmController.GetByID))
			atms.PUT("/:id", WithTraceability(atmController.Update))
			atms.DELETE("/:id", WithTraceability(atmController.Delete))
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
