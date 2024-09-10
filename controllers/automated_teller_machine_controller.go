package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/models"
)

type AutomatedTellerMachineRepositoryInterface interface {
	GetAllATMs() ([]models.AutomatedTellerMachine, error)
	GetATMByID(atmIdentifier string) (models.AutomatedTellerMachine, error)
}

type AutomatedTellerMachineController struct {
	repository AutomatedTellerMachineRepositoryInterface
}

// NewAutomatedTellerMachineController creates a new controller for Automated Teller Machines
func NewAutomatedTellerMachineController(repo AutomatedTellerMachineRepositoryInterface) *AutomatedTellerMachineController {
	return &AutomatedTellerMachineController{
		repository: repo,
	}
}

// GetATMs handles the request for fetching all ATMs or a single ATM by ID
func (c *AutomatedTellerMachineController) GetATMs(ctx *gin.Context) {
	// Check if "id" is present in the path parameters
	atmIdentifier := ctx.Param("id")

	if atmIdentifier != "" {
		// If "id" is present, fetch the ATM by its identifier
		log.Printf("Fetching ATM with ID: %s", atmIdentifier)
		atm, err := c.repository.GetATMByID(atmIdentifier)
		if err != nil {
			log.Printf("Error fetching ATM by ID: %v", err)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "ATM not found"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"data": atm})
		return
	}

	// If "id" is not present, fetch all ATMs
	log.Println("Fetching all ATMs")
	atms, err := c.repository.GetAllATMs()
	if err != nil {
		log.Printf("Error fetching all ATMs: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching ATMs"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": atms})
}
