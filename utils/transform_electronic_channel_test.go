package utils_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/sebasegovia01/base-template-go-gin/utils"
	"github.com/stretchr/testify/assert"
)

func TestTransformElectronicChannelData(t *testing.T) {
	// Caso 1: Estructura válida con todos los canales
	data := map[string]interface{}{
		"payload": map[string]interface{}{
			"BOPERS_WEB_CHANNEL": map[string]interface{}{
				"WEB_CHANNEL_TYPE":       "Public",
				"WEB_URL_ADDRESS":        "https://example.com",
				"WEB_AVAILABLE_SERVICES": "Banking",
				"WEB_ATTENTION_HOURS":    "24/7",
				"WEB_PLATFORM_TYPE":      "Mobile",
			},
			"BOPERS_EMAIL_CHANNEL": map[string]interface{}{
				"EMAIL_AVAILABLE_SERVICES": "Customer Support",
				"EMAIL_ADDRESS":            "support@example.com",
				"EMAIL_ATTENTION_HOURS":    "9am-5pm",
			},
			"BOPERS_SOCIAL_MEDIA_CHANNEL": map[string]interface{}{
				"SOCIAL_MEDIA_AVAILABLE_SERVICES": "Messaging",
				"SOCIAL_MEDIA_ACCOUNT":            "@example",
				"SOCIAL_MEDIA_ATTENTION_HOURS":    "24/7",
			},
		},
	}

	electronicChannels, err := utils.TransformElectronicChannelData(&data)
	assert.NoError(t, err)
	assert.NotNil(t, electronicChannels)

	// Verificación de web channel
	assert.Equal(t, "Public", electronicChannels.WebChannel.WebChannelType)
	assert.Equal(t, "https://example.com", electronicChannels.WebChannel.WebURLAddress)
	assert.Equal(t, "Banking", electronicChannels.WebChannel.WebAvailableServices)
	assert.Equal(t, "24/7", electronicChannels.WebChannel.WebAttentionHours)
	assert.Equal(t, "Mobile", electronicChannels.WebChannel.WebPlatformType)

	// Verificación de email channel
	assert.Equal(t, "Customer Support", electronicChannels.EmailChannel.EmailAvailableServices)
	assert.Equal(t, "support@example.com", electronicChannels.EmailChannel.EmailAddress)
	assert.Equal(t, "9am-5pm", electronicChannels.EmailChannel.EmailAttentionHours)

	// Verificación de social media channel
	assert.Equal(t, "Messaging", electronicChannels.SocialMediaChannel.SocialMediaAvailableServices)
	assert.Equal(t, "@example", electronicChannels.SocialMediaChannel.SocialMediaAccount)
	assert.Equal(t, "24/7", electronicChannels.SocialMediaChannel.SocialMediaAttentionHours)

	// Caso 2: Payload vacío o con estructura inválida
	invalidData := map[string]interface{}{
		"payload": "invalid structure",
	}
	_, err = utils.TransformElectronicChannelData(&invalidData)
	assert.Error(t, err)
	assert.Equal(t, "invalid payload structure", err.Error())

	// Caso 3: Estructura válida pero con algunos canales vacíos
	partialData := map[string]interface{}{
		"payload": map[string]interface{}{
			"BOPERS_WEB_CHANNEL": map[string]interface{}{
				"WEB_CHANNEL_TYPE": "Public",
			},
			"BOPERS_EMAIL_CHANNEL": map[string]interface{}{
				"EMAIL_ADDRESS": "support@example.com",
			},
		},
	}
	electronicChannels, err = utils.TransformElectronicChannelData(&partialData)
	assert.NoError(t, err)
	assert.NotNil(t, electronicChannels)

	// Verificación de web channel parcial
	assert.Equal(t, "Public", electronicChannels.WebChannel.WebChannelType)
	assert.Empty(t, electronicChannels.WebChannel.WebURLAddress) // Campo vacío

	// Verificación de email channel parcial
	assert.Equal(t, "support@example.com", electronicChannels.EmailChannel.EmailAddress)
	assert.Empty(t, electronicChannels.EmailChannel.EmailAvailableServices) // Campo vacío

	// Verificación de social media channel vacío
	assert.Nil(t, electronicChannels.SocialMediaChannel)
}

