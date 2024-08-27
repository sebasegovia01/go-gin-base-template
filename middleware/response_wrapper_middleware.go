package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	customErrors "github.com/sebasegovia01/base-template-go-gin/errors"
	"github.com/sebasegovia01/base-template-go-gin/models"
)

func ResponseWrapperMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		buf := &responseBuffer{ResponseWriter: c.Writer}
		c.Writer = buf

		c.Next()

		statusCode := buf.Status()
		originalBody := buf.body.Bytes()

		log.Printf("Original status code: %d", statusCode)
		log.Printf("Original body: %s", string(originalBody))

		var response interface{}
		var data interface{}

		if len(originalBody) > 0 {
			if err := json.Unmarshal(originalBody, &data); err != nil {
				log.Printf("Error unmarshaling body: %v", err)
				data = string(originalBody)
			}
		}

		log.Printf("Unmarshaled data: %+v", data)

		if len(c.Errors) > 0 {
			log.Println("Handling error response")
			err := c.Errors.Last()
			var errorResponse models.ErrorResponse
			if customErr, ok := err.Err.(*customErrors.CustomError); ok {
				log.Println("Handling CustomError")
				statusCode = customErr.StatusCode
				errorResponse = createErrorResponse(statusCode, customErr.Message)

				// Special handling for missing headers
				if strings.Contains(customErr.Message, "Missing required headers") {
					missingHeaders := strings.TrimPrefix(customErr.Message, "Missing required headers: ")
					errorResponse.Result.SourceError.Description = "Missing required headers"
					errorResponse.Result.SourceError.ErrorSourceDetails.MissingHeaders = strings.Split(missingHeaders, ", ")
				}
			} else {
				log.Println("Handling standard error")
				statusCode = http.StatusInternalServerError
				if strings.Contains(err.Error(), "database operation failed") {
					statusCode = http.StatusInternalServerError
				}
				errorResponse = createErrorResponse(statusCode, err.Error())
			}
			response = errorResponse
		} else if statusCode >= 200 && statusCode < 300 {
			log.Println("Handling success response")
			response = models.SuccessResponse{
				Result: struct {
					Status      enums.ResultStatus `json:"status"`
					Description string             `json:"description,omitempty"`
					Data        interface{}        `json:"data,omitempty"`
				}{
					Status:      enums.OK,
					Description: "Request processed successfully",
					Data:        data,
				},
			}
		} else {
			log.Println("Handling unexpected error response")
			response = createErrorResponse(statusCode, "An unexpected error occurred")
		}

		log.Printf("Final response: %+v", response)
		log.Printf("Final status code: %d", statusCode)

		// Reset the original writer
		c.Writer = buf.ResponseWriter

		// Write the response
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.WriteHeader(statusCode)
		responseJSON, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			c.Writer.WriteHeader(http.StatusInternalServerError)
			c.Writer.Write([]byte(`{"error": "Internal Server Error"}`))
			return
		}
		_, err = c.Writer.Write(responseJSON)
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	}
}

func createErrorResponse(statusCode int, message string) models.ErrorResponse {
	return models.ErrorResponse{
		Result: models.ResultError{
			Status: enums.ERROR,
			CanonicalError: &models.CanonicalError{
				Code:        strconv.Itoa(statusCode),
				Type:        enums.TEC,
				Description: http.StatusText(statusCode),
			},
			SourceError: &models.SourceError{
				Code:        strconv.Itoa(statusCode),
				Description: message,
				ErrorSourceDetails: models.ErrorSourceDetails{
					Source: "API",
				},
			},
		},
	}
}

type responseBuffer struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
}

func (r *responseBuffer) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *responseBuffer) WriteString(s string) (int, error) {
	return r.body.WriteString(s)
}

func (r *responseBuffer) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseBuffer) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}
