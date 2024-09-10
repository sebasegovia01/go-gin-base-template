package repositories

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/datastore"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/models"
	"google.golang.org/api/option"
)

type PresentialChannelInterface interface {
	Get(ctx context.Context, key *datastore.Key, dst interface{}) error
	GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error)
	Close() error
}

// PresentialChannelCreator is an interface for creating Datastore clients
type PresentialChannelCreator interface {
	NewClientWithDatabase(ctx context.Context, projectID string, databaseID string, opts ...option.ClientOption) (PresentialChannelInterface, error)
}

// RealPresentialChannelCreator implements PresentialChannelCreator for real Datastore clients
type RealPresentialChannelCreator struct{}

func (r RealPresentialChannelCreator) NewClientWithDatabase(ctx context.Context, projectID string, databaseID string, opts ...option.ClientOption) (PresentialChannelInterface, error) {
	return datastore.NewClientWithDatabase(ctx, projectID, databaseID, opts...)
}

type PresentialChannelRepository struct {
	client    PresentialChannelInterface
	dbName    string
	namespace string
	kind      string
}

// NewDatastorePresentialChannelRepository creates and returns a new PresentialChannelRepository
func NewDatastorePresentialChannelRepository(client PresentialChannelInterface, dbName, namespace, kind string) *PresentialChannelRepository {
	return &PresentialChannelRepository{
		client:    client,
		dbName:    dbName,
		namespace: namespace,
		kind:      kind,
	}
}

// NewPresentialChannelClient creates a new Datastore client with the specified database
func NewPresentialChannelClient(cfg *config.Config, creator PresentialChannelCreator) (PresentialChannelInterface, error) {
	log.Printf("Initializing Presential Channel repository for project: %s", cfg.GCPCredentials.ProjectID)
	ctx := context.Background()

	var clientOpts []option.ClientOption

	clientOpts = append(clientOpts, option.WithCredentialsJSON([]byte(cfg.GCPCredentials.JSON)))

	// Utiliza NewClientWithDatabase para especificar el databaseID (dbName)
	client, err := creator.NewClientWithDatabase(ctx, cfg.GCPCredentials.ProjectID, cfg.DataStoreDBName, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Datastore client: %w", err)
	}

	return client, nil
}

// GetAllPresentialChannels retrieves all PresentialChannel entities from the Datastore
func (r *PresentialChannelRepository) GetAllPresentialChannels() ([]models.PresentialChannel, error) {
	ctx := context.Background()
	var channels []models.PresentialChannel

	// Query to get all Presential Channels, using namespace and kind
	query := datastore.NewQuery(r.kind).Namespace(r.namespace)

	// Execute the query
	_, err := r.client.GetAll(ctx, query, &channels)
	if err != nil {
		return nil, fmt.Errorf("failed to get all Presential Channels: %w", err)
	}

	return channels, nil
}

// GetPresentialChannelByID retrieves a single PresentialChannel entity from the Datastore by PresentialChannelIdentifier
func (r *PresentialChannelRepository) GetPresentialChannelByID(channelIdentifier string) (models.PresentialChannel, error) {
	ctx := context.Background()
	var channel models.PresentialChannel

	// Create a key for the Presential Channel, specifying the kind and namespace
	key := datastore.NameKey(r.kind, channelIdentifier, nil)
	key.Namespace = r.namespace

	// Get the Presential Channel entity from Datastore
	err := r.client.Get(ctx, key, &channel)
	if err != nil {
		return models.PresentialChannel{}, fmt.Errorf("failed to get Presential Channel by identifier: %w", err)
	}

	return channel, nil
}

// Close closes the Datastore client connection
func (r *PresentialChannelRepository) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to close datastore client: %w", err)
	}
	return nil
}
