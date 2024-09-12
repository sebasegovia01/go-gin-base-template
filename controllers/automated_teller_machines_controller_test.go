package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para AutomatedTellerMachineRepositoryInterface
type MockAutomatedTellerMachineRepository struct {
	mock.Mock
}

func (m *MockAutomatedTellerMachineRepository) GetAllATMs() ([]models.AutomatedTellerMachine, error) {
	args := m.Called()
	return args.Get(0).([]models.AutomatedTellerMachine), args.Error(1)
}

func (m *MockAutomatedTellerMachineRepository) GetATMByID(atmIdentifier string) (models.AutomatedTellerMachine, error) {
	args := m.Called(atmIdentifier)
	return args.Get(0).(models.AutomatedTellerMachine), args.Error(1)
}

func TestGetATMs_AllATMsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockAutomatedTellerMachineRepository)

	expectedATMs := []models.AutomatedTellerMachine{
		{ATMIdentifier: "ATM001", StreetName: "Main St"},
		{ATMIdentifier: "ATM002", StreetName: "Second St"},
	}

	mockRepo.On("GetAllATMs").Return(expectedATMs, nil)

	controller := NewAutomatedTellerMachineController(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/atms", nil)

	controller.GetATMs(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Esto es para asegurarte de que estás comparando correctamente los tipos
	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedResponse := gin.H{
		"data": []interface{}{
			map[string]interface{}{
				"atmIdentifier":    "ATM001",
				"streetName":       "Main St",
				"atmAccessType":    "",
				"atmAttentionHour": "",
				"atmFromDatetime":  "",
				"atmServiceType":   "",
				"atmToDatetime":    "",
			},
			map[string]interface{}{
				"atmIdentifier":    "ATM002",
				"streetName":       "Second St",
				"atmAccessType":    "",
				"atmAttentionHour": "",
				"atmFromDatetime":  "",
				"atmServiceType":   "",
				"atmToDatetime":    "",
			},
		},
	}

	assert.Equal(t, expectedResponse, response)
	mockRepo.AssertExpectations(t)
}

func TestGetATMs_ATMByIDSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockAutomatedTellerMachineRepository)

	// Datos simulados
	expectedATM := models.AutomatedTellerMachine{ATMIdentifier: "ATM001", StreetName: "Main St"}

	// Configurar el mock
	mockRepo.On("GetATMByID", "ATM001").Return(expectedATM, nil)

	// Crear el controlador
	controller := NewAutomatedTellerMachineController(mockRepo)

	// Crear el contexto de prueba
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: "ATM001"})
	c.Request, _ = http.NewRequest(http.MethodGet, "/atms/ATM001", nil)

	// Llamar al método del controlador
	controller.GetATMs(c)

	// Verificar el código de respuesta
	assert.Equal(t, http.StatusOK, w.Code)

	// Verificar el cuerpo de la respuesta
	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Ajustar la respuesta esperada a map[string]interface{} para la comparación
	expectedResponse := gin.H{
		"data": map[string]interface{}{
			"atmIdentifier":    "ATM001",
			"streetName":       "Main St",
			"atmAccessType":    "",
			"atmAttentionHour": "",
			"atmFromDatetime":  "",
			"atmServiceType":   "",
			"atmToDatetime":    "",
		},
	}

	assert.Equal(t, expectedResponse, response)

	// Verificar que se cumplieron las expectativas del mock
	mockRepo.AssertExpectations(t)
}

func TestGetATMs_ATMByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockAutomatedTellerMachineRepository)

	// Configurar el mock para devolver un error
	mockRepo.On("GetATMByID", "ATM001").Return(models.AutomatedTellerMachine{}, errors.New("ATM not found"))

	// Crear el controlador
	controller := NewAutomatedTellerMachineController(mockRepo)

	// Crear el contexto de prueba
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: "ATM001"})
	c.Request, _ = http.NewRequest(http.MethodGet, "/atms/ATM001", nil)

	// Llamar al método del controlador
	controller.GetATMs(c)

	// Verificar el código de respuesta
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verificar el cuerpo de la respuesta
	expectedResponse := gin.H{"error": "ATM not found"}
	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	// Verificar que se cumplieron las expectativas del mock
	mockRepo.AssertExpectations(t)
}

func TestGetATMs_AllATMsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockAutomatedTellerMachineRepository)

	// Configurar el mock para devolver un error
	mockRepo.On("GetAllATMs").Return([]models.AutomatedTellerMachine{}, errors.New("error fetching ATMs"))

	controller := NewAutomatedTellerMachineController(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/atms", nil)

	controller.GetATMs(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	expectedResponse := gin.H{"error": "Error fetching ATMs"}
	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	mockRepo.AssertExpectations(t)
}
