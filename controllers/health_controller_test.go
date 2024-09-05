package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// Configurar Gin en modo test
	gin.SetMode(gin.TestMode)

	// Crear un nuevo controlador
	controller := NewHealthController()

	// Crear un contexto de prueba y un recorder para capturar la respuesta
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Ejecutar la función del controlador
	controller.HealthCheck(c)

	// Verificar el código de estado HTTP
	assert.Equal(t, http.StatusOK, w.Code)

	// Verificar el cuerpo de la respuesta
	expectedResponse := gin.H{
		"status":  "UP",
		"message": "API is healthy",
	}
	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}
