package controllers

import (
	"net/http"
	"strconv"

	"github.com/sebasegovia01/base-template-go-gin/models"
	"github.com/sebasegovia01/base-template-go-gin/services"

	"github.com/gin-gonic/gin"
)

type ATMController struct {
	service *services.AlloyDbService
}

func NewATMController(service *services.AlloyDbService) *ATMController {
	return &ATMController{service: service}
}

func (c *ATMController) Create(ctx *gin.Context) {
	var atm models.ATM
	if err := ctx.ShouldBindJSON(&atm); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `INSERT INTO presential_service_channels.automated_teller_machines (
		atmidentifier, atmaddress_streetname, atmaddress_buildingnumber, 
		atmtownname, atmdistrictname, atmcountrysubdivisionmajorname, 
		atmfromdatetime, atmtodatetime, atmtimetype, atmattentionhour, 
		atmservicetype, atmaccesstype
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) 
	RETURNING id, atmidentifier, atmaddress_streetname, atmaddress_buildingnumber, 
	atmtownname, atmdistrictname, atmcountrysubdivisionmajorname, 
	atmfromdatetime, atmtodatetime, atmtimetype, atmattentionhour, 
	atmservicetype, atmaccesstype`

	params := []interface{}{
		atm.ATMIdentifier, atm.ATMAddressStreetName, atm.ATMAddressBuildingNumber,
		atm.ATMTownName, atm.ATMDistrictName, atm.ATMCountrySubdivisionMajorName,
		atm.ATMFromDateTime, atm.ATMToDateTime, atm.ATMTimeType, atm.ATMAttentionHour,
		atm.ATMServiceType, atm.ATMAccessType,
	}

	results, err := c.service.ExecuteQuery(query, params, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(results) == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ATM"})
		return
	}

	ctx.JSON(http.StatusCreated, results[0])
}

func (c *ATMController) GetAll(ctx *gin.Context) {
	query := `SELECT id, atmidentifier, atmaddress_streetname, atmaddress_buildingnumber, 
	atmtownname, atmdistrictname, atmcountrysubdivisionmajorname, 
	atmfromdatetime, atmtodatetime, atmtimetype, atmattentionhour, 
	atmservicetype, atmaccesstype FROM presential_service_channels.automated_teller_machines`

	results, err := c.service.ExecuteQuery(query, nil, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (c *ATMController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	query := `SELECT id, atmidentifier, atmaddress_streetname, atmaddress_buildingnumber, 
	atmtownname, atmdistrictname, atmcountrysubdivisionmajorname, 
	atmfromdatetime, atmtodatetime, atmtimetype, atmattentionhour, 
	atmservicetype, atmaccesstype FROM presential_service_channels.automated_teller_machines WHERE id = $1`
	params := []interface{}{id}

	results, err := c.service.ExecuteQuery(query, params, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(results) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ATM not found"})
		return
	}

	ctx.JSON(http.StatusOK, results[0])
}

func (c *ATMController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var atm models.ATM
	if err := ctx.ShouldBindJSON(&atm); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE presential_service_channels.automated_teller_machines SET 
		atmidentifier = $1, atmaddress_streetname = $2, atmaddress_buildingnumber = $3,
		atmtownname = $4, atmdistrictname = $5, atmcountrysubdivisionmajorname = $6,
		atmfromdatetime = $7, atmtodatetime = $8, atmtimetype = $9, atmattentionhour = $10,
		atmservicetype = $11, atmaccesstype = $12
		WHERE id = $13
		RETURNING id, atmidentifier, atmaddress_streetname, atmaddress_buildingnumber, 
		atmtownname, atmdistrictname, atmcountrysubdivisionmajorname, 
		atmfromdatetime, atmtodatetime, atmtimetype, atmattentionhour, 
		atmservicetype, atmaccesstype`

	params := []interface{}{
		atm.ATMIdentifier, atm.ATMAddressStreetName, atm.ATMAddressBuildingNumber,
		atm.ATMTownName, atm.ATMDistrictName, atm.ATMCountrySubdivisionMajorName,
		atm.ATMFromDateTime, atm.ATMToDateTime, atm.ATMTimeType, atm.ATMAttentionHour,
		atm.ATMServiceType, atm.ATMAccessType, id,
	}

	results, err := c.service.ExecuteQuery(query, params, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(results) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ATM not found"})
		return
	}

	ctx.JSON(http.StatusOK, results[0])
}

func (c *ATMController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	query := "DELETE FROM presential_service_channels.automated_teller_machines WHERE id = $1"
	params := []interface{}{id}

	_, err := c.service.ExecuteQuery(query, params, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "ATM deleted successfully"})
}
