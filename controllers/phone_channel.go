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

type DataPhoneChannelsController struct {
	pubSubService        services.PubSubServiceInterface
	storageService       services.StorageServiceInterface
	pubSubPublishService services.PubSubPublishServiceInterface
}

func NewDataPhoneChannelsController(
	pubSubService services.PubSubServiceInterface,
	storageService services.StorageServiceInterface,
	pubSubPublishService services.PubSubPublishServiceInterface,
) *DataPhoneChannelsController {
	return &DataPhoneChannelsController{
		pubSubService:        pubSubService,
		storageService:       storageService,
		pubSubPublishService: pubSubPublishService,
	}
}

var transformPhonechannelDataFunc = utils.TransformPhoneChannelData
var customMarshalJSONFunc = utils.CustomMarshalJSON

func (c *DataPhoneChannelsController) HandlePushMessage(ctx *gin.Context) {
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

	// Si el storageEvent es nil, significa que se ha ignorado el mensaje (eventType no soportado)
	if storageEvent == nil {
		log.Printf("Storage event ignored due to unsupported eventType")
		ctx.JSON(http.StatusOK, gin.H{
			"status":      "Event ignored",
			"description": "Event type is not supported, no action taken",
		})
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

		transformedChannel, err := transformPhonechannelDataFunc(channelData)
		if err != nil {
			log.Printf("Error transforming channel data: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error transforming channel data: " + err.Error()})
			return
		}

		log.Printf("Transformed channel data: %+v", transformedChannel)

		channelJSON, err := customMarshalJSONFunc(transformedChannel)
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
		"status":     "Phone channel data processed and published successfully",
		"data_count": len(channelDataList),
	})
}
