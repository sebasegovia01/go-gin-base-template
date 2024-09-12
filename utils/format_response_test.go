package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformResponse(t *testing.T) {
	t.Run("Successful transformation", func(t *testing.T) {
		input := []byte(`{
			"Result": {
				"data": {
					"data": {
						"id": "12345",
						"name": "Test Item"
					}
				}
			}
		}`)
		expected := []byte(`{"id":"12345","name":"Test Item"}`)

		result, err := TransformResponse(input)
		assert.NoError(t, err)
		assert.JSONEq(t, string(expected), string(result))
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		input := []byte(`{invalid json}`)

		result, err := TransformResponse(input)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "error unmarshaling original response")
	})

	t.Run("Missing Result field", func(t *testing.T) {
		input := []byte(`{"SomeOtherField": {}}`)

		result, err := TransformResponse(input)
		assert.NoError(t, err)
		assert.Equal(t, "null", string(result))
	})

	t.Run("Missing data field in Result", func(t *testing.T) {
		input := []byte(`{"Result": {"someOtherField": {}}}`)

		result, err := TransformResponse(input)
		assert.NoError(t, err)
		assert.Equal(t, "null", string(result))
	})

	t.Run("Missing nested data field", func(t *testing.T) {
		input := []byte(`{"Result": {"data": {"someOtherField": {}}}}`)

		result, err := TransformResponse(input)
		assert.NoError(t, err)
		assert.Equal(t, "null", string(result))
	})

	t.Run("Empty nested data field", func(t *testing.T) {
		input := []byte(`{"Result": {"data": {"data": {}}}}`)

		result, err := TransformResponse(input)
		assert.NoError(t, err)
		assert.Equal(t, "{}", string(result))
	})

	t.Run("Non-object nested data field", func(t *testing.T) {
		input := []byte(`{"Result": {"data": {"data": "string value"}}}`)

		result, err := TransformResponse(input)
		assert.NoError(t, err)
		assert.Equal(t, `"string value"`, string(result))
	})
}
