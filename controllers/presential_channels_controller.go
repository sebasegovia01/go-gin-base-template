package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/models"
)

type PresentialChannelRepositoryInterface interface {
	GetAllPresentialChannels() ([]models.PresentialChannel, error)
	GetPresentialChannelByID(channelIdentifier string) (models.PresentialChannel, error)
}

type PresentialChannelController struct {
	repository PresentialChannelRepositoryInterface
}

// NewPresentialChannelController creates a new controller for Presential Channels
func NewPresentialChannelController(repo PresentialChannelRepositoryInterface) *PresentialChannelController {
	return &PresentialChannelController{
		repository: repo,
	}
}

// GetPresentialChannels handles the request for fetching all Presential Channels or a single Presential Channel by ID
func (c *PresentialChannelController) GetPresentialChannels(ctx *gin.Context) {
	// Check if "id" is present in the path parameters
	channelIdentifier := ctx.Param("id")

	if channelIdentifier != "" {
		// If "id" is present, fetch the Presential Channel by its identifier
		log.Printf("Fetching Presential Channel with ID: %s", channelIdentifier)
		channel, err := c.repository.GetPresentialChannelByID(channelIdentifier)
		if err != nil {
			log.Printf("Error fetching Presential Channel by ID: %v", err)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Presential Channel not found"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"data": channel})
		return
	}

	// If "id" is not present, fetch all Presential Channels
	log.Println("Fetching all Presential Channels")
	channels, err := c.repository.GetAllPresentialChannels()
	if err != nil {
		log.Printf("Error fetching all Presential Channels: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching Presential Channels"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": channels})
}
