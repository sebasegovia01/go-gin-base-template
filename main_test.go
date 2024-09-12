package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	"github.com/sebasegovia01/base-template-go-gin/models"
	"github.com/sebasegovia01/base-template-go-gin/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para AutomatedTellerMachineInterface
type MockAutomatedTellerMachineClient struct {
	mock.Mock
}

func (m *MockAutomatedTellerMachineClient) GetAllATMs() ([]models.AutomatedTellerMachine, error) {
	args := m.Called()
	return args.Get(0).([]models.AutomatedTellerMachine), args.Error(1)
}

func (m *MockAutomatedTellerMachineClient) GetATMByID(atmIdentifier string) (models.AutomatedTellerMachine, error) {
	args := m.Called(atmIdentifier)
	return args.Get(0).(models.AutomatedTellerMachine), args.Error(1)
}

func (m *MockAutomatedTellerMachineClient) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	args := m.Called(ctx, key, dst)
	return args.Error(0)
}

func (m *MockAutomatedTellerMachineClient) GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error) {
	args := m.Called(ctx, q, dst)
	return args.Get(0).([]*datastore.Key), args.Error(1)
}

func (m *MockAutomatedTellerMachineClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock para PresentialChannelInterface
type MockPresentialChannelClient struct {
	mock.Mock
}

// Mock para EngineRunner
type MockEngineRunner struct {
	mock.Mock
}

func (m *MockEngineRunner) Run(addr ...string) error {
	// Cambia el argumento aquí para usar el primer elemento del slice
	args := m.Called(addr[0]) // En lugar de pasar todo el slice, solo pasa el primer elemento
	return args.Error(0)
}

// Mock para gin.Engine
type MockGinEngine struct {
	mock.Mock
}

func (m *MockPresentialChannelClient) GetAllPresentialChannels() ([]models.PresentialChannel, error) {
	args := m.Called()
	return args.Get(0).([]models.PresentialChannel), args.Error(1)
}

func (m *MockPresentialChannelClient) GetPresentialChannelByID(channelIdentifier string) (models.PresentialChannel, error) {
	args := m.Called(channelIdentifier)
	return args.Get(0).(models.PresentialChannel), args.Error(1)
}

func (m *MockPresentialChannelClient) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	args := m.Called(ctx, key, dst)
	return args.Error(0)
}

func (m *MockPresentialChannelClient) GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error) {
	args := m.Called(ctx, q, dst)
	return args.Get(0).([]*datastore.Key), args.Error(1)
}

