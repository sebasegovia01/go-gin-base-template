package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sebasegovia01/base-template-go-gin/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para PresentialChannelRepositoryInterface
type MockPresentialChannelRepository struct {
	mock.Mock
}

func (m *MockPresentialChannelRepository) GetAllPresentialChannels() ([]models.PresentialChannel, error) {
	args := m.Called()
	return args.Get(0).([]models.PresentialChannel), args.Error(1)
}

func (m *MockPresentialChannelRepository) GetPresentialChannelByID(channelIdentifier string) (models.PresentialChannel, error) {
	args := m.Called(channelIdentifier)
	return args.Get(0).(models.PresentialChannel), args.Error(1)
}

func TestGetPresentialChannels_AllChannelsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockPresentialChannelRepository)

	expectedChannels := []models.PresentialChannel{
		{
			PresentialChannelIdentifier: "CH001",
			PresentialChannelType:       "Type1",
			StreetName:                  "Main St",
			PresentialAttentionHours:    "",
			PresentialFromDatetime:      "",
			PresentialToDatetime:        "",
			PresentialWeekDayCode:       "",
		},
		{
			PresentialChannelIdentifier: "CH002",
			PresentialChannelType:       "Type2",
			StreetName:                  "Second St",
			PresentialAttentionHours:    "",
			PresentialFromDatetime:      "",
			PresentialToDatetime:        "",
			PresentialWeekDayCode:       "",
		},
	}

	mockRepo.On("GetAllPresentialChannels").Return(expectedChannels, nil)

	controller := NewPresentialChannelController(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/channels", nil)

	controller.GetPresentialChannels(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedResponse := gin.H{
		"data": []interface{}{
			map[string]interface{}{
				"presentialChannelIdentifier": "CH001",
				"presentialChannelType":       "Type1",
				"streetName":                  "Main St",
				"presentialAttentionHours":    "",
				"presentialFromDatetime":      "",
				"presentialToDatetime":        "",
				"presentialWeekDayCode":       "",
			},
			map[string]interface{}{
				"presentialChannelIdentifier": "CH002",
				"presentialChannelType":       "Type2",
				"streetName":                  "Second St",
				"presentialAttentionHours":    "",
				"presentialFromDatetime":      "",
				"presentialToDatetime":        "",
				"presentialWeekDayCode":       "",
			},
		},
	}

	assert.Equal(t, expectedResponse, response)
	mockRepo.AssertExpectations(t)
}

func TestGetPresentialChannels_ChannelByIDSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockPresentialChannelRepository)

	expectedChannel := models.PresentialChannel{
		PresentialChannelIdentifier: "CH001",
		PresentialChannelType:       "Type1",
		StreetName:                  "Main St",
		PresentialAttentionHours:    "",
		PresentialFromDatetime:      "",
		PresentialToDatetime:        "",
		PresentialWeekDayCode:       "",
	}

	mockRepo.On("GetPresentialChannelByID", "CH001").Return(expectedChannel, nil)

	controller := NewPresentialChannelController(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: "CH001"})
	c.Request, _ = http.NewRequest(http.MethodGet, "/channels/CH001", nil)

	controller.GetPresentialChannels(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedResponse := gin.H{
		"data": map[string]interface{}{
			"presentialChannelIdentifier": "CH001",
			"presentialChannelType":       "Type1",
			"streetName":                  "Main St",
			"presentialAttentionHours":    "",
			"presentialFromDatetime":      "",
			"presentialToDatetime":        "",
			"presentialWeekDayCode":       "",
		},
	}

	assert.Equal(t, expectedResponse, response)
	mockRepo.AssertExpectations(t)
}

func TestGetPresentialChannels_ChannelByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockPresentialChannelRepository)

	mockRepo.On("GetPresentialChannelByID", "CH001").Return(models.PresentialChannel{}, errors.New("Presential Channel not found"))

	controller := NewPresentialChannelController(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: "CH001"})
	c.Request, _ = http.NewRequest(http.MethodGet, "/channels/CH001", nil)

	controller.GetPresentialChannels(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	expectedResponse := gin.H{"error": "Presential Channel not found"}
	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	mockRepo.AssertExpectations(t)
}

func TestGetPresentialChannels_AllChannelsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockPresentialChannelRepository)

	mockRepo.On("GetAllPresentialChannels").Return([]models.PresentialChannel{}, errors.New("error fetching channels"))

	controller := NewPresentialChannelController(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/channels", nil)

	controller.GetPresentialChannels(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	expectedResponse := gin.H{"error": "Error fetching Presential Channels"}
	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	mockRepo.AssertExpectations(t)
}
