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

type DataElectronicChannelsController struct {
	pubSubService        services.PubSubServiceInterface
	storageService       services.StorageServiceInterface
	pubSubPublishService services.PubSubPublishServiceInterface
}

func NewDataElectronicChannelsController(
	pubSubService services.PubSubServiceInterface,
	storageService services.StorageServiceInterface,
	pubSubPublishService services.PubSubPublishServiceInterface,
) *DataElectronicChannelsController {
	return &DataElectronicChannelsController{
		pubSubService:        pubSubService,
		storageService:       storageService,
		pubSubPublishService: pubSubPublishService,
	}
}

var transformElectronicChannelDataFunc = utils.TransformElectronicChannelData
var customMarshalJSONFunc = utils.CustomMarshalJSON

func (c *DataElectronicChannelsController) HandlePushMessage(ctx *gin.Context) {
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
	channelDataList, err := c.storageService.ProcessFile(storageEvent.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing file: " + err.Error()})
		return
	}

	log.Printf("channelDataList: %+v", channelDataList)

	var transformedChannels []json.RawMessage
	for _, channelData := range channelDataList {
		log.Printf("channelData before transformation: %+v", channelData)

		transformedChannel, err := transformElectronicChannelDataFunc(channelData)
		if err != nil {
			log.Printf("Error transforming channel data: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error transforming channel data: " + err.Error()})
			return
		}

		log.Printf("Transformed channel data: %+v", transformedChannel)

		// Test with regular json.Marshal instead of CustomMarshalJSON
		channelJSON, err := json.Marshal(transformedChannel)
		if err != nil {
			log.Printf("Error marshaling channel data: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error marshaling channel data: " + err.Error()})
			return
		}

		log.Printf("Marshalled channel data: %s", string(channelJSON))

		transformedChannels = append(transformedChannels, channelJSON)
	}

	// Iteramos y publicamos la data 1 x 1
	for _, channel := range transformedChannels {
		err = c.pubSubPublishService.PublishMessage(channel)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error publishing message: " + err.Error()})
			return
		}
	}

	// respuesta final del servicio
	ctx.JSON(http.StatusOK, gin.H{
		"status":     "Electronic channel data processed and published successfully",
		"data_count": len(channelDataList),
	})
}
