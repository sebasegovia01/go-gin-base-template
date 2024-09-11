package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/services"
	"github.com/sebasegovia01/base-template-go-gin/utils"
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

	// Obtener headers y almacenarlos en un mapa
	headers := make(map[string]string)
	for key, values := range ctx.Request.Header {
		headers[key] = values[0]
	}

	// Llamar al servicio HTTP con headers
	body, err := c.service.SendRequest("GET", url, nil, headers)
	if err != nil {
		log.Printf("Error fetching Automated Teller Machine: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ATM data"})
		return
	}

	// Transformar la respuesta para ajustar el formato
	transformedBody, err := utils.TransformResponse(body)
	if err != nil {
		log.Printf("Error transforming response: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transform ATM data"})
		return
	}

	// Enviar la respuesta transformada
	ctx.Data(http.StatusOK, "application/json", transformedBody)
}

// GetPresentialChannel llama al servicio HTTP para obtener datos de los canales presenciales
func (c *HTTPController) GetPresentialChannel(ctx *gin.Context) {
	id := ctx.Param("id")
	url := c.cfg.UrlMsPresentialChannels + "/" + id

	// Obtener headers y almacenarlos en un mapa
	headers := make(map[string]string)
	for key, values := range ctx.Request.Header {
		headers[key] = values[0]
	}

	// Llamar al servicio HTTP con headers
	body, err := c.service.SendRequest("GET", url, nil, headers)
	if err != nil {
		log.Printf("Error fetching Presential Channel: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Presential Channel data"})
		return
	}

	// Transformar la respuesta para ajustar el formato
	transformedBody, err := utils.TransformResponse(body)
	if err != nil {
		log.Printf("Error transforming response: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transform Presential Channel data"})
		return
	}

	// Enviar la respuesta transformada
	ctx.Data(http.StatusOK, "application/json", transformedBody)
}
