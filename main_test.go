package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	"github.com/sebasegovia01/base-template-go-gin/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2/google"
)

// Mock para StorageService
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) ProcessFile(filename string) ([]*map[string]interface{}, error) {
	args := m.Called(filename)
	return args.Get(0).([]*map[string]interface{}), args.Error(1)
}

func (m *MockStorageService) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock para PubSubService
type MockPubSubService struct {
	mock.Mock
}

func (m *MockPubSubService) ExtractStorageEvent(body io.Reader) (*services.StorageEvent, error) {
	args := m.Called(body)
	return args.Get(0).(*services.StorageEvent), args.Error(1)
}

func (m *MockPubSubService) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock para PubSubPublishService
type MockPubSubPublishService struct {
	mock.Mock
}

func (m *MockPubSubPublishService) PublishMessage(message json.RawMessage) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockPubSubPublishService) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock para EngineRunner
type MockEngineRunner struct {
	mock.Mock
}

func (m *MockEngineRunner) Run(addr ...string) error {
	args := m.Called(addr)
	return args.Error(0)
}

// Mock para gin.Engine
type MockGinEngine struct {
	mock.Mock
}

// Prueba para verificar la inicialización del servidor exitosamente
func TestSetupServer_Success(t *testing.T) {
	// Configuración simulada
	cfg := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	// Mocks de servicios
	mockStorageService := new(MockStorageService)
	mockPubSubService := new(MockPubSubService)
	mockPubSubPublishService := new(MockPubSubPublishService)

	// Configurar los mocks
	mockStorageService.On("Close").Return(nil)
	mockPubSubService.On("Close").Return(nil)
	mockPubSubPublishService.On("Close").Return(nil)

	// Funciones mockeadas
	mockNewStorageService := func(cfg *config.Config) (services.StorageServiceInterface, error) {
		return mockStorageService, nil
	}
	mockNewPubSubService := func(cfg *config.Config) (services.PubSubServiceInterface, error) {
		return mockPubSubService, nil
	}
	mockNewPubSubPublishService := func(cfg *config.Config) (services.PubSubPublishServiceInterface, error) {
		return mockPubSubPublishService, nil
	}

	// Ejecutar setupServer
	r, err := setupServer(cfg, mockNewStorageService, mockNewPubSubService, mockNewPubSubPublishService)

	// Verificar que no haya errores y que se retorne un *gin.Engine
	assert.NoError(t, err)
	assert.NotNil(t, r)
}

// Prueba para error en StorageService
func TestSetupServer_StorageServiceError(t *testing.T) {
	// Configuración simulada
	cfg := &config.Config{}

	// Mocks
	mockStorageService := new(MockStorageService)
	mockStorageService.On("Close").Return(nil)

	// Funciones mockeadas que devuelven error en StorageService
	mockNewStorageService := func(cfg *config.Config) (services.StorageServiceInterface, error) {
		return nil, errors.New("storage service error")
	}

	mockNewPubSubService := func(cfg *config.Config) (services.PubSubServiceInterface, error) {
		return new(MockPubSubService), nil
	}
	mockNewPubSubPublishService := func(cfg *config.Config) (services.PubSubPublishServiceInterface, error) {
		return new(MockPubSubPublishService), nil
	}

	// Ejecutar setupServer y verificar error
	r, err := setupServer(cfg, mockNewStorageService, mockNewPubSubService, mockNewPubSubPublishService)
	assert.Error(t, err)
	assert.Nil(t, r)
	assert.Contains(t, err.Error(), "storage service error")
}

// Prueba para error en PubSubService
func TestSetupServer_PubSubServiceError(t *testing.T) {
	// Configuración simulada
	cfg := &config.Config{}

	// Mocks
	mockStorageService := new(MockStorageService)
	mockStorageService.On("Close").Return(nil)

	// Funciones mockeadas
	mockNewStorageService := func(cfg *config.Config) (services.StorageServiceInterface, error) {
		return mockStorageService, nil
	}
	mockNewPubSubService := func(cfg *config.Config) (services.PubSubServiceInterface, error) {
		return nil, errors.New("pubsub service error")
	}
	mockNewPubSubPublishService := func(cfg *config.Config) (services.PubSubPublishServiceInterface, error) {
		return new(MockPubSubPublishService), nil
	}

	// Ejecutar setupServer y verificar error
	r, err := setupServer(cfg, mockNewStorageService, mockNewPubSubService, mockNewPubSubPublishService)
	assert.Error(t, err)
	assert.Nil(t, r)
	assert.Contains(t, err.Error(), "pubsub service error")
}