func TestCustomMarshalJSON(t *testing.T) {
	// Caso 1: Estructura con datos
	type TestStruct struct {
		Name  string `json:"name"`
		Empty string `json:"empty,omitempty"`
	}
	obj := &TestStruct{
		Name: "John",
	}
	jsonData, err := utils.CustomMarshalJSON(obj)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"John"}`, string(jsonData))

	// Caso 2: Estructura vacía
	obj = &TestStruct{}
	jsonData, err = utils.CustomMarshalJSON(obj)
	assert.NoError(t, err)
	assert.JSONEq(t, `{}`, string(jsonData))
}

func TestIsEmptyValue(t *testing.T) {

	// Cadena vacía
	assert.True(t, utils.IsEmptyValue(reflect.ValueOf("")))

	// Entero no vacío
	assert.False(t, utils.IsEmptyValue(reflect.ValueOf(1)))

	// Estructura con valores
	type TestStruct struct {
		Name string
	}
	testStruct := TestStruct{
		Name: "John",
	}
	assert.False(t, utils.IsEmptyValue(reflect.ValueOf(testStruct)))

	// Estructura vacía
	testStruct = TestStruct{}
	assert.True(t, utils.IsEmptyValue(reflect.ValueOf(testStruct)))

	// Booleano falso
	assert.True(t, utils.IsEmptyValue(reflect.ValueOf(false)))

	// Booleano verdadero
	assert.False(t, utils.IsEmptyValue(reflect.ValueOf(true)))

	// Tiempo vacío
	assert.True(t, utils.IsEmptyValue(reflect.ValueOf(time.Time{})))

	// Tiempo no vacío
	assert.False(t, utils.IsEmptyValue(reflect.ValueOf(time.Now())))
}

func TestOmitEmptyFields_MapHandling(t *testing.T) {
	// Caso 1: Map con valores válidos y valores que deben ser omitidos (vacíos o nil)
	testMap := map[string]interface{}{
		"key1": "valid value",
		"key2": "",     // Cadena vacía, debe ser omitida
		"key3": nil,    // Nil, debe ser omitido
		"key4": "data", // Valor válido, debe permanecer
	}

	result := utils.OmitEmptyFields(reflect.ValueOf(testMap))

	// Verificamos si el resultado no es nil y lo convertimos a map[string]interface{}
	if resultMap, ok := result.(map[string]interface{}); ok {
		// Solo los valores válidos deben estar presentes
		assert.Len(t, resultMap, 2)
		assert.Equal(t, "valid value", resultMap["key1"])
		assert.Equal(t, "data", resultMap["key4"])
		assert.NotContains(t, resultMap, "key2")
		assert.NotContains(t, resultMap, "key3")
	} else {
		t.Fatalf("Expected a map[string]interface{}, got %T", result)
	}

	// Caso 2: Map vacío debe retornar nil
	emptyMap := map[string]interface{}{}
	result = utils.OmitEmptyFields(reflect.ValueOf(emptyMap))
	assert.Nil(t, result)
}

func TestOmitEmptyFields_SliceAndArrayHandling(t *testing.T) {
	// Caso 1: Slice con valores válidos y valores vacíos
	testSlice := []interface{}{
		"valid", // Valor válido
		"",      // Cadena vacía, debe ser omitida
		nil,     // Nil, debe ser omitido
		123,     // Valor válido
	}

	result := utils.OmitEmptyFields(reflect.ValueOf(testSlice))

	// Verificamos si el resultado no es nil y lo convertimos a []interface{}
	if resultSlice, ok := result.([]interface{}); ok {
		assert.Len(t, resultSlice, 2)
		assert.Equal(t, "valid", resultSlice[0])
		assert.Equal(t, 123, resultSlice[1])
	} else {
		t.Fatalf("Expected a []interface{}, got %T", result)
	}

	// Caso 2: Slice vacío debe retornar nil
	emptySlice := []string{}
	result = utils.OmitEmptyFields(reflect.ValueOf(emptySlice))
	assert.Nil(t, result)

	// Caso 3: Array con valores válidos y valores vacíos
	testArray := [3]interface{}{
		"first", "",
		nil,
	}
	result = utils.OmitEmptyFields(reflect.ValueOf(testArray))

	// Verificamos si el resultado no es nil y lo convertimos a []interface{}
	if resultArray, ok := result.([]interface{}); ok {
		assert.Len(t, resultArray, 1)
		assert.Equal(t, "first", resultArray[0])
	} else {
		t.Fatalf("Expected a []interface{}, got %T", result)
	}

	// Caso 4: Array vacío debe retornar nil
	emptyArray := [0]string{}
	result = utils.OmitEmptyFields(reflect.ValueOf(emptyArray))
	assert.Nil(t, result)
}

func TestOmitEmptyFields_StructWithoutJsonTag(t *testing.T) {
	// Caso: Estructura con campo sin etiqueta json
	type TestStructNoTag struct {
		FieldWithoutTag string
	}
	objNoTag := TestStructNoTag{
		FieldWithoutTag: "No Tag Value",
	}

	resultMap := utils.OmitEmptyFields(reflect.ValueOf(objNoTag)).(map[string]interface{})

	// Debe usar el nombre del campo como clave en lugar de una etiqueta json
	assert.Equal(t, "No Tag Value", resultMap["FieldWithoutTag"])
}

func TestOmitEmptyFields_InvalidAndNilValues(t *testing.T) {
	// Caso 1: Reflect value no válido (cubre if !v.IsValid())
	var invalidValue reflect.Value
	result := utils.OmitEmptyFields(invalidValue)
	assert.Nil(t, result) // Debe retornar nil ya que el valor no es válido

	// Caso 2: Reflect value es nil (cubre if v.IsNil())
	var nilPointer *struct{}
	result = utils.OmitEmptyFields(reflect.ValueOf(nilPointer))
	assert.Nil(t, result) // Debe retornar nil ya que el valor es nil
}

func TestOmitEmptyFields_NonExportedFields(t *testing.T) {
	// Caso: Estructura con un campo no exportado (nombre empieza con minúscula)
	type TestStruct struct {
		ExportedField    string `json:"exported"`
		nonExportedField string // No debe ser considerado
	}

	obj := TestStruct{
		ExportedField:    "Exported Value",
		nonExportedField: "Non-exported Value",
	}

	result := utils.OmitEmptyFields(reflect.ValueOf(obj))

	// Verificamos que solo el campo exportado esté presente
	resultMap := result.(map[string]interface{})
	assert.Len(t, resultMap, 1)
	assert.Equal(t, "Exported Value", resultMap["exported"])
	assert.NotContains(t, resultMap, "nonExportedField") // Debe ser omitido
}

func TestOmitEmptyFields_JsonTagDash(t *testing.T) {
	// Caso: Campo con json:"-" debe ser omitido
	type TestStruct struct {
		IncludedField string `json:"included"`
		ExcludedField string `json:"-"` // No debe aparecer
	}

	obj := TestStruct{
		IncludedField: "Included Value",
		ExcludedField: "Excluded Value",
	}

	result := utils.OmitEmptyFields(reflect.ValueOf(obj))

	// Verificamos que solo el campo con json:"-" sea omitido
	resultMap := result.(map[string]interface{})
	assert.Len(t, resultMap, 1)
	assert.Equal(t, "Included Value", resultMap["included"])
	assert.NotContains(t, resultMap, "ExcludedField") // Debe ser omitido
}

func TestOmitEmptyFields_NilFieldValues(t *testing.T) {
	// Caso: Campos con valor nil deben ser omitidos
	type TestStruct struct {
		Field1 *string `json:"field1"`
		Field2 *string `json:"field2,omitempty"`
	}

	// Solo Field2 tiene valor
	field2Value := "Non-nil value"
	obj := TestStruct{
		Field1: nil, // Debe ser omitido
		Field2: &field2Value,
	}

	result := utils.OmitEmptyFields(reflect.ValueOf(obj))

	// Verificamos que el campo nil sea omitido y solo aparezca Field2
	resultMap := result.(map[string]interface{})
	assert.Len(t, resultMap, 1)
	assert.Equal(t, "Non-nil value", resultMap["field2"])
}

func TestIsEmptyValue_UintAndFloatHandling(t *testing.T) {
	// Caso 1: Uint vacío
	assert.True(t, utils.IsEmptyValue(reflect.ValueOf(uint(0))))
	// Uint no vacío
	assert.False(t, utils.IsEmptyValue(reflect.ValueOf(uint(123))))

	// Caso 2: Float vacío
	assert.True(t, utils.IsEmptyValue(reflect.ValueOf(float64(0))))
	// Float no vacío
	assert.False(t, utils.IsEmptyValue(reflect.ValueOf(float64(1.234))))
}

func TestIsEmptyValue_ComplexCase(t *testing.T) {
	// Caso de estructura compleja
	type ComplexStruct struct {
		IntField    int
		StringField string
		BoolField   bool
		TimeField   time.Time
	}

	// Estructura con campos vacíos
	emptyComplexStruct := ComplexStruct{
		IntField:    0,
		StringField: "",
		BoolField:   false,
		TimeField:   time.Time{}, // Zero time
	}

	// Todos los campos están vacíos, debería ser true
	assert.True(t, utils.IsEmptyValue(reflect.ValueOf(emptyComplexStruct)))

	// Estructura con campos no vacíos
	nonEmptyComplexStruct := ComplexStruct{
		IntField:    1,
		StringField: "Non-empty",
		BoolField:   true,
		TimeField:   time.Now(), // Tiempo no vacío
	}

	// Al menos un campo no está vacío, debería ser false
	assert.False(t, utils.IsEmptyValue(reflect.ValueOf(nonEmptyComplexStruct)))
}

func TestIsEmptyValue_ReturnFalseForUnhandledTypes(t *testing.T) {
	// Caso: Valor de tipo complejo (complex128), no manejado específicamente por el switch
	complexValue := complex(1, 2)
	assert.False(t, utils.IsEmptyValue(reflect.ValueOf(complexValue))) // Debe retornar false ya que no es un valor vacío

	// Caso: Valor de tipo func, tampoco manejado específicamente
	funcValue := func() {}
	assert.False(t, utils.IsEmptyValue(reflect.ValueOf(funcValue))) // Debe retornar false ya que no es nil ni vacío
}
