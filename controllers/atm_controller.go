package controllers

import (
	"net/http"
	"strconv"

	"github.com/sebasegovia01/base-template-go-gin/models"
	"github.com/sebasegovia01/base-template-go-gin/services"

	"github.com/gin-gonic/gin"
)

type ATMController struct {
	service *services.ATMService
}

func NewATMController(service *services.ATMService) *ATMController {
	return &ATMController{service: service}
}

func (c *ATMController) Create(ctx *gin.Context) {
	var atm models.ATM
	if err := ctx.ShouldBindJSON(&atm); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdATM, err := c.service.Create(atm)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdATM)
}

func (c *ATMController) GetAll(ctx *gin.Context) {
	atms, err := c.service.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, atms)
}

func (c *ATMController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	atm, err := c.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ATM not found"})
		return
	}

	ctx.JSON(http.StatusOK, atm)
}

func (c *ATMController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var atm models.ATM
	if err := ctx.ShouldBindJSON(&atm); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	atm.ID = id
	updatedATM, err := c.service.Update(atm)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedATM)
}

func (c *ATMController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	err := c.service.Delete(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "ATM deleted successfully"})
}
