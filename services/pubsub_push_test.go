package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2/google"
)

type MockPubSubService struct {
	mock.Mock
}

func (m *MockPubSubService) ExtractStorageEvent(body io.Reader) (*StorageEvent, error) {
	args := m.Called(body)
	return args.Get(0).(*StorageEvent), args.Error(1)
}

func TestNewPubSubService_Success(t *testing.T) {
	// Crear una configuración simulada con credenciales válidas
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
		},
	}

	// Ejecutar la función
	service, err := NewPubSubService(cfg)

	// Aserciones
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "test-project", cfg.GCPCredentials.ProjectID)
}

func TestExtractStorageEvent_Success(t *testing.T) {
	// Simular el cuerpo del mensaje PubSub válido
	pubSubMessage := PubSubMessage{
		Message: struct {
			Attributes   map[string]string `json:"attributes,omitempty"`
			Data         string            `json:"data"`
			ID           string            `json:"messageId"`
			MessageID    string            `json:"message_id,omitempty"`
			OrderingKey  *string           `json:"orderingKey,omitempty"`
			PublishTime  string            `json:"publishTime"`
			PublishTime2 string            `json:"publish_time,omitempty"`
		}{
			Attributes: map[string]string{
				"eventType": "OBJECT_FINALIZE", // Asegúrate de que el eventType sea válido
			},
			Data: base64.StdEncoding.EncodeToString([]byte(`{"bucket": "my-bucket", "name": "my-object"}`)),
			ID:   "test-id",
		},
		DeliveryAttempt: new(int),
	}

	messageBytes, _ := json.Marshal(pubSubMessage)
	body := io.NopCloser(bytes.NewReader(messageBytes))

	// Crear el servicio con credenciales válidas
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
		},
	}
	service, _ := NewPubSubService(cfg)

	// Llamar al método
	result, err := service.ExtractStorageEvent(body)

	// Aserciones
	assert.NoError(t, err)
	assert.Equal(t, "my-bucket", result.Bucket)
	assert.Equal(t, "my-object", result.Name)
}

func TestExtractStorageEvent_UnsupportedEventType(t *testing.T) {
	// Simular el cuerpo del mensaje PubSub con un eventType no soportado (por ejemplo, OBJECT_DELETE)
	pubSubMessage := PubSubMessage{
		Message: struct {
			Attributes   map[string]string `json:"attributes,omitempty"`
			Data         string            `json:"data"`
			ID           string            `json:"messageId"`
			MessageID    string            `json:"message_id,omitempty"`
			OrderingKey  *string           `json:"orderingKey,omitempty"`
			PublishTime  string            `json:"publishTime"`
			PublishTime2 string            `json:"publish_time,omitempty"`
		}{
			Attributes: map[string]string{
				"eventType": "OBJECT_DELETE", // Tipo de evento no soportado
			},
			Data: base64.StdEncoding.EncodeToString([]byte(`{"bucket": "my-bucket", "name": "my-object"}`)),
			ID:   "test-id",
		},
		DeliveryAttempt: new(int),
	}

	messageBytes, _ := json.Marshal(pubSubMessage)
	body := io.NopCloser(bytes.NewReader(messageBytes))

	// Crear el servicio con credenciales válidas
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
		},
	}
	service, _ := NewPubSubService(cfg)

	// Llamar al método
	result, err := service.ExtractStorageEvent(body)

	// Aserciones
	assert.NoError(t, err) // No debe haber error
	assert.Nil(t, result)  // El resultado debe ser nil
}

// Simular un lector que falla en la lectura del cuerpo
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

func TestExtractStorageEvent_ReadBodyError(t *testing.T) {
	// Simular un error en la lectura del cuerpo
	body := io.NopCloser(&errorReader{})

	// Crear el servicio con credenciales válidas
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
		},
	}
	service, _ := NewPubSubService(cfg)

	// Llamar al método y verificar el error
	result, err := service.ExtractStorageEvent(body)

	// Aserciones
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error reading request body")
}

