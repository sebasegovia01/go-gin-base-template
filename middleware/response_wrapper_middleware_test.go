package middleware_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	"github.com/sebasegovia01/base-template-go-gin/errors"
	"github.com/sebasegovia01/base-template-go-gin/middleware"
	"github.com/sebasegovia01/base-template-go-gin/models"
	"github.com/stretchr/testify/assert"
)

func TestResponseWrapperMiddleware_Success(t *testing.T) {
	// Configurar el contexto de Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.ResponseWrapperMiddleware())

	// Definir un handler de éxito
	router.GET("/success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
		})
	})

	req, _ := http.NewRequest("GET", "/success", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Convertir el tipo de data a map[string]interface{} para la comparación
	expectedData := map[string]interface{}{
		"message": "Success",
	}
	assert.Equal(t, enums.OK, response.Result.Status)
	assert.Equal(t, "Request processed successfully", response.Result.Description)
	assert.Equal(t, expectedData, response.Result.Data) // Comparación correcta
}

func TestResponseWrapperMiddleware_CustomError(t *testing.T) {
	// Configurar el contexto de Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.ResponseWrapperMiddleware())

	// Definir un handler que devuelva un CustomError
	router.GET("/custom_error", func(c *gin.Context) {
		customErr := &errors.CustomError{
			Message:    "Custom error occurred",
			StatusCode: http.StatusBadRequest,
		}
		c.Error(customErr)
	})

	req, _ := http.NewRequest("GET", "/custom_error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, enums.ERROR, response.Result.Status)
	assert.Equal(t, "Custom error occurred", response.Result.SourceError.Description)
}

func TestResponseWrapperMiddleware_StandardError(t *testing.T) {
	// Configurar el contexto de Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.ResponseWrapperMiddleware())

	// Definir un handler que devuelva un error estándar
	router.GET("/standard_error", func(c *gin.Context) {
		c.Error(errors.NewCustomError(500, "Standard error occurred"))
	})

	req, _ := http.NewRequest("GET", "/standard_error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, enums.ERROR, response.Result.Status)
	assert.Equal(t, "Standard error occurred", response.Result.SourceError.Description)
}

func TestResponseWrapperMiddleware_UnmarshalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.ResponseWrapperMiddleware())

	router.GET("/unmarshal_error", func(c *gin.Context) {
		c.Writer.WriteString("Invalid JSON data")
	})

	req, _ := http.NewRequest("GET", "/unmarshal_error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, enums.OK, response.Result.Status)
	assert.Equal(t, "Request processed successfully", response.Result.Description)
	assert.Equal(t, "Invalid JSON data", response.Result.Data)
}

func TestResponseWrapperMiddleware_MarshalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stdout)

	router.Use(middleware.ResponseWrapperMiddleware())

	router.GET("/marshal_error", func(c *gin.Context) {
		c.Set("key", make(chan int))
		// No establecemos el status aquí para que el middleware lo maneje
	})

	req, _ := http.NewRequest("GET", "/marshal_error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "json: unsupported type: chan int", "Los logs deben contener el error de marshaling")

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	assert.Equal(t, enums.ERROR, errorResponse.Result.Status)
	assert.Equal(t, "500", errorResponse.Result.CanonicalError.Code)
	assert.Equal(t, "Internal Server Error", errorResponse.Result.SourceError.Description)
}

func TestResponseWrapperMiddleware_Panic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stdout)

	router.Use(middleware.ResponseWrapperMiddleware())

	router.GET("/panic", func(c *gin.Context) {
		panic("Simulated panic")
	})

	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "Recovered from panic: Simulated panic", "Los logs deben contener el mensaje de pánico recuperado")
}

func TestResponseWrapperMiddleware_UnexpectedError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stdout)

	router.Use(middleware.ResponseWrapperMiddleware())

	router.GET("/unexpected-error", func(c *gin.Context) {
		c.Status(http.StatusBadRequest)
	})

	req, _ := http.NewRequest("GET", "/unexpected-error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	assert.Equal(t, enums.ERROR, errorResponse.Result.Status)
	assert.Equal(t, "400", errorResponse.Result.CanonicalError.Code)
	assert.Equal(t, "An unexpected error occurred", errorResponse.Result.SourceError.Description)

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "Handling unexpected error response", "Los logs deben contener el mensaje de manejo de error inesperado")
}

func TestResponseWrapperMiddleware_WriteError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stdout)

	router.Use(func(c *gin.Context) {
		c.Writer = &errorWriter{ResponseWriter: c.Writer}
		c.Next()
	})
	router.Use(middleware.ResponseWrapperMiddleware())

	router.GET("/write-error", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "This will fail to write"})
	})

	req, _ := http.NewRequest("GET", "/write-error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "Error writing response", "Los logs deben contener el mensaje de error de escritura")
}

type errorWriter struct {
	gin.ResponseWriter
}

func (ew *errorWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("simulated write error")
}

func (ew *errorWriter) WriteHeader(statusCode int) {
	// No hacemos nada aquí para evitar escribir el estado
}

func TestResponseWrapperMiddleware_StandardErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stdout)

	router.Use(middleware.ResponseWrapperMiddleware())

	router.GET("/standard-error", func(c *gin.Context) {
		err := fmt.Errorf("This is a standard error")
		c.Error(err)
	})

	req, _ := http.NewRequest("GET", "/standard-error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	assert.Equal(t, enums.ERROR, errorResponse.Result.Status)
	assert.Equal(t, "500", errorResponse.Result.CanonicalError.Code)
	assert.Equal(t, "This is a standard error", errorResponse.Result.SourceError.Description)

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "Handling standard error", "Los logs deben contener el mensaje de manejo de error estándar")
}

func TestResponseWrapperMiddleware_ErrorWritingErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stdout)

	// Usamos un middleware personalizado para reemplazar el writer con uno que falla al escribir
	router.Use(func(c *gin.Context) {
		c.Writer = &errorWriterWithHeaderWrite{ResponseWriter: c.Writer}
		c.Next()
	})

	router.Use(middleware.ResponseWrapperMiddleware())

	router.GET("/error-writing-error", func(c *gin.Context) {
		// Forzamos un error que el middleware intentará manejar
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("forced error"))
	})

	req, _ := http.NewRequest("GET", "/error-writing-error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "Error writing response", "Los logs deben contener el mensaje de error al escribir la respuesta")
}

type errorWriterWithHeaderWrite struct {
	gin.ResponseWriter
}

func (ew *errorWriterWithHeaderWrite) Write([]byte) (int, error) {
	return 0, fmt.Errorf("simulated write error")
}

func (ew *errorWriterWithHeaderWrite) WriteHeader(statusCode int) {
	// Permitimos que se escriba el estado, pero la escritura del cuerpo fallará
	ew.ResponseWriter.WriteHeader(statusCode)
}
