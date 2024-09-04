package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/services"
	"github.com/sebasegovia01/base-template-go-gin/utils"
)

type DataCustomerController struct {
	pubSubService        *services.PubSubService
	storageService       *services.StorageService
	pubSubPublishService *services.PubSubPublishService
}

func NewDataCustomerController(pubSubService *services.PubSubService, storageService *services.StorageService,
	pubSubPublishService *services.PubSubPublishService) *DataCustomerController {
	return &DataCustomerController{
		pubSubService:        pubSubService,
		storageService:       storageService,
		pubSubPublishService: pubSubPublishService,
	}
}

func (c *DataCustomerController) HandlePushMessage(ctx *gin.Context) {

	// Loguear el cuerpo de la solicitud
	body, _ := io.ReadAll(ctx.Request.Body)
	log.Printf("Received request body: %s", string(body))
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// obtenemos evento de storage
	storageEvent, err := c.pubSubService.ExtractStorageEvent(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Storage event data: %s", storageEvent)

	// obtenemos la data
	customerDataList, err := c.storageService.ProcessFile(storageEvent.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing file: " + err.Error()})
		return
	}

	log.Printf("customerDataList: %s", customerDataList)

	var transformedCustomers []json.RawMessage
	for _, customerData := range customerDataList {
		transformedCustomer, err := utils.TransformCustomerData(customerData)
		if err != nil {
			log.Printf("Error transforming customer data: %v", err)
			continue
		}

		customerJSON, err := utils.CustomMarshalJSON(transformedCustomer)
		if err != nil {
			log.Printf("Error marshaling customer data: %v", err)
			continue
		}

		transformedCustomers = append(transformedCustomers, customerJSON)
	}

	// Iteramos y publicamos la data 1 x 1
	for _, customer := range transformedCustomers {
		err = c.pubSubPublishService.PublishMessage(customer)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error publishing message: " + err.Error()})
			return
		}
	}

	// respuesta final del servicio
	ctx.JSON(http.StatusOK, gin.H{
		"status":     "Customer data processed and published successfully",
		"data_count": len(customerDataList),
	})
}