func TestExtractStorageEvent_UnmarshalMessageError(t *testing.T) {
	// Simular un cuerpo de mensaje PubSub inválido
	invalidMessage := []byte(`{"message": "invalid-data"}`)
	body := io.NopCloser(bytes.NewReader(invalidMessage))

	// Crear el servicio con credenciales válidas
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
		},
	}
	service, _ := NewPubSubService(cfg)

	// Llamar al método y verificar el error
	result, err := service.ExtractStorageEvent(body)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshalling message")
}

func TestExtractStorageEvent_Base64DecodeError(t *testing.T) {
	// Simular un mensaje PubSub con datos base64 inválidos
	pubSubMessage := PubSubMessage{
		Message: struct {
			Attributes   map[string]string `json:"attributes,omitempty"`
			Data         string            `json:"data"`
			ID           string            `json:"messageId"`
			MessageID    string            `json:"message_id,omitempty"`
			OrderingKey  *string           `json:"orderingKey,omitempty"`
			PublishTime  string            `json:"publishTime"`
			PublishTime2 string            `json:"publish_time,omitempty"`
		}{
			Attributes: map[string]string{
				"eventType": "OBJECT_FINALIZE", // Asegúrate de que el eventType sea válido
			},
			Data: "invalid-base64-data",
			ID:   "test-id",
		},
	}

	messageBytes, _ := json.Marshal(pubSubMessage)
	body := io.NopCloser(bytes.NewReader(messageBytes))

	// Crear el servicio con credenciales válidas
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
		},
	}
	service, _ := NewPubSubService(cfg)

	// Llamar al método y verificar el error
	result, err := service.ExtractStorageEvent(body)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error decoding message data")
}

func TestExtractStorageEvent_WithAttributes(t *testing.T) {
	// Simular el cuerpo del mensaje PubSub con atributos
	pubSubMessage := PubSubMessage{
		Message: struct {
			Attributes   map[string]string `json:"attributes,omitempty"`
			Data         string            `json:"data"`
			ID           string            `json:"messageId"`
			MessageID    string            `json:"message_id,omitempty"`
			OrderingKey  *string           `json:"orderingKey,omitempty"`
			PublishTime  string            `json:"publishTime"`
			PublishTime2 string            `json:"publish_time,omitempty"`
		}{
			Attributes: map[string]string{
				"eventType": "OBJECT_FINALIZE",
				"eventTime": "2024-09-30T18:20:43.506911Z",
			},
			Data: base64.StdEncoding.EncodeToString([]byte(`{"bucket": "my-bucket", "name": "my-object"}`)),
			ID:   "test-id",
		},
	}

	messageBytes, _ := json.Marshal(pubSubMessage)
	body := io.NopCloser(bytes.NewReader(messageBytes))

	// Crear el servicio con credenciales válidas
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
		},
	}
	service, _ := NewPubSubService(cfg)

	// Llamar al método
	result, err := service.ExtractStorageEvent(body)

	// Aserciones
	assert.NoError(t, err)
	assert.Equal(t, "my-bucket", result.Bucket)
	assert.Equal(t, "my-object", result.Name)
}

func TestExtractStorageEvent_EmptyBucketError(t *testing.T) {
	// Simular un mensaje PubSub con un StorageEvent que tiene el bucket vacío
	pubSubMessage := PubSubMessage{
		Message: struct {
			Attributes   map[string]string `json:"attributes,omitempty"`
			Data         string            `json:"data"`
			ID           string            `json:"messageId"`
			MessageID    string            `json:"message_id,omitempty"`
			OrderingKey  *string           `json:"orderingKey,omitempty"`
			PublishTime  string            `json:"publishTime"`
			PublishTime2 string            `json:"publish_time,omitempty"`
		}{
			Attributes: map[string]string{
				"eventType": "OBJECT_FINALIZE",
				"eventTime": "2024-09-30T18:20:43.506911Z",
			},
			Data: base64.StdEncoding.EncodeToString([]byte(`{"bucket": "", "name": "my-object"}`)),
			ID:   "test-id",
		},
	}

	messageBytes, _ := json.Marshal(pubSubMessage)
	body := io.NopCloser(bytes.NewReader(messageBytes))

	// Crear el servicio con credenciales válidas
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
		},
	}
	service, _ := NewPubSubService(cfg)

	// Llamar al método y verificar el error
	result, err := service.ExtractStorageEvent(body)

	// Aserciones
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "bucket name not found in message")
}

