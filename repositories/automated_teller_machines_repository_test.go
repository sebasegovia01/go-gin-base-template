package repositories

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// Mock para AutomatedTellerMachineInterface
type mockAutomatedTellerMachineClient struct {
	mock.Mock
}

func (m *mockAutomatedTellerMachineClient) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	args := m.Called(ctx, key, dst)
	return args.Error(0)
}

func (m *mockAutomatedTellerMachineClient) GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error) {
	args := m.Called(ctx, q, dst)
	return args.Get(0).([]*datastore.Key), args.Error(1)
}

func (m *mockAutomatedTellerMachineClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock para AutomatedTellerMachineCreator
type mockAutomatedTellerMachineCreator struct {
	mock.Mock
}

func (m *mockAutomatedTellerMachineCreator) NewClientWithDatabase(ctx context.Context, projectID string, databaseID string, opts ...option.ClientOption) (AutomatedTellerMachineInterface, error) {
	args := m.Called(ctx, projectID, databaseID, opts)
	return args.Get(0).(AutomatedTellerMachineInterface), args.Error(1)
}

func TestNewAutomatedTellerMachineClient_Success(t *testing.T) {
	mockCreator := new(mockAutomatedTellerMachineCreator)
	mockClient := new(mockAutomatedTellerMachineClient)

	// Establecer un ProjectID válido en la configuración
	cfg := &config.Config{
		GCPCredentials:  &google.Credentials{},
		DataStoreDBName: "test-db",
	}

	// Establecer el ProjectID en el mock para simular la creación del cliente
	mockCreator.On("NewClientWithDatabase", mock.Anything, "test-project", "test-db", mock.Anything).Return(mockClient, nil)

	// Asignar manualmente el ProjectID en la prueba (ya que no está en la configuración)
	cfg.GCPCredentials.ProjectID = "test-project"

	client, err := NewAutomatedTellerMachineClient(cfg, mockCreator)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	mockCreator.AssertExpectations(t)
}

func TestNewAutomatedTellerMachineClient_Error(t *testing.T) {
	mockCreator := new(mockAutomatedTellerMachineCreator)
	mockClient := new(mockAutomatedTellerMachineClient) // Crear un cliente simulado vacío

	// Establecer un ProjectID válido en la configuración
	cfg := &config.Config{
		GCPCredentials:  &google.Credentials{},
		DataStoreDBName: "test-db",
	}

	// Simular que se devuelve un cliente "válido" pero que ocurre un error lógico
	mockCreator.On("NewClientWithDatabase", mock.Anything, "test-project", "test-db", mock.Anything).Return(mockClient, fmt.Errorf("client error"))

	// Asignar manualmente el ProjectID en la prueba
	cfg.GCPCredentials.ProjectID = "test-project"

	client, err := NewAutomatedTellerMachineClient(cfg, mockCreator)

	// Verificar que se reciba el error correctamente
	assert.Error(t, err)
	assert.Nil(t, client) // El cliente debe ser nulo en caso de error
	assert.Contains(t, err.Error(), "failed to create Datastore client")

	mockCreator.AssertExpectations(t)
}

func TestAutomatedTellerMachineRepository_GetAllATMs_Success(t *testing.T) {
	mockClient := new(mockAutomatedTellerMachineClient)
	repo := NewDatastoreATMRepository(mockClient, "test-db", "test-namespace", "ATM")

	var expectedATMs = []models.AutomatedTellerMachine{
		{ATMIdentifier: "ATM001", StreetName: "Main St"},
		{ATMIdentifier: "ATM002", StreetName: "Second St"},
	}

	// Simular la respuesta de GetAll con una lista vacía de claves
	mockClient.On("GetAll", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dst := args.Get(2).(*[]models.AutomatedTellerMachine)
		*dst = expectedATMs
	}).Return([]*datastore.Key{}, nil).Once()

	atms, err := repo.GetAllATMs()
	assert.NoError(t, err)
	assert.Equal(t, expectedATMs, atms)

	mockClient.AssertExpectations(t)
}

