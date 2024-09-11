package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/services"
)

type HTTPController struct {
	service services.HTTPServiceInterface
	cfg     *config.Config
}

// NewHTTPController crea un nuevo controlador que usa un servicio HTTP y la configuración
func NewHTTPController(service services.HTTPServiceInterface, cfg *config.Config) *HTTPController {
	return &HTTPController{
		service: service,
		cfg:     cfg,
	}
}

// GetAutomatedTellerMachine llama al servicio HTTP para obtener datos de los cajeros automáticos
func (c *HTTPController) GetAutomatedTellerMachine(ctx *gin.Context) {
	id := ctx.Param("id")
	url := c.cfg.UrlMsAutomaticTellerMachines + "/" + id

	body, err := c.service.SendRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error fetching Automated Teller Machine: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ATM data"})
		return
	}

	// Retornar la respuesta del servicio HTTP
	ctx.Data(http.StatusOK, "application/json", body)
}

// GetPresentialChannel llama al servicio HTTP para obtener datos de los canales presenciales
func (c *HTTPController) GetPresentialChannel(ctx *gin.Context) {
	id := ctx.Param("id")
	url := c.cfg.UrlMsPresentialChannels + "/" + id

	body, err := c.service.SendRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error fetching Presential Channel: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Presential Channel data"})
		return
	}

	// Retornar la respuesta del servicio HTTP
	ctx.Data(http.StatusOK, "application/json", body)
}
