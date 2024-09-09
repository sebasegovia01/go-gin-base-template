package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// Mock PubSub Client
type MockPubSubClient struct {
	mock.Mock
}

func (m *MockPubSubClient) Topic(topicID string) PubSubTopicInterface {
	args := m.Called(topicID)
	return args.Get(0).(PubSubTopicInterface)
}

func (m *MockPubSubClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock PubSub Topic
type MockPubSubTopic struct {
	mock.Mock
}

func (m *MockPubSubTopic) Publish(ctx context.Context, msg *pubsub.Message) PubSubResultInterface {
	args := m.Called(ctx, msg)
	return args.Get(0).(PubSubResultInterface)
}

func (m *MockPubSubTopic) Exists(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockPubSubTopic) Stop() {
	m.Called()
}

// Mock PubSub Result
type MockPubSubResult struct {
	mock.Mock
}

func (m *MockPubSubResult) Get(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// Store the original pubsubNewClient function
var originalPubsubNewClient = pubsubNewClient

func TestNewPubSubPublishService(t *testing.T) {
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
			JSON:      []byte(`{"type": "service_account"}`),
		},
		Topics: []string{"topic1", "topic2"},
	}

	mockClient := new(MockPubSubClient)
	mockTopic1 := new(MockPubSubTopic)
	mockTopic2 := new(MockPubSubTopic)
	mockResult1 := new(MockPubSubResult)
	mockResult2 := new(MockPubSubResult)

	mockClient.On("Topic", "topic1").Return(mockTopic1).Once()
	mockClient.On("Topic", "topic2").Return(mockTopic2).Once()

	mockTopic1.On("Exists", mock.Anything).Return(true, nil).Once()
	mockTopic2.On("Exists", mock.Anything).Return(true, nil).Once()

	mockTopic1.On("Publish", mock.Anything, mock.AnythingOfType("*pubsub.Message")).Return(mockResult1).Once()
	mockTopic2.On("Publish", mock.Anything, mock.AnythingOfType("*pubsub.Message")).Return(mockResult2).Once()
	mockResult1.On("Get", mock.Anything).Return("id1", nil).Once()
	mockResult2.On("Get", mock.Anything).Return("id2", nil).Once()

	// Replace pubsubNewClient for this test
	pubsubNewClient = func(ctx context.Context, projectID string, opts ...option.ClientOption) (PubSubClientInterface, error) {
		return mockClient, nil
	}
	t.Cleanup(func() { pubsubNewClient = originalPubsubNewClient })

	service, err := NewPubSubPublishService(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, service)

	err = service.PublishMessage(json.RawMessage(`{"key": "value"}`))
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockTopic1.AssertExpectations(t)
	mockTopic2.AssertExpectations(t)
	mockResult1.AssertExpectations(t)
	mockResult2.AssertExpectations(t)
}

func TestPubsubNewClient_Success(t *testing.T) {
	// Store the original and replace it after the test
	originalCreator := defaultPubSubClientCreator
	defer func() { defaultPubSubClientCreator = originalCreator }()

	// Replace the default creator with our test version
	defaultPubSubClientCreator = func(ctx context.Context, projectID string, opts ...option.ClientOption) (*pubsub.Client, error) {
		return &pubsub.Client{}, nil // Return a dummy client with no error
	}

	// Call the function we want to test
	client, err := pubsubNewClient(context.Background(), "test-project")

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.IsType(t, &pubSubClientWrapper{}, client)
}

func TestPubsubNewClient_Error(t *testing.T) {
	// Store the original and replace it after the test
	originalCreator := defaultPubSubClientCreator
	defer func() { defaultPubSubClientCreator = originalCreator }()

	// Replace the default creator with our test version
	defaultPubSubClientCreator = func(ctx context.Context, projectID string, opts ...option.ClientOption) (*pubsub.Client, error) {
		return nil, assert.AnError // Return nil client with an error
	}

	// Call the function we want to test
	client, err := pubsubNewClient(context.Background(), "test-project")

	// Assert the results
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, assert.AnError, err)
}

func TestNewPubSubPublishService_ErrorCreatingClient(t *testing.T) {
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
			JSON:      []byte(`{"type": "service_account"}`),
		},
		Topics: []string{"topic1"},
	}

	// Replace pubsubNewClient for this test
	pubsubNewClient = func(ctx context.Context, projectID string, opts ...option.ClientOption) (PubSubClientInterface, error) {
		return nil, errors.New("client creation error")
	}
	t.Cleanup(func() { pubsubNewClient = originalPubsubNewClient })

	service, err := NewPubSubPublishService(cfg)

	assert.Nil(t, service)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create pubsub client")
}

