package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/models"
	"github.com/sebasegovia01/base-template-go-gin/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks para los servicios
type MockPubSubService struct {
	mock.Mock
}

func (m *MockPubSubService) ExtractStorageEvent(body io.Reader) (*services.StorageEvent, error) {
	args := m.Called(body)
	if args.Get(0) != nil {
		return args.Get(0).(*services.StorageEvent), args.Error(1)
	}
	return nil, args.Error(1)
}

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

func TestHandlePushMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		setupMocks         func(*MockPubSubService, *MockStorageService, *MockPubSubPublishService)
		expectedStatusCode int
		expectedResponse   gin.H
	}{
		{
			name: "Successful processing of electronic channels",
			setupMocks: func(mps *MockPubSubService, mss *MockStorageService, mpps *MockPubSubPublishService) {
				mps.On("ExtractStorageEvent", mock.Anything).Return(&services.StorageEvent{Name: "channels.json"}, nil)
				mss.On("ProcessFile", "channels.json").Return([]*map[string]interface{}{
					{
						"payload": map[string]interface{}{
							"BOPERS_WEB_CHANNEL": map[string]interface{}{
								"WEB_CHANNEL_TYPE": "Public",
							},
						},
					},
				}, nil)
				mpps.On("PublishMessage", mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: gin.H{
				"status":     "Electronic channel data processed and published successfully",
				"data_count": float64(1),
			},
		},
		{
			name: "Error extracting storage event",
			setupMocks: func(mps *MockPubSubService, mss *MockStorageService, mpps *MockPubSubPublishService) {
				mps.On("ExtractStorageEvent", mock.Anything).Return((*services.StorageEvent)(nil), errors.New("extraction error"))
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: gin.H{
				"error": "extraction error",
			},
		},
		{
			name: "Error processing file",
			setupMocks: func(mps *MockPubSubService, mss *MockStorageService, mpps *MockPubSubPublishService) {
				mps.On("ExtractStorageEvent", mock.Anything).Return(&services.StorageEvent{Name: "channels.json"}, nil)
				mss.On("ProcessFile", "channels.json").Return([]*map[string]interface{}{}, errors.New("processing error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": "Error processing file: processing error",
			},
		},
		{
			name: "Error transforming channel data",
			setupMocks: func(mps *MockPubSubService, mss *MockStorageService, mpps *MockPubSubPublishService) {
				mps.On("ExtractStorageEvent", mock.Anything).Return(&services.StorageEvent{Name: "channels.json"}, nil)
				mss.On("ProcessFile", "channels.json").Return([]*map[string]interface{}{
					{
						"payload": map[string]interface{}{
							"BOPERS_WEB_CHANNEL": map[string]interface{}{
								"WEB_CHANNEL_TYPE": "Public",
							},
						},
					},
				}, nil)

				// Simulamos un error en la transformación de datos
				originalTransformFunc := transformElectronicChannelDataFunc
				transformElectronicChannelDataFunc = func(data *map[string]interface{}) (*models.ElectronicChannels, error) {
					return nil, errors.New("transformation error")
				}
				t.Cleanup(func() { transformElectronicChannelDataFunc = originalTransformFunc })
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": "Error transforming channel data: transformation error",
			},
		},
		{
			name: "Error publishing message",
			setupMocks: func(mps *MockPubSubService, mss *MockStorageService, mpps *MockPubSubPublishService) {
				mps.On("ExtractStorageEvent", mock.Anything).Return(&services.StorageEvent{Name: "channels.json"}, nil)
				mss.On("ProcessFile", "channels.json").Return([]*map[string]interface{}{
					{
						"payload": map[string]interface{}{
							"BOPERS_WEB_CHANNEL": map[string]interface{}{
								"WEB_CHANNEL_TYPE": "Public",
							},
						},
					},
				}, nil)
				mpps.On("PublishMessage", mock.Anything).Return(errors.New("publishing error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": "Error publishing message: publishing error",
			},
		},
		{
			name: "Error marshaling channel data",
			setupMocks: func(mps *MockPubSubService, mss *MockStorageService, mpps *MockPubSubPublishService) {
				mps.On("ExtractStorageEvent", mock.Anything).Return(&services.StorageEvent{Name: "channels.json"}, nil)
				mss.On("ProcessFile", "channels.json").Return([]*map[string]interface{}{
					{
						"payload": map[string]interface{}{
							"BOPERS_WEB_CHANNEL": map[string]interface{}{
								"WEB_CHANNEL_TYPE": make(chan int), // Esto causará un error al serializar
							},
						},
					},
				}, nil)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": "Error marshaling channel data: json: unsupported type: chan int",
			},
		},
		{
			name: "Error publishing message",
			setupMocks: func(mps *MockPubSubService, mss *MockStorageService, mpps *MockPubSubPublishService) {
				mps.On("ExtractStorageEvent", mock.Anything).Return(&services.StorageEvent{Name: "channels.json"}, nil)
				mss.On("ProcessFile", "channels.json").Return([]*map[string]interface{}{
					{
						"payload": map[string]interface{}{
							"BOPERS_WEB_CHANNEL": map[string]interface{}{
								"WEB_CHANNEL_TYPE": "Public",
							},
						},
					},
				}, nil)
				mpps.On("PublishMessage", mock.Anything).Return(errors.New("publishing error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: gin.H{
				"error": "Error publishing message: publishing error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPubSubService := new(MockPubSubService)
			mockStorageService := new(MockStorageService)
			mockPubSubPublishService := new(MockPubSubPublishService)

			tt.setupMocks(mockPubSubService, mockStorageService, mockPubSubPublishService)

			controller := NewDataElectronicChannelsController(mockPubSubService, mockStorageService, mockPubSubPublishService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{}"))

			controller.HandlePushMessage(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response gin.H
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
		})
	}
}

func TestHandlePushMessage_ErrorPublishingMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPubSubService := new(MockPubSubService)
	mockStorageService := new(MockStorageService)
	mockPubSubPublishService := new(MockPubSubPublishService)

	mockPubSubService.On("ExtractStorageEvent", mock.Anything).Return(&services.StorageEvent{Name: "channels.json"}, nil)
	mockStorageService.On("ProcessFile", "channels.json").Return([]*map[string]interface{}{
		{
			"payload": map[string]interface{}{
				"BOPERS_WEB_CHANNEL": map[string]interface{}{
					"WEB_CHANNEL_TYPE": "Public",
				},
			},
		},
	}, nil)
	mockPubSubPublishService.On("PublishMessage", mock.Anything).Return(errors.New("publishing error"))

	controller := NewDataElectronicChannelsController(mockPubSubService, mockStorageService, mockPubSubPublishService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{}"))

	controller.HandlePushMessage(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Error publishing message: publishing error", response["error"])

	mockPubSubService.AssertExpectations(t)
	mockStorageService.AssertExpectations(t)
	mockPubSubPublishService.AssertExpectations(t)
}

func TestHandlePushMessage_UnsupportedEventType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mocks
	mockPubSubService := new(MockPubSubService)
	mockStorageService := new(MockStorageService)
	mockPubSubPublishService := new(MockPubSubPublishService)

	// Simulamos que el evento fue ignorado retornando nil sin error
	mockPubSubService.On("ExtractStorageEvent", mock.Anything).Return(nil, nil)

	controller := NewDataElectronicChannelsController(mockPubSubService, mockStorageService, mockPubSubPublishService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{}"))

	// Ejecutar la función del controlador
	controller.HandlePushMessage(c)

	// Verificar el código de estado HTTP
	assert.Equal(t, http.StatusOK, w.Code)

	// Verificar el contenido de la respuesta
	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Event ignored", response["status"])
	assert.Equal(t, "Event type is not supported, no action taken", response["description"])

	// Asegurarse de que se llamaron las expectativas de los mocks
	mockPubSubService.AssertExpectations(t)
	mockStorageService.AssertNotCalled(t, "ProcessFile", mock.Anything)
	mockPubSubPublishService.AssertNotCalled(t, "PublishMessage", mock.Anything)
}

func TestHandlePushMessage_ErrorCustomMarshalingData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPubSubService := new(MockPubSubService)
	mockStorageService := new(MockStorageService)
	mockPubSubPublishService := new(MockPubSubPublishService)

	mockPubSubService.On("ExtractStorageEvent", mock.Anything).Return(&services.StorageEvent{Name: "channels.json"}, nil)
	mockStorageService.On("ProcessFile", "channels.json").Return([]*map[string]interface{}{
		{
			"payload": map[string]interface{}{
				"BOPERS_WEB_CHANNEL": map[string]interface{}{
					"WEB_CHANNEL_TYPE": "Public",
				},
			},
		},
	}, nil)

	// Guardamos la función original para restaurarla después
	originalCustomMarshalJSONFunc := customMarshalJSONFunc

	// Reemplazamos customMarshalJSONFunc con una versión que siempre falla
	customMarshalJSONFunc = func(v interface{}) ([]byte, error) {
		return nil, fmt.Errorf("forced custom marshal error")
	}

	// Restauramos la función original al final de la prueba
	defer func() { customMarshalJSONFunc = originalCustomMarshalJSONFunc }()

	controller := NewDataElectronicChannelsController(mockPubSubService, mockStorageService, mockPubSubPublishService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{}"))

	controller.HandlePushMessage(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "Error marshaling channel data")

	mockPubSubService.AssertExpectations(t)
	mockStorageService.AssertExpectations(t)
}
