package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"google.golang.org/api/option"
)

type PubSubClientInterface interface {
	Topic(topicID string) PubSubTopicInterface
	Close() error
}

type PubSubTopicInterface interface {
	Publish(ctx context.Context, msg *pubsub.Message) PubSubResultInterface
	Exists(ctx context.Context) (bool, error)
	Stop()
}

type PubSubResultInterface interface {
	Get(ctx context.Context) (string, error)
}

type pubSubClientWrapper struct {
	client *pubsub.Client
}

// Implementación del método Topic para devolver PubSubTopicInterface
func (w *pubSubClientWrapper) Topic(topicID string) PubSubTopicInterface {
	// Envuelve el *pubsub.Topic dentro de pubSubTopicWrapper
	return &pubSubTopicWrapper{Topic: w.client.Topic(topicID)}
}

// Implementación del método Close para cerrar el cliente
func (w *pubSubClientWrapper) Close() error {
	return w.client.Close()
}

type pubSubTopicWrapper struct {
	Topic *pubsub.Topic
}

func (w *pubSubTopicWrapper) Publish(ctx context.Context, msg *pubsub.Message) PubSubResultInterface {
	return &pubSubResultWrapper{w.Topic.Publish(ctx, msg)}
}

func (w *pubSubTopicWrapper) Exists(ctx context.Context) (bool, error) {
	return w.Topic.Exists(ctx)
}

func (w *pubSubTopicWrapper) Stop() {
	w.Topic.Stop()
}

type pubSubResultWrapper struct {
	*pubsub.PublishResult
}

func (w *pubSubResultWrapper) Get(ctx context.Context) (string, error) {
	return w.PublishResult.Get(ctx)
}

// PubSubPublishServiceInterface defines the interface for publishing messages
type PubSubPublishServiceInterface interface {
	PublishMessage(message json.RawMessage) error
	Close() error
}

// PubSubPublishService is the actual service for publishing messages
type PubSubPublishService struct {
	client PubSubClientInterface
	topics map[string]PubSubTopicInterface
	cfg    *config.Config
}

// PubSubClientCreator is a function type for creating PubSub clients
type PubSubClientCreator func(ctx context.Context, projectID string, opts ...option.ClientOption) (*pubsub.Client, error)

// Default implementation using the actual pubsub.NewClient
var defaultPubSubClientCreator PubSubClientCreator = pubsub.NewClient

var pubsubNewClient = func(ctx context.Context, projectID string, opts ...option.ClientOption) (PubSubClientInterface, error) {
	client, err := defaultPubSubClientCreator(ctx, projectID, opts...)
	if err != nil {
		return nil, err
	}
	// Devuelve el envoltorio que implementa la interfaz PubSubClientInterface
	return &pubSubClientWrapper{client: client}, nil
}

// NewPubSubPublishService initializes a new PubSubPublishService
func NewPubSubPublishService(cfg *config.Config) (*PubSubPublishService, error) {
	log.Printf("Initializing PubSub publish service for project: %s", cfg.GCPCredentials.ProjectID)

	ctx := context.Background()

	credJSON := []byte(cfg.GCPCredentials.JSON)

	client, err := pubsubNewClient(ctx, cfg.GCPCredentials.ProjectID, option.WithCredentialsJSON(credJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	service := &PubSubPublishService{
		client: client,
		topics: make(map[string]PubSubTopicInterface),
		cfg:    cfg,
	}

	// Initialize topics
	for _, topicName := range cfg.Topics {
		topic := service.client.Topic(topicName)
		exists, err := topic.Exists(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check if topic %s exists: %w", topicName, err)
		}
		if !exists {
			return nil, fmt.Errorf("topic %s does not exist", topicName)
		}
		service.topics[topicName] = topic
	}

	log.Printf("PubSubPublishService initialized with %d topics", len(service.topics))
	return service, nil
}

// PublishMessage publishes a message to all topics
func (s *PubSubPublishService) PublishMessage(message json.RawMessage) error {
	ctx := context.Background()
	var publishErrors []error

	for topicName, topic := range s.topics {
		result := topic.Publish(ctx, &pubsub.Message{
			Data: message,
		})

		id, err := result.Get(ctx)
		if err != nil {
			publishErrors = append(publishErrors, fmt.Errorf("failed to publish message to topic %s: %w", topicName, err))
		} else {
			log.Printf("Published message to topic %s; msg ID: %v", topicName, id)
		}
	}

	if len(publishErrors) > 0 {
		return fmt.Errorf("errors occurred while publishing: %v", publishErrors)
	}

	return nil
}

// Close closes the PubSub client and stops all topics
func (s *PubSubPublishService) Close() error {
	for _, topic := range s.topics {
		topic.Stop()
	}
	return s.client.Close()
}
