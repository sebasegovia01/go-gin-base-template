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

// Mock para PresentialChannelInterface
type mockPresentialChannelClient struct {
	mock.Mock
}

func (m *mockPresentialChannelClient) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	args := m.Called(ctx, key, dst)
	return args.Error(0)
}

func (m *mockPresentialChannelClient) GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error) {
	args := m.Called(ctx, q, dst)
	return args.Get(0).([]*datastore.Key), args.Error(1)
}

func (m *mockPresentialChannelClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock para PresentialChannelCreator
type mockPresentialChannelCreator struct {
	mock.Mock
}

func (m *mockPresentialChannelCreator) NewClientWithDatabase(ctx context.Context, projectID string, databaseID string, opts ...option.ClientOption) (PresentialChannelInterface, error) {
	args := m.Called(ctx, projectID, databaseID, opts)
	return args.Get(0).(PresentialChannelInterface), args.Error(1)
}

// Prueba para la creación exitosa de un cliente de Presential Channel
func TestNewPresentialChannelClient_Success(t *testing.T) {
	mockCreator := new(mockPresentialChannelCreator)
	mockClient := new(mockPresentialChannelClient)

	cfg := &config.Config{
		GCPCredentials:  &google.Credentials{},
		DataStoreDBName: "test-db",
	}

	mockCreator.On("NewClientWithDatabase", mock.Anything, "test-project", "test-db", mock.Anything).Return(mockClient, nil)

	cfg.GCPCredentials.ProjectID = "test-project"

	client, err := NewPresentialChannelClient(cfg, mockCreator)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	mockCreator.AssertExpectations(t)
}

// Prueba para error en la creación de un cliente de Presential Channel
func TestNewPresentialChannelClient_Error(t *testing.T) {
	mockCreator := new(mockPresentialChannelCreator)
	mockClient := new(mockPresentialChannelClient)

	cfg := &config.Config{
		GCPCredentials:  &google.Credentials{},
		DataStoreDBName: "test-db",
	}

	mockCreator.On("NewClientWithDatabase", mock.Anything, "test-project", "test-db", mock.Anything).Return(mockClient, fmt.Errorf("client error"))

	cfg.GCPCredentials.ProjectID = "test-project"

	client, err := NewPresentialChannelClient(cfg, mockCreator)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to create Datastore client")

	mockCreator.AssertExpectations(t)
}

// Prueba para obtener todos los canales presenciales exitosamente
func TestPresentialChannelRepository_GetAllPresentialChannels_Success(t *testing.T) {
	mockClient := new(mockPresentialChannelClient)
	repo := NewDatastorePresentialChannelRepository(mockClient, "test-db", "test-namespace", "PresentialChannel")

	expectedChannels := []models.PresentialChannel{
		{PresentialChannelIdentifier: "CHANNEL001", StreetName: "First St"},
		{PresentialChannelIdentifier: "CHANNEL002", StreetName: "Second St"},
	}

	mockClient.On("GetAll", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dst := args.Get(2).(*[]models.PresentialChannel)
		*dst = expectedChannels
	}).Return([]*datastore.Key{}, nil).Once()

	channels, err := repo.GetAllPresentialChannels()
	assert.NoError(t, err)
	assert.Equal(t, expectedChannels, channels)

	mockClient.AssertExpectations(t)
}

// Prueba para error al obtener todos los canales presenciales
func TestPresentialChannelRepository_GetAllPresentialChannels_Error(t *testing.T) {
	mockClient := new(mockPresentialChannelClient)
	repo := NewDatastorePresentialChannelRepository(mockClient, "test-db", "test-namespace", "PresentialChannel")

	mockClient.On("GetAll", mock.Anything, mock.Anything, mock.Anything).Return([]*datastore.Key{}, fmt.Errorf("datastore error")).Once()

	channels, err := repo.GetAllPresentialChannels()
	assert.Error(t, err)
	assert.Nil(t, channels)
	assert.Contains(t, err.Error(), "failed to get all Presential Channels")

	mockClient.AssertExpectations(t)
}

// Prueba para obtener un canal presencial por ID exitosamente
func TestPresentialChannelRepository_GetPresentialChannelByID_Success(t *testing.T) {
	mockClient := new(mockPresentialChannelClient)
	repo := NewDatastorePresentialChannelRepository(mockClient, "test-db", "test-namespace", "PresentialChannel")

	expectedChannel := models.PresentialChannel{PresentialChannelIdentifier: "CHANNEL001", StreetName: "First St"}

	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		dst := args.Get(2).(*models.PresentialChannel)
		*dst = expectedChannel
	}).Return(nil).Once()

	channel, err := repo.GetPresentialChannelByID("CHANNEL001")
	assert.NoError(t, err)
	assert.Equal(t, expectedChannel, channel)

	mockClient.AssertExpectations(t)
}

// Prueba para error al obtener un canal presencial por ID
func TestPresentialChannelRepository_GetPresentialChannelByID_Error(t *testing.T) {
	mockClient := new(mockPresentialChannelClient)
	repo := NewDatastorePresentialChannelRepository(mockClient, "test-db", "test-namespace", "PresentialChannel")

	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("datastore error")).Once()

	channel, err := repo.GetPresentialChannelByID("CHANNEL001")
	assert.Error(t, err)
	assert.Equal(t, models.PresentialChannel{}, channel)
	assert.Contains(t, err.Error(), "failed to get Presential Channel by identifier")

	mockClient.AssertExpectations(t)
}

// Prueba para cerrar el cliente de Datastore exitosamente
func TestPresentialChannelRepository_Close_Success(t *testing.T) {
	mockClient := new(mockPresentialChannelClient)
	repo := NewDatastorePresentialChannelRepository(mockClient, "test-db", "test-namespace", "PresentialChannel")

	mockClient.On("Close").Return(nil).Once()

	err := repo.Close()
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

// Prueba para error al cerrar el cliente de Datastore
func TestPresentialChannelRepository_Close_Error(t *testing.T) {
	mockClient := new(mockPresentialChannelClient)
	repo := NewDatastorePresentialChannelRepository(mockClient, "test-db", "test-namespace", "PresentialChannel")

	mockClient.On("Close").Return(fmt.Errorf("close error")).Once()

	err := repo.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to close datastore client")

	mockClient.AssertExpectations(t)
}

func TestNewPresentialChannelClientWithRealDatastore_Success(t *testing.T) {
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
	creator := RealPresentialChannelCreator{}

	// Simular las opciones del cliente de Datastore
	ctx := context.Background()

	// Esta línea intentará crear un cliente real de Datastore con un contexto y credenciales simuladas
	client, err := creator.NewClientWithDatabase(ctx, cfg.GCPCredentials.ProjectID, cfg.DataStoreDBName, option.WithCredentialsJSON(cfg.GCPCredentials.JSON))

	// Verificar que no hubo errores
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