func TestExtractStorageEvent_EmptyNameError(t *testing.T) {
	// Simular un mensaje PubSub con un StorageEvent que tiene el nombre vacío
	pubSubMessage := PubSubMessage{
		Message: struct {
			Attributes   map[string]string `json:"attributes,omitempty"`
			Data         string            `json:"data"`
			ID           string            `json:"messageId"`
			MessageID    string            `json:"message_id,omitempty"`
			OrderingKey  *string           `json:"orderingKey,omitempty"`
			PublishTime  string            `json:"publishTime"`
			PublishTime2 string            `json:"publish_time,omitempty"`
		}{
			Attributes: map[string]string{
				"eventType": "OBJECT_FINALIZE",
				"eventTime": "2024-09-30T18:20:43.506911Z",
			},
			Data: base64.StdEncoding.EncodeToString([]byte(`{"bucket": "my-bucket", "name": ""}`)),
			ID:   "test-id",
		},
	}

	messageBytes, _ := json.Marshal(pubSubMessage)
	body := io.NopCloser(bytes.NewReader(messageBytes))

	// Crear el servicio con credenciales válidas
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
		},
	}
	service, _ := NewPubSubService(cfg)

	// Llamar al método y verificar el error
	result, err := service.ExtractStorageEvent(body)

	// Aserciones
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "object name not found in message")
}

func TestExtractStorageEvent_UnmarshalStorageEventError(t *testing.T) {
	// Simular un mensaje PubSub con datos válidos en base64, pero con JSON inválido
	pubSubMessage := PubSubMessage{
		Message: struct {
			Attributes   map[string]string `json:"attributes,omitempty"`
			Data         string            `json:"data"`
			ID           string            `json:"messageId"`
			MessageID    string            `json:"message_id,omitempty"`
			OrderingKey  *string           `json:"orderingKey,omitempty"`
			PublishTime  string            `json:"publishTime"`
			PublishTime2 string            `json:"publish_time,omitempty"`
		}{
			Attributes: map[string]string{
				"eventType": "OBJECT_FINALIZE",
				"eventTime": "2024-09-30T18:20:43.506911Z",
			},
			Data: base64.StdEncoding.EncodeToString([]byte(`{"invalid-json"`)), // JSON inválido
			ID:   "test-id",
		},
	}

	messageBytes, _ := json.Marshal(pubSubMessage)
	body := io.NopCloser(bytes.NewReader(messageBytes))

	// Crear el servicio con credenciales válidas
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
		},
	}
	service, _ := NewPubSubService(cfg)

	// Llamar al método y verificar el error
	result, err := service.ExtractStorageEvent(body)

	// Aserciones
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshalling storage event data")
}

func TestExtractStorageEvent_QueryUnescapeError(t *testing.T) {
	// Simular un mensaje PubSub con datos válidos en base64, pero con nombre del objeto no decodificable
	pubSubMessage := PubSubMessage{
		Message: struct {
			Attributes   map[string]string `json:"attributes,omitempty"`
			Data         string            `json:"data"`
			ID           string            `json:"messageId"`
			MessageID    string            `json:"message_id,omitempty"`
			OrderingKey  *string           `json:"orderingKey,omitempty"`
			PublishTime  string            `json:"publishTime"`
			PublishTime2 string            `json:"publish_time,omitempty"`
		}{
			Attributes: map[string]string{
				"eventType": "OBJECT_FINALIZE",
				"eventTime": "2024-09-30T18:20:43.506911Z",
			},
			Data: base64.StdEncoding.EncodeToString([]byte(`{"bucket": "my-bucket", "name": "%invalid%name"}`)),
			ID:   "test-id",
		},
	}

	messageBytes, _ := json.Marshal(pubSubMessage)
	body := io.NopCloser(bytes.NewReader(messageBytes))

	// Crear el servicio con credenciales válidas
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
		},
	}
	service, _ := NewPubSubService(cfg)

	// Llamar al método y verificar el error
	result, err := service.ExtractStorageEvent(body)

	// Aserciones
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error decoding object name")
}