// Prueba para error en PubSubPublishService
func TestSetupServer_PubSubPublishServiceError(t *testing.T) {
	// Configuración simulada
	cfg := &config.Config{}

	// Mocks
	mockStorageService := new(MockStorageService)
	mockPubSubService := new(MockPubSubService)
	mockStorageService.On("Close").Return(nil)

	// Funciones mockeadas
	mockNewStorageService := func(cfg *config.Config) (services.StorageServiceInterface, error) {
		return mockStorageService, nil
	}
	mockNewPubSubService := func(cfg *config.Config) (services.PubSubServiceInterface, error) {
		return mockPubSubService, nil
	}
	mockNewPubSubPublishService := func(cfg *config.Config) (services.PubSubPublishServiceInterface, error) {
		return nil, errors.New("pubsub publish service error")
	}

	// Ejecutar setupServer y verificar error
	r, err := setupServer(cfg, mockNewStorageService, mockNewPubSubService, mockNewPubSubPublishService)
	assert.Error(t, err)
	assert.Nil(t, r)
	assert.Contains(t, err.Error(), "pubsub publish service error")
}

func TestSetupServer_StorageServiceClose(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	// Mocks de servicios
	mockStorageService := new(MockStorageService)
	mockPubSubService := new(MockPubSubService)
	mockPubSubPublishService := new(MockPubSubPublishService)

	// Configurar los mocks para que Close() sea llamado
	mockStorageService.On("Close").Return(nil).Once()
	mockPubSubService.On("Close").Return(nil)
	mockPubSubPublishService.On("Close").Return(nil)

	// Funciones mockeadas
	mockNewStorageService := func(cfg *config.Config) (services.StorageServiceInterface, error) {
		return mockStorageService, fmt.Errorf("error initializing storage service")
	}
	mockNewPubSubService := func(cfg *config.Config) (services.PubSubServiceInterface, error) {
		return mockPubSubService, nil
	}
	mockNewPubSubPublishService := func(cfg *config.Config) (services.PubSubPublishServiceInterface, error) {
		return mockPubSubPublishService, nil
	}

	_, err := setupServer(cfg, mockNewStorageService, mockNewPubSubService, mockNewPubSubPublishService)

	// Verificar que se haya producido el error esperado
	assert.Error(t, err)
	mockStorageService.AssertCalled(t, "Close")
}

func TestSetupServer_PubSubPublishServiceClose(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	// Mocks de servicios
	mockStorageService := new(MockStorageService)
	mockPubSubService := new(MockPubSubService)
	mockPubSubPublishService := new(MockPubSubPublishService)

	// Configurar los mocks para que Close() sea llamado
	mockStorageService.On("Close").Return(nil)
	mockPubSubService.On("Close").Return(nil)
	mockPubSubPublishService.On("Close").Return(nil).Once()

	// Funciones mockeadas
	mockNewStorageService := func(cfg *config.Config) (services.StorageServiceInterface, error) {
		return mockStorageService, nil
	}
	mockNewPubSubService := func(cfg *config.Config) (services.PubSubServiceInterface, error) {
		return mockPubSubService, nil
	}
	mockNewPubSubPublishService := func(cfg *config.Config) (services.PubSubPublishServiceInterface, error) {
		return mockPubSubPublishService, fmt.Errorf("error initializing PubSub publish service")
	}

	_, err := setupServer(cfg, mockNewStorageService, mockNewPubSubService, mockNewPubSubPublishService)

	// Verificar que se haya producido el error esperado
	assert.Error(t, err)
	mockPubSubPublishService.AssertCalled(t, "Close")
}

func TestRun_ConfigLoadError(t *testing.T) {
	// Mock de la función loadConfig que devuelve un error
	mockLoadConfig := func() (*config.Config, error) {
		return nil, errors.New("config load error")
	}

	// Mock de setupServer (no debería ser llamada en este caso)
	mockSetupServer := func(
		cfg *config.Config,
		newStorageService func(cfg *config.Config) (services.StorageServiceInterface, error),
		newPubSubService func(cfg *config.Config) (services.PubSubServiceInterface, error),
		newPubSubPublishService func(cfg *config.Config) (services.PubSubPublishServiceInterface, error),
	) (EngineRunner, error) {
		t.Error("setupServer should not be called")
		return nil, nil
	}

	// Ejecutar run
	err := run(mockLoadConfig, mockSetupServer)

	// Verificar que se produjo el error esperado
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error loading config: config load error")
}

