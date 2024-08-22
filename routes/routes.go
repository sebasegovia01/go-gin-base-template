package routes

import (
	"database/sql"

	"github.com/sebasegovia01/base-template-go-gin/controllers"
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
			atms.POST("/", atmController.Create)
			atms.GET("/", atmController.GetAll)
			atms.GET("/:id", atmController.GetByID)
			atms.PUT("/:id", atmController.Update)
			atms.DELETE("/:id", atmController.Delete)
		}
	}
}
