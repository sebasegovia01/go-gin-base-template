package config

import (
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

	env := enums.Environment(os.Getenv("ENV"))

	// Verificar si estamos en un entorno de desarrollo
	if env != enums.Prod {
		err := godotenv.Load()
		if err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

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
