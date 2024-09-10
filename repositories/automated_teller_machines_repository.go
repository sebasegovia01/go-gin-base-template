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

type AutomatedTellerMachineInterface interface {
	Get(ctx context.Context, key *datastore.Key, dst interface{}) error
	GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error)
	Close() error
}

// AutomatedTellerMachineCreator is an interface for creating Datastore clients
type AutomatedTellerMachineCreator interface {
	NewClientWithDatabase(ctx context.Context, projectID string, databaseID string, opts ...option.ClientOption) (AutomatedTellerMachineInterface, error)
}

// RealAutomatedTellerMachineCreator implements AutomatedTellerMachineCreator for real Datastore clients
type RealAutomatedTellerMachineCreator struct{}

func (r RealAutomatedTellerMachineCreator) NewClientWithDatabase(ctx context.Context, projectID string, databaseID string, opts ...option.ClientOption) (AutomatedTellerMachineInterface, error) {
	return datastore.NewClientWithDatabase(ctx, projectID, databaseID, opts...)
}

type AutomatedTellerMachineRepository struct {
	client    AutomatedTellerMachineInterface
	dbName    string
	namespace string
	kind      string
}

// NewDatastoreATMRepository creates and returns a new DatastoreATMRepository
func NewDatastoreATMRepository(client AutomatedTellerMachineInterface, dbName, namespace, kind string) *AutomatedTellerMachineRepository {
	return &AutomatedTellerMachineRepository{
		client:    client,
		dbName:    dbName,
		namespace: namespace,
		kind:      kind,
	}
}

// NewAutomatedTellerMachineClient creates a new Datastore client with the specified database
func NewAutomatedTellerMachineClient(cfg *config.Config, creator AutomatedTellerMachineCreator) (AutomatedTellerMachineInterface, error) {
	log.Printf("Initializing Automated Teller Machine repository for project: %s", cfg.GCPCredentials.ProjectID)
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

// GetAllATMs retrieves all ATM entities from the Datastore
func (r *AutomatedTellerMachineRepository) GetAllATMs() ([]models.AutomatedTellerMachine, error) {
	ctx := context.Background()
	var atms []models.AutomatedTellerMachine

	// Query to get all ATMs, using namespace and kind
	query := datastore.NewQuery(r.kind).Namespace(r.namespace)

	// Execute the query
	_, err := r.client.GetAll(ctx, query, &atms)
	if err != nil {
		return nil, fmt.Errorf("failed to get all ATMs: %w", err)
	}

	return atms, nil
}

// GetATMByID retrieves a single ATM entity from the Datastore by ATMIdentifier
func (r *AutomatedTellerMachineRepository) GetATMByID(atmIdentifier string) (models.AutomatedTellerMachine, error) {
	ctx := context.Background()
	var atm models.AutomatedTellerMachine

	// Create a key for the ATM, specifying the kind and namespace
	key := datastore.NameKey(r.kind, atmIdentifier, nil)
	key.Namespace = r.namespace

	// Get the ATM entity from Datastore
	err := r.client.Get(ctx, key, &atm)
	if err != nil {
		return models.AutomatedTellerMachine{}, fmt.Errorf("failed to get ATM by identifier: %w", err)
	}

	return atm, nil
}

// Close closes the Datastore client connection
func (r *AutomatedTellerMachineRepository) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to close datastore client: %w", err)
	}
	return nil
}