func (m *MockPresentialChannelClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Prueba para verificar la inicialización del servidor exitosamente
func TestSetupServer_Success(t *testing.T) {
	// Configuración simulada
	cfg := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	// Mocks de clientes Datastore
	mockATMClient := new(MockAutomatedTellerMachineClient)
	mockPresentialChannelClient := new(MockPresentialChannelClient)

	// Configurar las funciones mockeadas para que devuelvan los clientes mock
	mockNewAutomatedTellerMachineClient := func(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error) {
		return mockATMClient, nil
	}
	mockNewPresentialChannelClient := func(cfg *config.Config) (repositories.PresentialChannelInterface, error) {
		return mockPresentialChannelClient, nil
	}

	// Ejecutar setupServer
	r, err := setupServer(cfg, mockNewAutomatedTellerMachineClient, mockNewPresentialChannelClient)

	// Verificar que no haya errores y que se retorne un *gin.Engine
	assert.NoError(t, err)
	assert.NotNil(t, r)
}

func TestSetupServer_AutomatedTellerMachineClientError(t *testing.T) {
	// Configuración simulada
	cfg := &config.Config{}

	// Mocks para los clientes
	//mockATMClient := new(MockAutomatedTellerMachineClient)
	mockPresentialChannelClient := new(MockPresentialChannelClient)

	// Funciones mockeadas que devuelven un error en la creación de AutomatedTellerMachineClient
	mockNewAutomatedTellerMachineClient := func(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error) {
		return nil, errors.New("datastore ATM client error")
	}
	mockNewPresentialChannelClient := func(cfg *config.Config) (repositories.PresentialChannelInterface, error) {
		return mockPresentialChannelClient, nil
	}

	// Ejecutar setupServer y verificar error
	r, err := setupServer(cfg, mockNewAutomatedTellerMachineClient, mockNewPresentialChannelClient)
	assert.Error(t, err)
	assert.Nil(t, r)
	assert.Contains(t, err.Error(), "datastore ATM client error")
}

func TestSetupServer_PresentialChannelClientError(t *testing.T) {
	// Configuración simulada
	cfg := &config.Config{}

	// Mocks para los clientes
	mockATMClient := new(MockAutomatedTellerMachineClient)

	// Funciones mockeadas que devuelven un error en la creación de PresentialChannelClient
	mockNewAutomatedTellerMachineClient := func(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error) {
		return mockATMClient, nil
	}
	mockNewPresentialChannelClient := func(cfg *config.Config) (repositories.PresentialChannelInterface, error) {
		return nil, errors.New("datastore Presential Channel client error")
	}

	// Ejecutar setupServer y verificar error
	r, err := setupServer(cfg, mockNewAutomatedTellerMachineClient, mockNewPresentialChannelClient)
	assert.Error(t, err)
	assert.Nil(t, r)
	assert.Contains(t, err.Error(), "datastore Presential Channel client error")
}

func TestRun_ConfigLoadError(t *testing.T) {
	// Mock de la función loadConfig que devuelve un error
	mockLoadConfig := func() (*config.Config, error) {
		return nil, errors.New("config load error")
	}

	// Mock de setupServer (no debería ser llamada en este caso)
	mockSetupServer := func(
		cfg *config.Config,
		newAutomatedTellerMachineClient func(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error),
		newPresentialChannelClient func(cfg *config.Config) (repositories.PresentialChannelInterface, error),
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
	mockRunner := new(MockEngineRunner)
	mockRunner.On("Run", ":8080").Return(nil) // Cambia aquí para esperar un string, no un slice

	var capturedConfig *config.Config
	mockSetupServer := func(
		cfg *config.Config,
		newAutomatedTellerMachineClient func(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error),
		newPresentialChannelClient func(cfg *config.Config) (repositories.PresentialChannelInterface, error),
	) (EngineRunner, error) {
		capturedConfig = cfg
		return mockRunner, nil
	}

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
	mockRunner := new(MockEngineRunner)
	mockRunner.On("Run", ":8080").Return(nil)

	var capturedConfig *config.Config
	mockSetupServer := func(
		cfg *config.Config,
		newAutomatedTellerMachineClient func(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error),
		newPresentialChannelClient func(cfg *config.Config) (repositories.PresentialChannelInterface, error),
	) (EngineRunner, error) {
		capturedConfig = cfg
		return mockRunner, nil
	}

	// Ejecutar la función run
	err := run(mockLoadConfig, mockSetupServer)

	// Verificar que no hubo errores
	assert.NoError(t, err)

	// Verificar que se usó el entorno por defecto
	assert.Equal(t, enums.Dev, capturedConfig.Environment, "El entorno debería ser Dev por defecto")
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Fatalf(format string, v ...interface{}) {
	m.Called(format, v)
}

func TestRun_FatalfWhenSetupServerFails(t *testing.T) {
	// Configuración de los mocks
	mockConfig := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	// Simulamos un error en el setupServer para forzar que se llame a fatalfFunc
	mockSetupServer := func(
		cfg *config.Config,
		newAutomatedTellerMachineClient func(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error),
		newPresentialChannelClient func(cfg *config.Config) (repositories.PresentialChannelInterface, error),
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
	err := run(func() (*config.Config, error) { return mockConfig, nil }, mockSetupServer)

	// Verificar que se llamó a Fatalf
	mockLogger.AssertExpectations(t)

	// Verificar que se retornó un error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setup server error")
}

func TestMain_Fatalf(t *testing.T) {
	// Simular un error en la función run para forzar la llamada a fatalfFunc
	mockLoadConfig := func() (*config.Config, error) {
		return nil, errors.New("config load error")
	}

	mockSetupServer := func(
		cfg *config.Config,
		newAutomatedTellerMachineClient func(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error),
		newPresentialChannelClient func(cfg *config.Config) (repositories.PresentialChannelInterface, error),
	) (EngineRunner, error) {
		return nil, nil
	}

	// Crear y configurar el mock logger para fatalfFunc
	mockLogger := new(MockLogger)

	// Mock para fatalfFunc, asegurándonos de que simplemente lance un panic
	mockLogger.On("Fatalf", "%v", "error loading config: config load error").Run(func(args mock.Arguments) {
		panic(fmt.Sprintf(args.String(0), args.Get(1)))
	}).Once()

	// Reemplazar las funciones globales con mocks
	originalLoadConfigFunc := loadConfigFunc
	originalSetupServerFunc := setupServerFunc
	originalFatalfFunc := fatalfFunc

	loadConfigFunc = mockLoadConfig
	setupServerFunc = mockSetupServer
	fatalfFunc = mockLogger.Fatalf

	defer func() {
		// Restaurar las funciones originales después de la prueba
		loadConfigFunc = originalLoadConfigFunc
		setupServerFunc = originalSetupServerFunc
		fatalfFunc = originalFatalfFunc
	}()

	// Validar que el main cause un panic cuando run falla
	assert.Panics(t, func() {
		main()
	})
}

func TestSetupServer_NotFound(t *testing.T) {
	// Crear una configuración simulada
	mockConfig := &config.Config{
		ServerAddress: ":8080",
		Environment:   "test",
	}

	// Mocks para los clientes necesarios
	mockATMClient := new(MockAutomatedTellerMachineClient)
	mockPresentialChannelClient := new(MockPresentialChannelClient)

	// Crear funciones mockeadas para clientes
	mockNewAutomatedTellerMachineClient := func(cfg *config.Config) (repositories.AutomatedTellerMachineInterface, error) {
		return mockATMClient, nil
	}
	mockNewPresentialChannelClient := func(cfg *config.Config) (repositories.PresentialChannelInterface, error) {
		return mockPresentialChannelClient, nil
	}

	// Ejecutar setupServer
	engine, err := setupServer(mockConfig, mockNewAutomatedTellerMachineClient, mockNewPresentialChannelClient)
	assert.NoError(t, err)

	// Crear un router de prueba
	router, ok := engine.(engineRunnerAdapter)
	assert.True(t, ok)

	// Crear una solicitud de prueba a una ruta inexistente
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/non-existent-route", nil)

	// Ejecutar la solicitud
	router.ServeHTTP(w, req)

	// Verificar el código de estado
	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")

	// Intentar decodificar la respuesta JSON
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verificar la estructura de la respuesta actual
	expectedResponse := map[string]interface{}{
		"Result": map[string]interface{}{
			"CanonicalError": map[string]interface{}{
				"code":        "404",
				"description": "Not Found",
				"type":        "TEC",
			},
			"SourceError": map[string]interface{}{
				"code":        "404",
				"description": "An unexpected error occurred",
				"ErrorSourceDetails": map[string]interface{}{
					"source": "API",
				},
			},
			"status": "ERROR",
		},
	}

	// Verificar que la respuesta coincida con la esperada
	assert.Equal(t, expectedResponse, response)
}
