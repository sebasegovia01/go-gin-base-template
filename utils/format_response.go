package utils

import (
	"encoding/json"
	"fmt"
)

// TransformResponse extrae el campo "data" y elimina la estructura anidada redundante por estandar bancario.
func TransformResponse(originalResponse []byte) ([]byte, error) {
	var original map[string]interface{}
	err := json.Unmarshal(originalResponse, &original)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling original response: %w", err)
	}

	result, _ := original["Result"].(map[string]interface{})

	dataContainer, _ := result["data"].(map[string]interface{})

	return json.Marshal(dataContainer["data"])
}
