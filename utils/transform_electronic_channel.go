package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sebasegovia01/base-template-go-gin/models"
)

// TransformElectronicChannelData transforms the incoming data for electronic channels into the corresponding models
func TransformElectronicChannelData(data *map[string]interface{}) (*models.ElectronicChannels, error) {
	payload, ok := (*data)["payload"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid payload structure")
	}

	electronicChannels := &models.ElectronicChannels{}
	webChannel := models.WebChannel{}
	emailChannel := models.EmailChannel{}
	socialMediaChannel := models.SocialMediaChannel{}

	// BOPERS_WEB_CHANNEL
	if webChannelData, ok := payload["BOPERS_WEB_CHANNEL"].(map[string]interface{}); ok {
		if webChannelType, ok := webChannelData["WEB_CHANNEL_TYPE"].(string); ok {
			webChannel.WebChannelType = webChannelType
		}
		if webURLAddress, ok := webChannelData["WEB_URL_ADDRESS"].(string); ok {
			webChannel.WebURLAddress = webURLAddress
		}
		if webAvailableServices, ok := webChannelData["WEB_AVAILABLE_SERVICES"].(string); ok {
			webChannel.WebAvailableServices = webAvailableServices
		}
		if webAttentionHours, ok := webChannelData["WEB_ATTENTION_HOURS"].(string); ok {
			webChannel.WebAttentionHours = webAttentionHours
		}
		if webPlatformType, ok := webChannelData["WEB_PLATFORM_TYPE"].(string); ok {
			webChannel.WebPlatformType = webPlatformType
		}
	}

	// BOPERS_EMAIL_CHANNEL
	if emailChannelData, ok := payload["BOPERS_EMAIL_CHANNEL"].(map[string]interface{}); ok {
		if emailAvailableServices, ok := emailChannelData["EMAIL_AVAILABLE_SERVICES"].(string); ok {
			emailChannel.EmailAvailableServices = emailAvailableServices
		}
		if emailAddress, ok := emailChannelData["EMAIL_ADDRESS"].(string); ok {
			emailChannel.EmailAddress = emailAddress
		}
		if emailAttentionHours, ok := emailChannelData["EMAIL_ATTENTION_HOURS"].(string); ok {
			emailChannel.EmailAttentionHours = emailAttentionHours
		}
	}

	// BOPERS_SOCIAL_MEDIA_CHANNEL
	if socialMediaChannelData, ok := payload["BOPERS_SOCIAL_MEDIA_CHANNEL"].(map[string]interface{}); ok {
		if socialMediaAvailableServices, ok := socialMediaChannelData["SOCIAL_MEDIA_AVAILABLE_SERVICES"].(string); ok {
			socialMediaChannel.SocialMediaAvailableServices = socialMediaAvailableServices
		}
		if socialMediaAccount, ok := socialMediaChannelData["SOCIAL_MEDIA_ACCOUNT"].(string); ok {
			socialMediaChannel.SocialMediaAccount = socialMediaAccount
		}
		if socialMediaAttentionHours, ok := socialMediaChannelData["SOCIAL_MEDIA_ATTENTION_HOURS"].(string); ok {
			socialMediaChannel.SocialMediaAttentionHours = socialMediaAttentionHours
		}
	}

	// Assign channels to the main struct if not empty
	if webChannel != (models.WebChannel{}) {
		electronicChannels.WebChannel = webChannel
	}
	if emailChannel != (models.EmailChannel{}) {
		electronicChannels.EmailChannel = emailChannel
	}
	if socialMediaChannel != (models.SocialMediaChannel{}) {
		electronicChannels.SocialMediaChannel = socialMediaChannel
	}

	// Remove empty fields from the entire struct
	electronicChannelsJSON, _ := CustomMarshalJSON(electronicChannels)
	json.Unmarshal(electronicChannelsJSON, electronicChannels)

	return electronicChannels, nil
}

func OmitEmptyFields(v reflect.Value) interface{} {
	if !v.IsValid() {
		return nil
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		result := make(map[string]interface{})
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if !field.CanInterface() {
				continue // Omitir campos no exportados
			}
			structField := v.Type().Field(i)
			jsonTag := structField.Tag.Get("json")
			jsonName := strings.Split(jsonTag, ",")[0]
			if jsonName == "-" {
				continue // Omitir campos con json:"-"
			}
			if jsonName == "" {
				jsonName = structField.Name
			}
			fieldValue := OmitEmptyFields(field)
			if fieldValue != nil {
				result[jsonName] = fieldValue
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	case reflect.Map:
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)

			// Verificar si el valor es inválido o nil
			if !value.IsValid() || value.IsNil() {
				continue // Omitir valores vacíos o nil
			}

			fieldValue := OmitEmptyFields(value)
			if fieldValue == nil {
				continue // Omitir valores vacíos
			}
			// Verificar si el valor es una cadena vacía
			if str, ok := fieldValue.(string); ok && str == "" {
				continue // Omitir cadenas vacías
			}
			result[key.String()] = fieldValue
		}
		if len(result) == 0 {
			return nil
		}
		return result
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			fieldValue := OmitEmptyFields(elem)
			if fieldValue != nil {
				result = append(result, fieldValue)
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	default:
		if IsEmptyValue(v) {
			return nil
		}
		return v.Interface()
	}
}

func IsEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time).IsZero()
		}
		// Check if all fields of the struct are empty
		for i := 0; i < v.NumField(); i++ {
			if !IsEmptyValue(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}

func CustomMarshalJSON(v interface{}) ([]byte, error) {
	result := OmitEmptyFields(reflect.ValueOf(v))
	if result == nil {
		return json.Marshal(struct{}{})
	}
	return json.Marshal(result)
}
