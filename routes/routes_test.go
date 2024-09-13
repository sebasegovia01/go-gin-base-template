package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/controllers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para HTTPServiceInterface
type MockHTTPService struct {
	mock.Mock
}

// Actualizamos la firma de SendRequest para que coincida con la esperada por HTTPController
func (m *MockHTTPService) SendRequest(method string, url string, headers map[string]string, queryParams map[string]string) ([]byte, error) {
	args := m.Called(method, url, headers, queryParams)
	return args.Get(0).([]byte), args.Error(1)
}

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Crear un router de prueba
	r := gin.New()

	// Crear mocks para el servicio HTTP
	mockHTTPService := new(MockHTTPService)

	// Crear una configuración de prueba
	testConfig := &config.Config{
		UrlMsAutomaticTellerMachines: "http://test-url-automated-teller-machines",
		UrlMsPresentialChannels:      "http://test-url-presential-channels",
	}

	// Crear una instancia real de HTTPController con los mocks y la configuración
	httpController := controllers.NewHTTPController(mockHTTPService, testConfig)

	// Configurar las rutas con el controlador real
	SetupRoutes(r, httpController)

	// Headers requeridos por el middleware
	headers := map[string]string{
		"Consumer-Sys-Code":          "CHL-SIT-SEG",
		"Consumer-Enterprise-Code":   "BANCORIPLEY-CHL",
		"Consumer-Country-Code":      "CHL",
		"Trace-Client-Req-Timestamp": "2024-09-05 12:00:00.123456+0000",
		"Trace-Source-Id":            "123e4567-e89b-12d3-a456-426614174000",
		"Channel-Name":               "SEGUROS",
		"Channel-Mode":               "PRESENCIAL",
	}

	// Simular comportamiento del servicio HTTP para cajeros automáticos
	mockHTTPService.On("SendRequest", "GET", "http://test-url-automated-teller-machines/", mock.Anything, mock.Anything).Return([]byte(`{"message": "ATM data"}`), nil)

	// Probar la ruta GET /automated-teller-machines/
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/service-channels/v1/api/automated-teller-machines/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(w, req)

	// Verificar respuesta
	assert.Equal(t, http.StatusOK, w.Code)
	mockHTTPService.AssertCalled(t, "SendRequest", "GET", "http://test-url-automated-teller-machines/", mock.Anything, mock.Anything)

	// Simular comportamiento del servicio HTTP para canales presenciales
	mockHTTPService.On("SendRequest", "GET", "http://test-url-presential-channels/", mock.Anything, mock.Anything).Return([]byte(`{"message": "Presential Channel data"}`), nil)

	// Probar la ruta GET /presentialchannels/
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/service-channels/v1/api/presentialchannels/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(w, req)

	// Verificar respuesta
	assert.Equal(t, http.StatusOK, w.Code)
	mockHTTPService.AssertCalled(t, "SendRequest", "GET", "http://test-url-presential-channels/", mock.Anything, mock.Anything)
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Crear un router de prueba
	r := gin.New()

	// Crear una instancia real de HealthController
	healthController := controllers.NewHealthController()

	// Configurar la ruta de health check
	r.GET("/service-channels/v1/api/health", healthController.HealthCheck)

	// Probar la ruta GET /health
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/service-channels/v1/api/health", nil)
	r.ServeHTTP(w, req)

	// Verificar la respuesta de health check
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"UP","message":"API is healthy"}`, w.Body.String())
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
