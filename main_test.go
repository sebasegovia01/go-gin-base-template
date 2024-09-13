package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	"github.com/sebasegovia01/base-template-go-gin/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPService es un mock para HTTPServiceInterface
type MockHTTPService struct {
	mock.Mock
}

func (m *MockHTTPService) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHTTPService) SendRequest(method, url string, headers, body map[string]string) ([]byte, error) {
	args := m.Called(method, url, headers, body)
	return args.Get(0).([]byte), args.Error(1)
}

// MockEngineRunner es un mock para EngineRunner
type MockEngineRunner struct {
	mock.Mock
}

func (m *MockEngineRunner) Run(addr ...string) error {
	args := m.Called(addr)
	return args.Error(0)
}

// TestSetupServer_Success prueba la configuración exitosa del servidor
func TestSetupServer_Success(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	mockHTTPService := new(MockHTTPService)
	mockHTTPService.On("Close").Return(nil)
	mockHTTPService.On("SendRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte("response"), nil)

	mockNewHTTPService := func(cfg *config.Config) services.HTTPServiceInterface {
		return mockHTTPService
	}

	r, err := setupServer(cfg, mockNewHTTPService)

	assert.NoError(t, err)
	assert.NotNil(t, r)
}

// TestRun_ConfigLoadError prueba el manejo de errores al cargar la configuración
func TestRun_ConfigLoadError(t *testing.T) {
	mockLoadConfig := func() (*config.Config, error) {
		return nil, errors.New("config load error")
	}

	mockSetupServer := func(
		cfg *config.Config,
		newHTTPService func(cfg *config.Config) services.HTTPServiceInterface,
	) (EngineRunner, error) {
		t.Error("setupServer should not be called")
		return nil, nil
	}

	err := run(mockLoadConfig, mockSetupServer)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error loading config: config load error")
}

// TestRun_Success prueba la ejecución exitosa de run
func TestRun_Success(t *testing.T) {
	mockConfig := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	mockEngineRunner := new(MockEngineRunner)
	mockEngineRunner.On("Run", []string{":8080"}).Return(nil)

	mockLoadConfig := func() (*config.Config, error) {
		return mockConfig, nil
	}

	mockSetupServer := func(
		cfg *config.Config,
		newHTTPService func(cfg *config.Config) services.HTTPServiceInterface,
	) (EngineRunner, error) {
		return mockEngineRunner, nil
	}

	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(nil)

	err := run(mockLoadConfig, mockSetupServer)

	assert.NoError(t, err)
	mockEngineRunner.AssertCalled(t, "Run", []string{":8080"})
}

func TestRun_DefaultServerAddress(t *testing.T) {
	mockConfig := &config.Config{
		Environment: "test",
	}

	mockLoadConfig := func() (*config.Config, error) {
		return mockConfig, nil
	}

	mockRunner := new(MockEngineRunner)
	mockRunner.On("Run", []string{":8080"}).Return(nil)

	mockSetupServer := func(
		cfg *config.Config,
		newHTTPService func(cfg *config.Config) services.HTTPServiceInterface,
	) (EngineRunner, error) {
		assert.Equal(t, ":8080", cfg.ServerAddress)
		return mockRunner, nil
	}

	oldLogPrintf := logPrintf
	logPrintf = func(format string, v ...interface{}) {}
	defer func() { logPrintf = oldLogPrintf }()

	err := run(mockLoadConfig, mockSetupServer)

	assert.NoError(t, err)
	mockRunner.AssertExpectations(t)
}

// TestRun_DefaultEnvironment prueba el uso del entorno por defecto
func TestRun_DefaultEnvironment(t *testing.T) {
	mockConfig := &config.Config{
		ServerAddress: ":8080",
		// No definimos Environment aquí para probar el valor por defecto
	}

	mockLoadConfig := func() (*config.Config, error) {
		return mockConfig, nil
	}

	mockRunner := new(MockEngineRunner)
	mockRunner.On("Run", []string{":8080"}).Return(nil)

	var capturedConfig *config.Config
	mockSetupServer := func(
		cfg *config.Config,
		newHTTPService func(cfg *config.Config) services.HTTPServiceInterface,
	) (EngineRunner, error) {
		capturedConfig = cfg
		return mockRunner, nil
	}

	oldLogPrintf := logPrintf
	var loggedMessage string
	logPrintf = func(format string, v ...interface{}) {
		loggedMessage = fmt.Sprintf(format, v...)
	}
	defer func() { logPrintf = oldLogPrintf }()

	err := run(mockLoadConfig, mockSetupServer)

	assert.NoError(t, err)
	assert.Equal(t, enums.Dev, capturedConfig.Environment, "El entorno debería ser Dev por defecto")
	expectedLogMessage := fmt.Sprintf("Server starting on port %s, environment is %s", ":8080", enums.Dev)
	assert.Equal(t, expectedLogMessage, loggedMessage, "El mensaje de log no coincide con lo esperado")
	mockRunner.AssertExpectations(t)
}

// TestSetupServer_NoRoute prueba el manejo de rutas no existentes
func TestSetupServer_NoRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	mockHTTPService := new(MockHTTPService)
	mockNewHTTPService := func(cfg *config.Config) services.HTTPServiceInterface {
		return mockHTTPService
	}

	engine, err := setupServer(cfg, mockNewHTTPService)
	assert.NoError(t, err)

	router, ok := engine.(engineRunnerAdapter)
	assert.True(t, ok)

	// Capturar logs para evitar que se impriman durante la prueba
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(os.Stderr)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/non-existent-route", nil)
	router.ServeHTTP(w, req)

	// Restaurar la salida de log estándar
	log.SetOutput(os.Stderr)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Not Found")
}