func TestMain_Success(t *testing.T) {
	// Configuración de los mocks
	mockConfig := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	// Mock del servidor
	mockEngineRunner := new(MockEngineRunner)
	mockEngineRunner.On("Run", []string{":8080"}).Return(nil)

	// Mock de funciones de servicio
	mockLoadConfig := func() (*config.Config, error) {
		return mockConfig, nil
	}
	mockSetupServer := func(
		cfg *config.Config,
		newStorageService func(cfg *config.Config) (services.StorageServiceInterface, error),
		newPubSubService func(cfg *config.Config) (services.PubSubServiceInterface, error),
		newPubSubPublishService func(cfg *config.Config) (services.PubSubPublishServiceInterface, error),
	) (EngineRunner, error) {
		return mockEngineRunner, nil
	}

	// Capturar la salida de log
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(nil)

	// Cambiar las funciones globales
	loadConfigFunc = mockLoadConfig
	setupServerFunc = mockSetupServer

	// Ejecutar main en un goroutine para capturar el pánico si ocurre
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Verificar que el pánico contiene el mensaje esperado
				logOutputStr := logOutput.String()
				assert.Contains(t, logOutputStr, "FATAL")
				assert.Contains(t, logOutputStr, "expected error message") // Ajustar según el mensaje esperado
			}
		}()
		main()
	}()

	// Verificar que se llamó a Run con la dirección correcta
	mockEngineRunner.AssertCalled(t, "Run", []string{":8080"})
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Fatalf(format string, v ...interface{}) {
	m.Called(format, v)
}

func TestRun_FatalfCoverage(t *testing.T) {
	// Configuración de los mocks
	mockConfig := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	mockLoadConfig := func() (*config.Config, error) {
		return mockConfig, nil
	}

	mockSetupServer := func(
		cfg *config.Config,
		newStorageService func(cfg *config.Config) (services.StorageServiceInterface, error),
		newPubSubService func(cfg *config.Config) (services.PubSubServiceInterface, error),
		newPubSubPublishService func(cfg *config.Config) (services.PubSubPublishServiceInterface, error),
	) (EngineRunner, error) {
		return nil, errors.New("setup server error")
	}

	// Crear y configurar el mock logger
	mockLogger := new(MockLogger)
	mockLogger.On("Fatalf", "%v", mock.Anything).Once()

	// Reemplazar la función fatalfFunc con nuestro mock
	originalFatalf := fatalfFunc
	fatalfFunc = mockLogger.Fatalf
	defer func() { fatalfFunc = originalFatalf }()

	// Ejecutar la función run
	err := run(mockLoadConfig, mockSetupServer)

	// Verificar que se llamó a Fatalf
	mockLogger.AssertExpectations(t)

	// Verificar que se retornó un error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setup server error")
}

// mockEngineRunner es un mock para EngineRunner
type mockEngineRunner struct {
	mock.Mock
}

func (m *mockEngineRunner) Run(addr ...string) error {
	args := m.Called(addr[0])
	return args.Error(0)
}

func TestRun_DefaultServerAddress(t *testing.T) {
	// Configuración de los mocks
	mockConfig := &config.Config{
		Environment: "test",
		// No definimos ServerAddress aquí
	}

	mockLoadConfig := func() (*config.Config, error) {
		return mockConfig, nil
	}

	// Mock para EngineRunner
	mockRunner := new(mockEngineRunner)
	mockRunner.On("Run", ":8080").Return(nil)

	var capturedConfig *config.Config
	mockSetupServer := func(
		cfg *config.Config,
		newStorageService func(cfg *config.Config) (services.StorageServiceInterface, error),
		newPubSubService func(cfg *config.Config) (services.PubSubServiceInterface, error),
		newPubSubPublishService func(cfg *config.Config) (services.PubSubPublishServiceInterface, error),
	) (EngineRunner, error) {
		capturedConfig = cfg
		return mockRunner, nil
	}

	// Mock para logPrintf
	oldLogPrintf := logPrintf
	logPrintf = func(format string, v ...interface{}) {}
	defer func() { logPrintf = oldLogPrintf }()

	// Ejecutar la función run
	err := run(mockLoadConfig, mockSetupServer)

	// Verificar que no hubo errores
	assert.NoError(t, err)

	// Verificar que se usó la dirección por defecto
	assert.Equal(t, ":8080", capturedConfig.ServerAddress, "La dirección del servidor debería ser :8080 por defecto")

	// Verificar que se llamó a Run con la dirección correcta
	mockRunner.AssertExpectations(t)
}