func TestPublishMessage_ErrorPublishing(t *testing.T) {
	mockClient := new(MockPubSubClient)
	mockTopic1 := new(MockPubSubTopic)
	mockTopic2 := new(MockPubSubTopic)
	mockResult1 := new(MockPubSubResult)
	mockResult2 := new(MockPubSubResult)

	service := &PubSubPublishService{
		client: mockClient,
		topics: map[string]PubSubTopicInterface{
			"topic1": mockTopic1,
			"topic2": mockTopic2,
		},
	}

	message := json.RawMessage(`{"key": "value"}`)

	mockTopic1.On("Publish", mock.Anything, mock.AnythingOfType("*pubsub.Message")).Return(mockResult1)
	mockTopic2.On("Publish", mock.Anything, mock.AnythingOfType("*pubsub.Message")).Return(mockResult2)
	mockResult1.On("Get", mock.Anything).Return("", errors.New("publish error 1"))
	mockResult2.On("Get", mock.Anything).Return("", errors.New("publish error 2"))

	err := service.PublishMessage(message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish error 1")
	assert.Contains(t, err.Error(), "publish error 2")

	mockTopic1.AssertExpectations(t)
	mockTopic2.AssertExpectations(t)
	mockResult1.AssertExpectations(t)
	mockResult2.AssertExpectations(t)
}

func TestClosePubsubPushService(t *testing.T) {
	mockClient := new(MockPubSubClient)
	mockTopic1 := new(MockPubSubTopic)
	mockTopic2 := new(MockPubSubTopic)

	service := &PubSubPublishService{
		client: mockClient,
		topics: map[string]PubSubTopicInterface{
			"topic1": mockTopic1,
			"topic2": mockTopic2,
		},
	}

	mockTopic1.On("Stop").Return().Once()
	mockTopic2.On("Stop").Return().Once()
	mockClient.On("Close").Return(nil).Once()

	err := service.Close()
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockTopic1.AssertExpectations(t)
	mockTopic2.AssertExpectations(t)
}

func TestNewPubSubPublishService_TopicDoesNotExist(t *testing.T) {
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
			JSON:      []byte(`{"type": "service_account"}`),
		},
		Topics: []string{"topic1", "topic2"},
	}

	mockClient := new(MockPubSubClient)
	mockTopic1 := new(MockPubSubTopic)
	mockTopic2 := new(MockPubSubTopic)

	mockClient.On("Topic", "topic1").Return(mockTopic1)
	mockClient.On("Topic", "topic2").Return(mockTopic2)

	mockTopic1.On("Exists", mock.Anything).Return(true, nil)
	mockTopic2.On("Exists", mock.Anything).Return(false, nil) // topic2 doesn't exist

	// Replace pubsubNewClient for this test
	pubsubNewClient = func(ctx context.Context, projectID string, opts ...option.ClientOption) (PubSubClientInterface, error) {
		return mockClient, nil
	}
	t.Cleanup(func() { pubsubNewClient = originalPubsubNewClient })

	_, err := NewPubSubPublishService(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "topic topic2 does not exist")

	mockClient.AssertExpectations(t)
	mockTopic1.AssertExpectations(t)
	mockTopic2.AssertExpectations(t)
}

func TestNewPubSubPublishService_TopicExistsError(t *testing.T) {
	cfg := &config.Config{
		GCPCredentials: &google.Credentials{
			ProjectID: "test-project",
			JSON:      []byte(`{"type": "service_account"}`),
		},
		Topics: []string{"topic1", "topic2"},
	}

	mockClient := new(MockPubSubClient)
	mockTopic1 := new(MockPubSubTopic)
	mockTopic2 := new(MockPubSubTopic)

	// Configure the mock to return the topics
	mockClient.On("Topic", "topic1").Return(mockTopic1)
	mockClient.On("Topic", "topic2").Return(mockTopic2)

	// Make the first topic exist, but have an error when checking the second topic
	mockTopic1.On("Exists", mock.Anything).Return(true, nil)
	mockTopic2.On("Exists", mock.Anything).Return(false, errors.New("topic existence check failed"))

	// Store the original and replace it after the test
	originalNewClient := pubsubNewClient
	defer func() { pubsubNewClient = originalNewClient }()

	// Replace pubsubNewClient with our test version
	pubsubNewClient = func(ctx context.Context, projectID string, opts ...option.ClientOption) (PubSubClientInterface, error) {
		return mockClient, nil
	}

	// Call NewPubSubPublishService
	service, err := NewPubSubPublishService(cfg)

	// Assert the results
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to check if topic topic2 exists")

	// Verify that the mocks were called as expected
	mockClient.AssertExpectations(t)
	mockTopic1.AssertExpectations(t)
	mockTopic2.AssertExpectations(t)
}

