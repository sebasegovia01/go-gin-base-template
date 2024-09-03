package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sebasegovia01/base-template-go-gin/models"
)

func TransformCustomerData(data map[string]interface{}) (*models.Customer, error) {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid payload structure")
	}

	customer := &models.Customer{
		PersonalIdentification: models.PersonalCustomerIdentification{},
		PersonalAdditionalInfo: models.PersonalCustomerAdditionalInfo{},
	}

	if name, ok := payload["NAME"].(string); ok {
		names := strings.Fields(name)
		if len(names) > 0 {
			customer.PersonalIdentification.CustomerFirstName = names[0]
		}
		if len(names) > 1 {
			customer.PersonalIdentification.CustomerLastName = names[len(names)-1]
		}
		if len(names) > 2 {
			middleName := strings.Join(names[1:len(names)-1], " ")
			if middleName != "" {
				customer.PersonalIdentification.CustomerMiddleName = middleName
			}
		}
	}

	if id, ok := payload["ID"].(string); ok && id != "" {
		customer.PersonalAdditionalInfo.LegalRepresentativeIdentification = id
	}

	if createdAt, ok := payload["CREATED_AT"].(string); ok {
		t, err := time.Parse(time.RFC3339, createdAt)
		if err == nil && !t.IsZero() {
			customer.PersonalIdentification.CustomerInitDate = t
		}
	}

	// more mapping here if needed...

	return customer, nil
}

// CustomMarshalJSON es una función auxiliar para omitir campos vacíos en la serialización JSON
func CustomMarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(omitEmptyFields(v))
}

func omitEmptyFields(v interface{}) interface{} {
	switch value := v.(type) {
	case models.Customer:
		m := make(map[string]interface{})
		if !isEmptyStruct(value.PersonalIdentification) {
			m["personalIdentification"] = omitEmptyFields(value.PersonalIdentification)
		}
		if !isEmptyStruct(value.PersonalAdditionalInfo) {
			m["personalAdditionalInfo"] = omitEmptyFields(value.PersonalAdditionalInfo)
		}
		if !isEmptyStruct(value.LegalEntityIdentification) {
			m["legalEntityIdentification"] = omitEmptyFields(value.LegalEntityIdentification)
		}
		if !isEmptyStruct(value.LegalEntityAdditionalInfo) {
			m["legalEntityAdditionalInfo"] = omitEmptyFields(value.LegalEntityAdditionalInfo)
		}
		return m
	case models.PersonalCustomerIdentification, models.PersonalCustomerAdditionalInfo,
		models.LegalEntityIdentification, models.LegalEntityAdditionalInfo:
		return omitEmptyFieldsInStruct(value)
	default:
		return v
	}
}

func omitEmptyFieldsInStruct(v interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	val := reflect.ValueOf(v)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.IsZero() {
			m[val.Type().Field(i).Tag.Get("json")] = field.Interface()
		}
	}
	return m
}

func isEmptyStruct(v interface{}) bool {
	return reflect.ValueOf(v).IsZero()
}