func TestRun_DefaultEnvironment(t *testing.T) {
	// Configuración de los mocks
	mockConfig := &config.Config{
		ServerAddress: ":8080",
		// No definimos Environment aquí
	}

	mockLoadConfig := func() (*config.Config, error) {
		return mockConfig, nil
	}

	// Mock para EngineRunner
	mockRunner := new(mockEngineRunner)
	mockRunner.On("Run", ":8080").Return(nil)

	var capturedConfig *config.Config
	mockSetupServer := func(
		cfg *config.Config,
		newStorageService func(cfg *config.Config) (services.StorageServiceInterface, error),
		newPubSubService func(cfg *config.Config) (services.PubSubServiceInterface, error),
		newPubSubPublishService func(cfg *config.Config) (services.PubSubPublishServiceInterface, error),
	) (EngineRunner, error) {
		capturedConfig = cfg
		return mockRunner, nil
	}

	// Mock para logPrintf
	oldLogPrintf := logPrintf
	var loggedMessage string
	logPrintf = func(format string, v ...interface{}) {
		loggedMessage = fmt.Sprintf(format, v...)
	}
	defer func() { logPrintf = oldLogPrintf }()

	// Ejecutar la función run
	err := run(mockLoadConfig, mockSetupServer)

	// Verificar que no hubo errores
	assert.NoError(t, err)

	// Verificar que se usó el entorno por defecto
	assert.Equal(t, enums.Dev, capturedConfig.Environment, "El entorno debería ser Dev por defecto")

	// Verificar que el mensaje de log contiene el entorno correcto
	assert.Contains(t, loggedMessage, "environment is dev", "El mensaje de log debería mencionar el entorno dev")

	// Verificar que se llamó a Run con la dirección correcta
	mockRunner.AssertExpectations(t)
}

func TestSetupServer_NoRoute(t *testing.T) {
	// Crear una configuración mock
	mockConfig := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	// Crear mocks para los servicios
	mockStorageService := new(MockStorageService)
	mockPubSubService := new(MockPubSubService)
	mockPubSubPublishService := new(MockPubSubPublishService)

	// Configurar los mocks de servicios para que no devuelvan errores
	mockNewStorageService := func(cfg *config.Config) (services.StorageServiceInterface, error) {
		return mockStorageService, nil
	}
	mockNewPubSubService := func(cfg *config.Config) (services.PubSubServiceInterface, error) {
		return mockPubSubService, nil
	}
	mockNewPubSubPublishService := func(cfg *config.Config) (services.PubSubPublishServiceInterface, error) {
		return mockPubSubPublishService, nil
	}

	// Ejecutar setupServer
	engine, err := setupServer(mockConfig, mockNewStorageService, mockNewPubSubService, mockNewPubSubPublishService)
	assert.NoError(t, err)

	// Crear un router de prueba
	router, ok := engine.(engineRunnerAdapter)
	assert.True(t, ok)

	// Crear una solicitud de prueba a una ruta inexistente
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/non-existent-route", nil)

	// Capturar logs para evitar que se impriman durante la prueba
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(os.Stderr)

	// Ejecutar la solicitud
	router.ServeHTTP(w, req)

	// Restaurar la salida de log estándar
	log.SetOutput(os.Stderr)

	// Imprimir logs capturados para depuración
	t.Logf("Captured logs: %s", logBuffer.String())

	// Imprimir la respuesta completa para depuración
	t.Logf("Response Status: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	// Verificar el código de estado
	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")

	// Intentar decodificar la respuesta JSON
	var response struct {
		Result struct {
			Status         string `json:"status"`
			CanonicalError struct {
				Code        string `json:"code"`
				Type        string `json:"type"`
				Description string `json:"description"`
			} `json:"CanonicalError"`
			SourceError struct {
				Code               string `json:"code"`
				Description        string `json:"description"`
				ErrorSourceDetails struct {
					Source string `json:"source"`
				} `json:"ErrorSourceDetails"`
			} `json:"SourceError"`
		} `json:"Result"`
	}

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verificar la estructura de la respuesta
	assert.Equal(t, "ERROR", response.Result.Status)
	assert.Equal(t, "404", response.Result.CanonicalError.Code)
	assert.Equal(t, "TEC", response.Result.CanonicalError.Type)
	assert.Equal(t, "Not Found", response.Result.CanonicalError.Description)
	assert.Equal(t, "404", response.Result.SourceError.Code)
	assert.Equal(t, "An unexpected error occurred", response.Result.SourceError.Description)
	assert.Equal(t, "API", response.Result.SourceError.ErrorSourceDetails.Source)
}

func TestNewStorageServiceInterface(t *testing.T) {
	// Crear una configuración de prueba con los valores necesarios
	cfg := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
		BucketName:    "test-bucket",
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
			JSON:      []byte(`{"type": "service_account"}`), // Simula credenciales JSON
		},
	}

	// Llamar a la función NewStorageServiceInterface
	service, err := NewStorageServiceInterface(cfg)

	// Verificar que no haya errores
	assert.NoError(t, err)
	assert.NotNil(t, service, "El servicio de almacenamiento no debería ser nulo")
}

