package middleware

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	customErrors "github.com/sebasegovia01/base-template-go-gin/errors"
	"github.com/sebasegovia01/base-template-go-gin/models"
)

// StatusCoder is an interface for errors that carry HTTP status codes
type StatusCoder interface {
	StatusCode() int
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			var statusCode int
			var errorMsg string

			if customErr, ok := err.Err.(*customErrors.CustomError); ok {
				// Handle custom error
				statusCode = customErr.StatusCode
				errorMsg = customErr.Message
			} else {
				// Try to get status code from the error
				var statusCoder StatusCoder
				if errors.As(err.Err, &statusCoder) {
					statusCode = statusCoder.StatusCode()
				} else {
					statusCode = http.StatusInternalServerError
				}
				errorMsg = err.Error()
			}

			errorResponse := models.ErrorResponse{
				Result: models.ResultError{
					Status: enums.ERROR,
					CanonicalError: &models.CanonicalError{
						Code:        strconv.Itoa(statusCode),
						Type:        enums.TEC,
						Description: http.StatusText(statusCode),
					},
					SourceError: &models.SourceError{
						Code:        strconv.Itoa(statusCode),
						Description: errorMsg,
						ErrorSourceDetails: models.ErrorSourceDetails{
							Source: "API",
						},
					},
				},
			}

			c.JSON(statusCode, errorResponse)
			c.Abort()
		}
	}
}