func TestPubSubClientWrapper_Topic_Real(t *testing.T) {
	// Crear un cliente de PubSub real (en un entorno controlado)
	mockPubsubClient := new(pubsub.Client)

	// Crear una instancia de pubSubClientWrapper usando el cliente real de Pub/Sub
	clientWrapper := &pubSubClientWrapper{client: mockPubsubClient}

	// Invocar el método Topic directamente en pubSubClientWrapper
	topicWrapper := clientWrapper.Topic("test-topic")

	// Asegurarse de que el resultado no sea nulo
	assert.NotNil(t, topicWrapper)

	// Asegurarse de que el tipo del resultado es pubSubTopicWrapper
	assert.IsType(t, &pubSubTopicWrapper{}, topicWrapper)

	// Asegurarse de que el tópico envuelto no sea nulo
	assert.NotNil(t, topicWrapper.(*pubSubTopicWrapper).Topic)

	// Verificar que el nombre del tópico sea el esperado
	assert.Equal(t, "projects//topics/test-topic", topicWrapper.(*pubSubTopicWrapper).Topic.String())
}

type MockPubsubClient struct {
	mock.Mock
}

func (m *MockPubsubClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestPubSubClientWrapper_Close(t *testing.T) {
	// Create a real PubSub client (in a controlled environment)
	ctx := context.Background()
	mockPubsubClient, err := pubsub.NewClient(ctx, "your-project-id")
	if err != nil {
		t.Fatalf("Failed to create PubSub client: %v", err)
	}

	// Create an instance of pubSubClientWrapper using the real Pub/Sub client
	clientWrapper := &pubSubClientWrapper{client: mockPubsubClient}

	// Call the Close method and verify that no error occurs
	err = clientWrapper.Close()
	assert.NoError(t, err)
}

func TestPubSubTopicWrapper_Publish(t *testing.T) {
	// Create a real PubSub client and topic (in a controlled environment)
	ctx := context.Background()
	mockPubsubClient, err := pubsub.NewClient(ctx, "your-project-id")
	if err != nil {
		t.Fatalf("Failed to create PubSub client: %v", err)
	}
	defer mockPubsubClient.Close()

	topic := mockPubsubClient.Topic("your-topic-id")

	// Create an instance of pubSubTopicWrapper using the real Pub/Sub topic
	topicWrapper := &pubSubTopicWrapper{Topic: topic}

	// Create a test message
	msg := &pubsub.Message{
		Data: []byte("test message"),
	}

	// Call the Publish method and verify that no error occurs
	result := topicWrapper.Publish(ctx, msg)
	assert.NotNil(t, result)

	// Verify that the result is of type pubSubResultWrapper
	_, ok := result.(*pubSubResultWrapper)
	assert.True(t, ok)
}

func TestPubSubTopicWrapper_Exists(t *testing.T) {
	// Create a real PubSub client and topic (in a controlled environment)
	ctx := context.Background()
	mockPubsubClient, err := pubsub.NewClient(ctx, "your-project-id")
	if err != nil {
		t.Fatalf("Failed to create PubSub client: %v", err)
	}
	defer mockPubsubClient.Close()

	topic := mockPubsubClient.Topic("your-topic-id")

	// Create an instance of pubSubTopicWrapper using the real Pub/Sub topic
	topicWrapper := &pubSubTopicWrapper{Topic: topic}

	// Call the Exists method and verify that no error occurs
	exists, err := topicWrapper.Exists(ctx)
	assert.NoError(t, err)

	// Verify that the topic exists
	assert.True(t, exists)
}

func TestPubSubTopicWrapper_Stop(t *testing.T) {
	// Create a real PubSub client and topic (in a controlled environment)
	ctx := context.Background()
	mockPubsubClient, err := pubsub.NewClient(ctx, "your-project-id")
	if err != nil {
		t.Fatalf("Failed to create PubSub client: %v", err)
	}
	defer mockPubsubClient.Close()

	topic := mockPubsubClient.Topic("your-topic-id")

	// Create an instance of pubSubTopicWrapper using the real Pub/Sub topic
	topicWrapper := &pubSubTopicWrapper{Topic: topic}

	// Call the Stop method
	topicWrapper.Stop()

	// Verify that the topic is stopped (you can add additional checks here)
}