// TestMain_Success simula la ejecución exitosa de la función main
func TestMain_Success(t *testing.T) {
	// Guardar las funciones originales
	originalLoadConfigFunc := loadConfigFunc
	originalSetupServerFunc := setupServerFunc
	originalFatalfFunc := fatalfFunc
	originalLogPrintf := logPrintf

	// Restaurar las funciones originales al final del test
	defer func() {
		loadConfigFunc = originalLoadConfigFunc
		setupServerFunc = originalSetupServerFunc
		fatalfFunc = originalFatalfFunc
		logPrintf = originalLogPrintf
	}()

	// Mock de loadConfigFunc
	loadConfigFunc = func() (*config.Config, error) {
		return &config.Config{ServerAddress: ":8080", Environment: "test"}, nil
	}

	// Mock de setupServerFunc
	setupServerFunc = func(
		cfg *config.Config,
		newHTTPService func(cfg *config.Config) services.HTTPServiceInterface,
	) (EngineRunner, error) {
		mockRunner := new(MockEngineRunner)
		mockRunner.On("Run", []string{":8080"}).Return(nil)
		return mockRunner, nil
	}

	// Mock de fatalfFunc para capturar llamadas fatales
	var fatalfCalled bool
	fatalfFunc = func(format string, v ...interface{}) {
		fatalfCalled = true
	}

	// Mock de logPrintf para evitar problemas con la salida de log
	logPrintf = func(format string, v ...interface{}) {}

	// Ejecutar main
	main()

	// Verificar que no se llamó a fatalfFunc
	assert.False(t, fatalfCalled, "fatalfFunc no debería haber sido llamado")
}

func TestRun_SetupServerError(t *testing.T) {
	mockConfig := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	mockLoadConfig := func() (*config.Config, error) {
		return mockConfig, nil
	}

	expectedError := errors.New("setup server error")
	mockSetupServer := func(
		cfg *config.Config,
		newHTTPService func(cfg *config.Config) services.HTTPServiceInterface,
	) (EngineRunner, error) {
		return nil, expectedError
	}

	// Guardar la función original y restaurarla después
	originalFatalfFunc := fatalfFunc
	defer func() { fatalfFunc = originalFatalfFunc }()

	// Mock de fatalfFunc para capturar la llamada
	var fatalfCalled bool
	var fatalfMessage string
	fatalfFunc = func(format string, v ...interface{}) {
		fatalfCalled = true
		fatalfMessage = fmt.Sprintf(format, v...)
	}

	// Ejecutar run
	err := run(mockLoadConfig, mockSetupServer)

	// Verificar que fatalfFunc fue llamada con el error correcto
	assert.True(t, fatalfCalled, "fatalfFunc debería haber sido llamada")
	assert.Equal(t, expectedError.Error(), fatalfMessage, "El mensaje de error no coincide")

	// Verificar que se retornó el error esperado
	assert.Equal(t, expectedError, err, "El error retornado no coincide con el esperado")
}

var runFunc = run

func TestMain_RunError(t *testing.T) {
	// Guardar las funciones originales
	originalLoadConfigFunc := loadConfigFunc
	originalSetupServerFunc := setupServerFunc
	originalFatalfFunc := fatalfFunc
	originalRunFunc := runFunc

	// Restaurar las funciones originales al final del test
	defer func() {
		loadConfigFunc = originalLoadConfigFunc
		setupServerFunc = originalSetupServerFunc
		fatalfFunc = originalFatalfFunc
		runFunc = originalRunFunc
	}()

	// Mock de loadConfigFunc
	loadConfigFunc = func() (*config.Config, error) {
		return &config.Config{ServerAddress: ":8080"}, nil
	}

	// Mock de setupServerFunc
	setupServerFunc = func(
		cfg *config.Config,
		newHTTPService func(cfg *config.Config) services.HTTPServiceInterface,
	) (EngineRunner, error) {
		mockRunner := new(MockEngineRunner)
		mockRunner.On("Run", []string{":8080"}).Return(errors.New("run error"))
		return mockRunner, nil
	}

	// Simular un error en la función run
	expectedError := errors.New("run error")
	runFunc = run // Usar la función run real

	// Mock de fatalfFunc para capturar la llamada
	var fatalfCalled bool
	var fatalfMessage string
	fatalfFunc = func(format string, v ...interface{}) {
		fatalfCalled = true
		fatalfMessage = fmt.Sprintf(format, v...)
	}

	// Ejecutar main
	main()

	// Verificar que fatalfFunc fue llamada con el error correcto
	assert.True(t, fatalfCalled, "fatalfFunc debería haber sido llamada")
	assert.Equal(t, expectedError.Error(), fatalfMessage, "El mensaje de error no coincide")
}

// TestNewHTTPService prueba la creación de un nuevo servicio HTTP
func TestNewHTTPService(t *testing.T) {
	// Crear una configuración mock
	cfg := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	// Llamar a la función NewHTTPService
	service := NewHTTPService(cfg)

	// Verificar que el servicio no es nil
	assert.NotNil(t, service, "El servicio HTTP no debería ser nil")

	// Verificar que el servicio es del tipo correcto
	_, ok := service.(services.HTTPServiceInterface)
	assert.True(t, ok, "El servicio debería implementar HTTPServiceInterface")
}

// TestSetupServer_WithNewHTTPService prueba la configuración del servidor con un nuevo servicio HTTP
func TestSetupServer_WithNewHTTPService(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	engine, err := setupServer(cfg, NewHTTPService)

	assert.NoError(t, err)
	assert.NotNil(t, engine)
}
