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

type PubSubPublishServiceInterface interface {
	PublishMessage(message json.RawMessage) error
}
type PubSubPublishService struct {
	client *pubsub.Client
	topics map[string]*pubsub.Topic
	cfg    *config.Config
}

func NewPubSubPublishService(cfg *config.Config) (*PubSubPublishService, error) {
	log.Printf("Initializing PubSub publish service for project: %s", cfg.GCPCredentials.ProjectID)

	ctx := context.Background()

	// Convertir las credenciales de string a []byte
	credJSON := []byte(cfg.GCPCredentials.JSON)

	client, err := pubsub.NewClient(ctx, cfg.GCPCredentials.ProjectID, option.WithCredentialsJSON(credJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	service := &PubSubPublishService{
		client: client,
		topics: make(map[string]*pubsub.Topic),
		cfg:    cfg,
	}

	// Initialize topics
	for _, topicName := range cfg.Topics {
		topic := client.Topic(topicName)
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

func (s *PubSubPublishService) PublishMessage(message json.RawMessage) error {
	// Usamos directamente el mensaje como json.RawMessage
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

func (s *PubSubPublishService) Close() error {
	for _, topic := range s.topics {
		topic.Stop()
	}
	return s.client.Close()
}
