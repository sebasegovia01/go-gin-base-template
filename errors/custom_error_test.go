package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCustomError(t *testing.T) {
	statusCode := 400
	message := "Bad Request"

	err := NewCustomError(statusCode, message)

	assert.NotNil(t, err)
	assert.Equal(t, statusCode, err.StatusCode)
	assert.Equal(t, message, err.Message)
}

func TestCustomError_Error(t *testing.T) {
	message := "Not Found"
	err := &CustomError{
		StatusCode: 404,
		Message:    message,
	}

	assert.Equal(t, message, err.Error())
}

func TestCustomError_Fields(t *testing.T) {
	statusCode := 500
	message := "Internal Server Error"

	err := &CustomError{
		StatusCode: statusCode,
		Message:    message,
	}

	assert.Equal(t, statusCode, err.StatusCode)
	assert.Equal(t, message, err.Message)
}
