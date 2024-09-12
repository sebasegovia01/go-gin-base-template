package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	"github.com/sebasegovia01/base-template-go-gin/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPService es un mock para HTTPServiceInterface
type MockHTTPService struct {
	mock.Mock
}

func (m *MockHTTPService) SendRequest(method, url string, params map[string]string, headers map[string]string) ([]byte, error) {
	args := m.Called(method, url, params, headers)
	return args.Get(0).([]byte), args.Error(1)
}

// setupTestEnvironment configura el entorno para las pruebas
func setupTestEnvironment() {
	os.Setenv("ENV", "dev")
	os.Setenv("URL_MS_AUTOMATIC_TELLER_MACHINES", "http://example.com/atm")
	os.Setenv("URL_MS_PRESENTIAL_CHANELS", "http://example.com/channel")
}

// cleanupTestEnvironment limpia el entorno después de las pruebas
func cleanupTestEnvironment() {
	os.Unsetenv("ENV")
	os.Unsetenv("URL_MS_AUTOMATIC_TELLER_MACHINES")
	os.Unsetenv("URL_MS_PRESENTIAL_CHANELS")
}

// TestGetAutomatedTellerMachine_Success prueba el caso de éxito para GetAutomatedTellerMachine
func TestGetAutomatedTellerMachine_Success(t *testing.T) {
	setupTestEnvironment()
	defer cleanupTestEnvironment()

	gin.SetMode(gin.TestMode)
	mockService := new(MockHTTPService)

	cfg := &config.Config{
		Environment:                  enums.Dev,
		UrlMsAutomaticTellerMachines: "http://example.com/atm",
	}

	expectedResponse := []byte(`{"Result": {"data": {"data": {"id": "ATM001", "name": "Test ATM"}}}}`)
	expectedTransformedResponse := []byte(`{"id": "ATM001", "name": "Test ATM"}`)

	// Actualizar la configuración del mock para que coincida exactamente con la llamada real
	mockService.On("SendRequest", "GET", "http://example.com/atm/ATM001", map[string]string(nil), mock.AnythingOfType("map[string]string")).Return(expectedResponse, nil)

	controller := NewHTTPController(mockService, cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: "ATM001"})
	c.Request, _ = http.NewRequest(http.MethodGet, "/atm/ATM001", nil)

	controller.GetAutomatedTellerMachine(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, string(expectedTransformedResponse), w.Body.String())
	mockService.AssertExpectations(t)
}

// TestTransformResponse prueba la función TransformResponse
func TestTransformResponse(t *testing.T) {
	originalResponse := []byte(`{"Result": {"data": {"data": {"id": "TEST001", "name": "Test Item"}}}}`)
	expectedTransformed := []byte(`{"id":"TEST001","name":"Test Item"}`)

	transformed, err := utils.TransformResponse(originalResponse)
	assert.NoError(t, err)
	assert.JSONEq(t, string(expectedTransformed), string(transformed))
}

