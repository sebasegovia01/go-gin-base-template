package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sebasegovia01/base-template-go-gin/enums"
)

type Config struct {
	DatabaseURL   string
	ServerAddress string
	Environment   enums.Environment
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
		ServerAddress: os.Getenv("PORT"),
		Environment:   env,
	}, nil
}