func TestNewPubSubServiceInterface(t *testing.T) {
	// Crear una configuración de prueba con los valores necesarios
	cfg := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
			JSON:      []byte(`{"type": "service_account"}`), // Simula credenciales JSON
		},
	}

	// Llamar a la función NewPubSubServiceInterface
	service, err := NewPubSubServiceInterface(cfg)

	// Verificar que no haya errores
	assert.NoError(t, err)
	assert.NotNil(t, service, "El servicio de PubSub no debería ser nulo")

	// Verificar que el servicio tenga la configuración correcta
	assert.Equal(t, "test-project", cfg.GCPCredentials.ProjectID, "El ProjectID debería coincidir")
}

func TestNewPubSubPublishServiceInterface(t *testing.T) {
	// Crear una configuración mock
	mockConfig := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
			JSON:      []byte("{}"),
		},
		Topics: []string{"test-topic"},
	}

	// Guardar la función original
	originalNewPubSubPublishServiceFunc := newPubSubPublishServiceFunc
	// Restaurar la función original después de la prueba
	defer func() { newPubSubPublishServiceFunc = originalNewPubSubPublishServiceFunc }()

	// Reemplazar newPubSubPublishServiceFunc con una función mock
	newPubSubPublishServiceFunc = func(cfg *config.Config) (*services.PubSubPublishService, error) {
		// Aquí podríamos crear un mock de PubSubPublishService si fuera necesario
		return &services.PubSubPublishService{}, nil
	}

	// Llamar a la función que queremos probar
	service, err := NewPubSubPublishServiceInterface(mockConfig)

	// Verificar que no hay error
	assert.NoError(t, err)

	// Verificar que se devuelve un servicio
	assert.NotNil(t, service)

	// Verificar que el servicio es del tipo correcto
	assert.IsType(t, &services.PubSubPublishService{}, service)
}

func TestRun_FatalfFinish(t *testing.T) {
	// Configuración de los mocks
	mockConfig := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	// Simulamos la carga de configuración sin error
	mockLoadConfig := func() (*config.Config, error) {
		return mockConfig, nil
	}

	// Simulamos un error en el setupServer para forzar que se llame a fatalfFunc
	mockSetupServer := func(
		cfg *config.Config,
		newStorageService func(cfg *config.Config) (services.StorageServiceInterface, error),
		newPubSubService func(cfg *config.Config) (services.PubSubServiceInterface, error),
		newPubSubPublishService func(cfg *config.Config) (services.PubSubPublishServiceInterface, error),
	) (EngineRunner, error) {
		return nil, errors.New("setup server error")
	}

	// Crear y configurar el mock logger para fatalfFunc
	mockLogger := new(MockLogger)

	// Mock para fatalfFunc, asegurándonos de que simplemente lance un panic
	mockLogger.On("Fatalf", "%v", mock.Anything).Run(func(args mock.Arguments) {
		panic(fmt.Sprintf(args.String(0), args.Get(1)))
	}).Once()

	// Reemplazar la función fatalfFunc con nuestro mock
	originalFatalf := fatalfFunc
	fatalfFunc = mockLogger.Fatalf
	defer func() { fatalfFunc = originalFatalf }() // Restaurar después de la prueba

	// Ejecutar la función run y validar que cause un panic
	assert.Panics(t, func() {
		_ = run(mockLoadConfig, mockSetupServer)
	})
}