func TestAutomatedTellerMachineRepository_GetAllATMs_Error(t *testing.T) {
	mockClient := new(mockAutomatedTellerMachineClient)
	repo := NewDatastoreATMRepository(mockClient, "test-db", "test-namespace", "ATM")

	// Simular error en GetAll
	mockClient.On("GetAll", mock.Anything, mock.Anything, mock.Anything).Return([]*datastore.Key{}, fmt.Errorf("datastore error")).Once()

	atms, err := repo.GetAllATMs()
	assert.Error(t, err)
	assert.Nil(t, atms)
	assert.Contains(t, err.Error(), "failed to get all ATMs")

	mockClient.AssertExpectations(t)
}
func TestAutomatedTellerMachineRepository_GetATMByID_Success(t *testing.T) {
	// Configuración de prueba
	mockClient := new(mockAutomatedTellerMachineClient)
	repo := NewDatastoreATMRepository(mockClient, "test-db", "test-namespace", "ATM")

	expectedATM := models.AutomatedTellerMachine{ATMIdentifier: "ATM001", StreetName: "Main St"}

	// Mock para el método Get
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dst := args.Get(2).(*models.AutomatedTellerMachine)
		*dst = expectedATM
	}).Return(nil).Once()

	atm, err := repo.GetATMByID("ATM001")
	assert.NoError(t, err)
	assert.Equal(t, expectedATM, atm)

	mockClient.AssertExpectations(t)
}

func TestAutomatedTellerMachineRepository_GetATMByID_Error(t *testing.T) {
	// Configuración de prueba
	mockClient := new(mockAutomatedTellerMachineClient)
	repo := NewDatastoreATMRepository(mockClient, "test-db", "test-namespace", "ATM")

	// Mock para error en Get
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("datastore error")).Once()

	atm, err := repo.GetATMByID("ATM001")
	assert.Error(t, err)
	assert.Equal(t, models.AutomatedTellerMachine{}, atm)
	assert.Contains(t, err.Error(), "failed to get ATM by identifier")

	mockClient.AssertExpectations(t)
}

func TestAutomatedTellerMachineRepository_Close_Success(t *testing.T) {
	// Configuración de prueba
	mockClient := new(mockAutomatedTellerMachineClient)
	repo := NewDatastoreATMRepository(mockClient, "test-db", "test-namespace", "ATM")

	// Mock para Close exitoso
	mockClient.On("Close").Return(nil).Once()

	err := repo.Close()
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestAutomatedTellerMachineRepository_Close_Error(t *testing.T) {
	// Configuración de prueba
	mockClient := new(mockAutomatedTellerMachineClient)
	repo := NewDatastoreATMRepository(mockClient, "test-db", "test-namespace", "ATM")

	// Mock para error en Close
	mockClient.On("Close").Return(fmt.Errorf("close error")).Once()

	err := repo.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to close datastore client")

	mockClient.AssertExpectations(t)
}

func TestNewAutomatedTellerMachineClientWithRealDatastore_Success(t *testing.T) {
	// Simular credenciales válidas
	fakeCredentials := []byte(`{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "test-private-key-id",
		"private_key": "-----BEGIN PRIVATE KEY-----\nFAKE_PRIVATE_KEY\n-----END PRIVATE KEY-----\n",
		"client_email": "fake-email@test-project.iam.gserviceaccount.com",
		"client_id": "fake-client-id",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/fake-email@test-project.iam.gserviceaccount.com"
	}`)

	// Simular la configuración
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
			JSON:      fakeCredentials, // Usar el JSON simulado
		},
		DataStoreDBName: "test-db",
	}

	// Simular el creador del cliente real
	creator := RealAutomatedTellerMachineCreator{}

	// Simular las opciones del cliente de Datastore
	ctx := context.Background()

	// Esta línea intentará crear un cliente real de Datastore con un contexto y credenciales simuladas
	client, err := creator.NewClientWithDatabase(ctx, cfg.GCPCredentials.ProjectID, cfg.DataStoreDBName, option.WithCredentialsJSON(cfg.GCPCredentials.JSON))

	// Verificar que no hubo errores (puede que necesites ajustar esto dependiendo del entorno de prueba)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
