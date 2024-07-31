package routes

import (
	"database/sql"

	"github.com/sebasegovia01/base-template-go-gin/controllers"
	"github.com/sebasegovia01/base-template-go-gin/repositories"
	"github.com/sebasegovia01/base-template-go-gin/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, db *sql.DB) {
	atmRepo := repositories.NewATMRepository(db)
	atmService := services.NewATMService(atmRepo)
	atmController := controllers.NewATMController(atmService)

	api := r.Group("/api")
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
