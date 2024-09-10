package config

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	"golang.org/x/oauth2/google"
)

// Exporta la función para facilitar el mock en las pruebas
var LoadEnvFile = godotenv.Load

// Exporta la función que envuelve google.CredentialsFromJSON para facilitar el mock en las pruebas
var GetGoogleCredentialsFromJSON = google.CredentialsFromJSON

// Envuelve la función loadFromEnvironments para poder mockearla en pruebas
var LoadFromEnvironments = loadFromEnvironments

type Config struct {
	ServerAddress                       string
	Environment                         enums.Environment
	GCPCredentials                      *google.Credentials
	DataStoreDBName                     string
	DataStoreNamespace                  string
	DatastorePresentialChannelKind      string
	DatastoreAutomaticTellerMachineKind string
}

func Load() (*Config, error) {
	var env enums.Environment

	// Intentar cargar desde ENVIRONMENTS primero (para Cloud Run)
	if err := LoadFromEnvironments(); err != nil {
		// Revisar si el error es porque ENVIRONMENTS está vacío
		if err.Error() == "ENVIRONMENTS variable is empty" {
			// Si ENVIRONMENTS está vacío, intentamos cargar el archivo .env
			fmt.Printf("No ENVIRONMENTS found, attempting to load .env file: %v\n", err)

			if err := LoadEnvFile(); err != nil {
				return nil, fmt.Errorf("error loading .env file: %w", err)
			}
		} else {
			// Si hay otro tipo de error, devolverlo inmediatamente
			return nil, fmt.Errorf("error loading from ENVIRONMENTS: %w", err)
		}
	}

	// Continuar con la carga de configuraciones
	envString := os.Getenv("ENV")
	fmt.Printf("ENV value: '%s'\n", envString)

	env = enums.Environment(strings.TrimSpace(strings.ToLower(envString)))

	if env != enums.Prod && env != enums.Dev && env != enums.QA {
		fmt.Printf("Warning: Unknown environment '%s' specified, defaulting to Dev\n", envString)
		env = enums.Dev
	}

	fmt.Printf("Current environment loaded: %s\n", env)

	// Manejar las credenciales de GCP
	var gcpCreds *google.Credentials
	gcpCredsJSON := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if gcpCredsJSON != "" {
		var err error
		gcpCreds, err = GetGoogleCredentialsFromJSON(context.Background(), []byte(gcpCredsJSON))
		if err != nil {
			return nil, fmt.Errorf("error parsing GCP credentials JSON: %w", err)
		}
	}

	// Manejo configs datastore
	dataStoreDBName := os.Getenv("DATASTORE_DB_NAME")

	if dataStoreDBName == "" {
		fmt.Println("DATASTORE_DB_NAME is empty, setting to: 'default'")
	}

	dataStoreNamespace := os.Getenv("DATASTORE_NAMESPACE")

	if dataStoreNamespace == "" {
		return nil, fmt.Errorf("DATASTORE_NAMESPACE is empty")
	}

	dataStorePresentialChannelsKind := os.Getenv("DATASTORE_PRESENTIAL_CHANNELS_KIND")

	if dataStorePresentialChannelsKind == "" {
		return nil, fmt.Errorf("DATASTORE_PRESENTIAL_CHANNELS_KIND is empty")
	}

	dataStoreAutomaticTellerMachinesKind := os.Getenv("DATASTORE_AUTOMATIC_TELLER_MACHINES_KIND")

	if dataStoreAutomaticTellerMachinesKind == "" {
		return nil, fmt.Errorf("DATASTORE_AUTOMATIC_TELLER_MACHINES_KIND is empty")
	}

	return &Config{
		ServerAddress:                       os.Getenv("PORT"),
		Environment:                         env,
		GCPCredentials:                      gcpCreds,
		DataStoreDBName:                     dataStoreDBName,
		DataStoreNamespace:                  dataStoreNamespace,
		DatastorePresentialChannelKind:      dataStorePresentialChannelsKind,
		DatastoreAutomaticTellerMachineKind: dataStoreAutomaticTellerMachinesKind,
	}, nil
}

func loadFromEnvironments() error {
	// Define this env name in cloud run for secrets!
	environmentsContent := os.Getenv("ENVIRONMENTS")

	if environmentsContent == "" {
		return fmt.Errorf("ENVIRONMENTS variable is empty")
	}

	scanner := bufio.NewScanner(strings.NewReader(environmentsContent))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
			fmt.Printf("Set environment variable: %s\n", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading ENVIRONMENTS: %w", err)
	}

	return nil
}
