package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPService es un mock de HTTPServiceInterface
type MockHTTPService struct {
	mock.Mock
}

func (m *MockHTTPService) SendRequest(method, url string, params map[string]string, headers map[string]string) ([]byte, error) {
	args := m.Called(method, url, params, headers)
	return args.Get(0).([]byte), args.Error(1)
}

// ... (Las pruebas anteriores permanecen igual)

func TestHTTPService_SendRequest(t *testing.T) {
	t.Run("Successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "value", r.URL.Query().Get("param"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}))
		defer server.Close()

		service := NewHTTPService()

		response, err := service.SendRequest("GET", server.URL, map[string]string{"param": "value"}, map[string]string{"Content-Type": "application/json"})

		assert.NoError(t, err)
		assert.Equal(t, []byte("success"), response)
	})

	t.Run("Failed to create request", func(t *testing.T) {
		service := NewHTTPService()
		invalidURL := "://invalid-url"

		response, err := service.SendRequest("GET", invalidURL, nil, nil)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to create request")
	})

	t.Run("Request failed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond) // Simular un retraso
		}))
		defer server.Close()

		service := &HTTPService{
			client: &http.Client{
				Timeout: 50 * time.Millisecond, // Timeout más corto que el retraso del servidor
			},
		}

		response, err := service.SendRequest("GET", server.URL, nil, nil)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "request failed")
	})

	t.Run("Failed to read response body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1") // Establecer un Content-Length incorrecto
			w.WriteHeader(http.StatusOK)
			// No escribir nada en el cuerpo
		}))
		defer server.Close()

		service := NewHTTPService()

		response, err := service.SendRequest("GET", server.URL, nil, nil)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to read response body")
	})

	t.Run("Received non-200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))
		}))
		defer server.Close()

		service := NewHTTPService()

		response, err := service.SendRequest("GET", server.URL, nil, nil)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "received non-200 response")
	})
}