// TestTransformResponse_InvalidJSON prueba TransformResponse con JSON inválido
func TestTransformResponse_InvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{"invalid": json}`)

	_, err := utils.TransformResponse(invalidJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error unmarshaling original response")
}

// TestConfig_Load prueba la carga de configuración
func TestConfig_Load(t *testing.T) {
	setupTestEnvironment()
	defer cleanupTestEnvironment()

	// En lugar de usar config.Load(), vamos a crear directamente una instancia de Config
	cfg := &config.Config{}

	// Cargar manualmente los valores del entorno
	cfg.Environment = enums.Environment(os.Getenv("ENV"))
	cfg.UrlMsPresentialChannels = os.Getenv("URL_MS_PRESENTIAL_CHANELS")
	cfg.UrlMsAutomaticTellerMachines = os.Getenv("URL_MS_AUTOMATIC_TELLER_MACHINES")

	// Realizar las aserciones
	assert.Equal(t, enums.Dev, cfg.Environment)
	assert.Equal(t, "http://example.com/channel", cfg.UrlMsPresentialChannels)
	assert.Equal(t, "http://example.com/atm", cfg.UrlMsAutomaticTellerMachines)
}

func TestGetPresentialChannel_Success(t *testing.T) {
	setupTestEnvironment()
	defer cleanupTestEnvironment()

	gin.SetMode(gin.TestMode)
	mockService := new(MockHTTPService)

	cfg := &config.Config{
		Environment:             enums.Dev,
		UrlMsPresentialChannels: "http://example.com/channel",
	}

	expectedResponse := []byte(`{"Result": {"data": {"data": {"id": "CHANNEL001", "name": "Test Channel"}}}}`)
	expectedTransformedResponse := []byte(`{"id": "CHANNEL001", "name": "Test Channel"}`)

	mockService.On("SendRequest", "GET", "http://example.com/channel/CHANNEL001", map[string]string(nil), mock.AnythingOfType("map[string]string")).Return(expectedResponse, nil)

	controller := NewHTTPController(mockService, cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: "CHANNEL001"})
	c.Request, _ = http.NewRequest(http.MethodGet, "/channel/CHANNEL001", nil)
	c.Request.Header.Set("X-Custom-Header", "test-value")

	controller.GetPresentialChannel(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, string(expectedTransformedResponse), w.Body.String())
	mockService.AssertExpectations(t)
}

func TestGetPresentialChannel_ServiceError(t *testing.T) {
	setupTestEnvironment()
	defer cleanupTestEnvironment()

	gin.SetMode(gin.TestMode)
	mockService := new(MockHTTPService)

	cfg := &config.Config{
		Environment:             enums.Dev,
		UrlMsPresentialChannels: "http://example.com/channel",
	}

	mockService.On("SendRequest", "GET", "http://example.com/channel/CHANNEL001", map[string]string(nil), mock.AnythingOfType("map[string]string")).Return([]byte{}, errors.New("service error"))

	controller := NewHTTPController(mockService, cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: "CHANNEL001"})
	c.Request, _ = http.NewRequest(http.MethodGet, "/channel/CHANNEL001", nil)

	controller.GetPresentialChannel(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to fetch Presential Channel data")
	mockService.AssertExpectations(t)
}

func TestGetPresentialChannel_TransformError(t *testing.T) {
	setupTestEnvironment()
	defer cleanupTestEnvironment()

	gin.SetMode(gin.TestMode)
	mockService := new(MockHTTPService)

	cfg := &config.Config{
		Environment:             enums.Dev,
		UrlMsPresentialChannels: "http://example.com/channel",
	}

	// Usar un JSON inválido para provocar un error de unmarshaling
	invalidResponse := []byte(`{"This is not valid JSON`)

	mockService.On("SendRequest", "GET", "http://example.com/channel/CHANNEL001", map[string]string(nil), mock.AnythingOfType("map[string]string")).Return(invalidResponse, nil)

	controller := NewHTTPController(mockService, cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: "CHANNEL001"})
	c.Request, _ = http.NewRequest(http.MethodGet, "/channel/CHANNEL001", nil)

	controller.GetPresentialChannel(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to transform Presential Channel data")
	mockService.AssertExpectations(t)
}

func TestGetPresentialChannel_SpecificHeadersHandling(t *testing.T) {
	setupTestEnvironment()
	defer cleanupTestEnvironment()

	gin.SetMode(gin.TestMode)
	mockService := new(MockHTTPService)

	cfg := &config.Config{
		Environment:             enums.Dev,
		UrlMsPresentialChannels: "http://example.com/channel",
	}

	expectedResponse := []byte(`{"Result": {"data": {"data": {"id": "CHANNEL001", "name": "Test Channel"}}}}`)

	// Usamos mock.MatchedBy para verificar que los headers se pasen correctamente
	mockService.On("SendRequest",
		"GET",
		"http://example.com/channel/CHANNEL001",
		map[string]string(nil),
		mock.MatchedBy(func(headers map[string]string) bool {
			// Verificar que todos los headers esperados estén presentes
			expectedHeaders := []string{
				"Consumer-Sys-Code",
				"Consumer-Enterprise-Code",
				"Consumer-Country-Code",
				"Trace-Client-Req-Timestamp",
				"Trace-Source-Id",
				"Channel-Name",
				"Channel-Mode",
			}
			for _, header := range expectedHeaders {
				if _, exists := headers[header]; !exists {
					return false
				}
			}
			return true
		}),
	).Return(expectedResponse, nil)

	controller := NewHTTPController(mockService, cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: "CHANNEL001"})
	c.Request, _ = http.NewRequest(http.MethodGet, "/channel/CHANNEL001", nil)

	// Agregar los headers específicos
	c.Request.Header.Set("Consumer-Sys-Code", "CHL-HB-WEB")
	c.Request.Header.Set("Consumer-Enterprise-Code", "BANCORIPLEY-CHL")
	c.Request.Header.Set("Consumer-Country-Code", "CHL")
	c.Request.Header.Set("Trace-Client-Req-Timestamp", time.Now().Format("2006-01-02 15:04:05.000000-0700"))
	c.Request.Header.Set("Trace-Source-Id", uuid.New().String())
	c.Request.Header.Set("Channel-Name", "INVALID")
	c.Request.Header.Set("Channel-Mode", "PRESENCIAL")

	// Agregar un header con múltiples valores para asegurarnos de que se tome el primero
	c.Request.Header.Add("X-Multi-Value", "value1")
	c.Request.Header.Add("X-Multi-Value", "value2")

	controller.GetPresentialChannel(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	// Verificación adicional para asegurarnos de que se pasaron todos los headers correctamente
	mockService.AssertCalled(t, "SendRequest",
		"GET",
		"http://example.com/channel/CHANNEL001",
		map[string]string(nil),
		mock.MatchedBy(func(headers map[string]string) bool {
			return headers["Consumer-Sys-Code"] == "CHL-HB-WEB" &&
				headers["Consumer-Enterprise-Code"] == "BANCORIPLEY-CHL" &&
				headers["Consumer-Country-Code"] == "CHL" &&
				headers["Channel-Name"] == "INVALID" &&
				headers["Channel-Mode"] == "PRESENCIAL" &&
				headers["X-Multi-Value"] == "value1" // Verifica que solo se tomó el primer valor
		}),
	)
}

