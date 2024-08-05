package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"cloud.google.com/go/pubsub"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"google.golang.org/api/option"
)

type PubSubService struct {
	client         *pubsub.Client
	subscription   *pubsub.Subscription
	storageService *StorageService
}

func NewPubSubService(cfg *config.Config, storageService *StorageService) (*PubSubService, error) {
	ctx := context.Background()

	client, err := createPubSubClient(ctx, cfg.GCPProjectID, cfg.GCPCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create PubSub client: %w", err)
	}

	subscription := client.Subscription(cfg.GCPSubID)
	if err := verifySubscription(ctx, subscription); err != nil {
		client.Close()
		return nil, err
	}

	log.Printf("Successfully initialized PubSub service for project %s and subscription %s", cfg.GCPProjectID, cfg.GCPSubID)

	return &PubSubService{
		client:         client,
		subscription:   subscription,
		storageService: storageService,
	}, nil
}

func createPubSubClient(ctx context.Context, projectID, credentialsFile string) (*pubsub.Client, error) {
	if err := logServiceAccountInfo(credentialsFile); err != nil {
		return nil, err
	}

	return pubsub.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsFile))
}

func logServiceAccountInfo(credentialsFile string) error {
	creds, err := os.ReadFile(credentialsFile)
	if err != nil {
		return fmt.Errorf("error reading credentials file: %w", err)
	}

	var credInfo map[string]interface{}
	if err := json.Unmarshal(creds, &credInfo); err != nil {
		return fmt.Errorf("error parsing credentials: %w", err)
	}

	log.Printf("Using service account: %s", credInfo["client_email"])
	return nil
}

func verifySubscription(ctx context.Context, subscription *pubsub.Subscription) error {
	exists, err := subscription.Exists(ctx)
	if err != nil {
		return fmt.Errorf("error checking subscription existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("subscription does not exist")
	}
	return nil
}

func (s *PubSubService) ReceiveMessages() error {
	ctx := context.Background()
	var receivedCount int64

	log.Printf("Starting to receive messages...")

	err := s.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		newCount := atomic.AddInt64(&receivedCount, 1)
		log.Printf("Received message: ID=%s", msg.ID)
		log.Printf("Total messages received: %d", newCount)

		var storageEvent map[string]interface{}
		if err := json.Unmarshal(msg.Data, &storageEvent); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			msg.Nack()
			return
		}

		bucketName, ok := storageEvent["bucket"].(string)
		if !ok {
			log.Printf("Bucket name not found in message")
			msg.Nack()
			return
		}

		objectName, ok := storageEvent["name"].(string)
		if !ok {
			log.Printf("Object name not found in message")
			msg.Nack()
			return
		}

		if err := s.storageService.ProcessFile(bucketName, objectName); err != nil {
			log.Printf("Error processing file: %v", err)
			msg.Nack()
			return
		}

		msg.Ack()
	})

	if err != nil {
		return fmt.Errorf("error receiving messages: %w", err)
	}

	return nil
}

func (s *PubSubService) Close() {
	if err := s.client.Close(); err != nil {
		log.Printf("Error closing PubSub client: %v", err)
	}
}
