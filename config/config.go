package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	ServerAddress string
}

func Load() (*Config, error) {
	// Cargar variables de entorno desde el archivo .env
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
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
		DatabaseURL:   dbURL,
		ServerAddress: os.Getenv("SERVER_ADDRESS"),
	}, nil
}
