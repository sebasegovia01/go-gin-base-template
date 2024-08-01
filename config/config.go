package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sebasegovia01/base-template-go-gin/enums"
)

type Config struct {
	DatabaseURL    string
	ServerAddress  string
	Environment    enums.Environment
	GCPProjectID   string
	GCPSubID       string
	GCPCredentials string
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

	// Leer GOOGLE_APPLICATION_CREDENTIALS del .env y establecerlo como variable de entorno
	gcpCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if gcpCreds != "" {
		absPath, err := filepath.Abs(gcpCreds)
		if err != nil {
			return nil, fmt.Errorf("error getting absolute path for credentials: %w", err)
		}
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", absPath)
	}

	// Construir la URL de la base de datos
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	return &Config{
		DatabaseURL:    dbURL,
		ServerAddress:  os.Getenv("PORT"),
		Environment:    env,
		GCPProjectID:   os.Getenv("GCP_PROJECT_ID"),
		GCPSubID:       os.Getenv("GCP_SUB_ID"),
		GCPCredentials: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
	}, nil
}
