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

type Config struct {
	ServerAddress  string
	Environment    enums.Environment
	Topics         []string
	BucketName     string
	GCPCredentials *google.Credentials
}

func Load() (*Config, error) {
	var env enums.Environment

	// Intentar cargar desde ENVIRONMENTS primero (para Cloud Run)
	if err := loadFromEnvironments(); err != nil {
		fmt.Printf("No ENVIRONMENTS found or error loading from it: %v\n", err)

		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	// Determinar el entorno
	envString := os.Getenv("ENV")
	fmt.Printf("ENV value: '%s'\n", envString)

	env = enums.Environment(strings.TrimSpace(strings.ToLower(envString)))

	if env != enums.Prod && env != enums.Dev && env != enums.QA {
		fmt.Printf("Warning: Unknown environment '%s' specified, defaulting to Dev\n", envString)
		env = enums.Dev
	}

	fmt.Printf("Current environment loaded: %s\n", env)

	// Obtener los tópicos y dividirlos en un slice
	topicsString := os.Getenv("PUBSUB_TOPICS")
	topics := strings.Split(topicsString, ",")

	// Trim espacios en blanco de cada tópico
	for i, topic := range topics {
		topics[i] = strings.TrimSpace(topic)
	}

	bucketName := os.Getenv("BUCKET_NAME")

	// Manejar las credenciales de GCP
	var gcpCreds *google.Credentials
	gcpCredsJSON := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if gcpCredsJSON != "" {
		var err error
		gcpCreds, err = google.CredentialsFromJSON(context.Background(), []byte(gcpCredsJSON))
		if err != nil {
			return nil, fmt.Errorf("error parsing GCP credentials JSON: %w", err)
		}
	}

	return &Config{
		ServerAddress:  os.Getenv("PORT"),
		Environment:    env,
		Topics:         topics,
		BucketName:     bucketName,
		GCPCredentials: gcpCreds,
	}, nil
}

func loadFromEnvironments() error {
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
