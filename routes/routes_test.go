package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/controllers"
	"github.com/sebasegovia01/base-template-go-gin/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks para AutomatedTellerMachineRepository
type MockAutomatedTellerMachineRepo struct {
	mock.Mock
}

func (m *MockAutomatedTellerMachineRepo) GetAllATMs() ([]models.AutomatedTellerMachine, error) {
	args := m.Called()
	return args.Get(0).([]models.AutomatedTellerMachine), args.Error(1)
}

func (m *MockAutomatedTellerMachineRepo) GetATMByID(atmIdentifier string) (models.AutomatedTellerMachine, error) {
	args := m.Called(atmIdentifier)
	return args.Get(0).(models.AutomatedTellerMachine), args.Error(1)
}

// Mocks para PresentialChannelRepository
type MockPresentialChannelRepo struct {
	mock.Mock
}

func (m *MockPresentialChannelRepo) GetAllPresentialChannels() ([]models.PresentialChannel, error) {
	args := m.Called()
	return args.Get(0).([]models.PresentialChannel), args.Error(1)
}

func (m *MockPresentialChannelRepo) GetPresentialChannelByID(channelIdentifier string) (models.PresentialChannel, error) {
	args := m.Called(channelIdentifier)
	return args.Get(0).(models.PresentialChannel), args.Error(1)
}

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Crear un router de prueba
	r := gin.New()

	// Crear mocks para los repositorios
	mockATMRepo := new(MockAutomatedTellerMachineRepo)
	mockPresentialChannelRepo := new(MockPresentialChannelRepo)

	// Crear instancias reales de los controladores con los repositorios mock
	atmController := controllers.NewAutomatedTellerMachineController(mockATMRepo)
	presentialChannelController := controllers.NewPresentialChannelController(mockPresentialChannelRepo)

	// Configurar las rutas con los controladores
	SetupRoutes(r, atmController, presentialChannelController)

	// Prueba de la ruta de HealthCheck
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/service-channels/v1/api/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"UP","message":"API is healthy"}`, w.Body.String())

	// Prueba de la ruta /automated-teller-machines/
	mockATMRepo.On("GetAllATMs").Return([]models.AutomatedTellerMachine{
		{
			ATMIdentifier: "ATM001",
			StreetName:    "Main Street",
		},
	}, nil)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/service-channels/v1/api/automated-teller-machines/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Prueba de la ruta /presentialchannels/
	mockPresentialChannelRepo.On("GetAllPresentialChannels").Return([]models.PresentialChannel{
		{
			PresentialChannelIdentifier: "CH001",
			StreetName:                  "Channel One",
		},
	}, nil)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/service-channels/v1/api/presentialchannels/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWithTraceability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Crear un router de prueba
	r := gin.New()

	// Definir un handler de prueba
	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "Handler executed"})
	}

	// Registrar la ruta con el middleware `WithTraceability`
	r.GET("/test", WithTraceability(testHandler))

	// Crear una solicitud de prueba
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)

	// Añadir los headers requeridos por el middleware
	req.Header.Set("Consumer-Sys-Code", "CHL-SIT-SEG")
	req.Header.Set("Consumer-Enterprise-Code", "BANCORIPLEY-CHL")
	req.Header.Set("Consumer-Country-Code", "CHL")
	req.Header.Set("Trace-Client-Req-Timestamp", "2024-09-05 12:00:00.123456+0000")
	req.Header.Set("Trace-Source-Id", "123e4567-e89b-12d3-a456-426614174000")
	req.Header.Set("Channel-Name", "SEGUROS")
	req.Header.Set("Channel-Mode", "PRESENCIAL")

	// Ejecutar la solicitud
	r.ServeHTTP(w, req)

	// Verificar que el middleware no haya abortado y que el handler fue ejecutado
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, handlerCalled, "El handler debería haber sido llamado")

	// Verificar que la respuesta sea JSON válida
	assert.JSONEq(t, `{"message": "Handler executed"}`, w.Body.String())
}