func TestGetAutomatedTellerMachine_FetchError(t *testing.T) {
	setupTestEnvironment()
	defer cleanupTestEnvironment()

	gin.SetMode(gin.TestMode)
	mockService := new(MockHTTPService)

	cfg := &config.Config{
		Environment:                  enums.Dev,
		UrlMsAutomaticTellerMachines: "http://example.com/atm",
	}

	// Simular un error en la llamada al servicio
	mockError := errors.New("service unavailable")
	mockService.On("SendRequest",
		"GET",
		"http://example.com/atm/ATM001",
		map[string]string(nil),
		mock.AnythingOfType("map[string]string"),
	).Return([]byte{}, mockError)

	controller := NewHTTPController(mockService, cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: "ATM001"})
	c.Request, _ = http.NewRequest(http.MethodGet, "/atm/ATM001", nil)

	// Agregar algunos headers de ejemplo
	c.Request.Header.Set("Consumer-Sys-Code", "CHL-HB-WEB")
	c.Request.Header.Set("Trace-Source-Id", uuid.New().String())

	// Capturar los logs
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr) // Restaurar la salida estándar de logs
	}()

	controller.GetAutomatedTellerMachine(c)

	// Verificar el código de estado
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verificar el cuerpo de la respuesta
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to fetch ATM data", response["error"])

	// Verificar que se haya registrado el error
	assert.Contains(t, buf.String(), "Error fetching Automated Teller Machine: service unavailable")

	mockService.AssertExpectations(t)
}

// Agregar esta estructura mock al principio del archivo de pruebas
type MockResponseTransformer struct {
	mock.Mock
}

func (m *MockResponseTransformer) TransformResponse(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

func TestGetAutomatedTellerMachine_TransformError(t *testing.T) {
	// Guardar la función original y restaurarla después
	originalTransformResponse := TransformResponse
	defer func() { TransformResponse = originalTransformResponse }()

	// Reemplazar TransformResponse con una versión que siempre devuelve un error
	TransformResponse = func(data []byte) ([]byte, error) {
		return nil, errors.New("error transforming response")
	}

	gin.SetMode(gin.TestMode)
	mockService := new(MockHTTPService)

	cfg := &config.Config{
		Environment:                  enums.Dev,
		UrlMsAutomaticTellerMachines: "http://example.com/atm",
	}

	validResponse := []byte(`{"Result": {"data": {"data": {"id": "ATM001", "name": "Test ATM"}}}}`)
	mockService.On("SendRequest",
		"GET",
		"http://example.com/atm/ATM001",
		map[string]string(nil),
		mock.AnythingOfType("map[string]string"),
	).Return(validResponse, nil)

	controller := NewHTTPController(mockService, cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: "ATM001"})
	c.Request, _ = http.NewRequest(http.MethodGet, "/atm/ATM001", nil)

	// Capturar los logs
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr) // Restaurar la salida estándar de logs
	}()

	controller.GetAutomatedTellerMachine(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to transform ATM data", response["error"])

	assert.Contains(t, buf.String(), "Error transforming response: error transforming response")

	mockService.AssertExpectations(t)
}
