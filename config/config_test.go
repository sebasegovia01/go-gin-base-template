package config_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/google"
)

func TestLoad(t *testing.T) {
	// Mockear la función LoadEnvFile para evitar la carga de un archivo .env real
	config.LoadEnvFile = func(filenames ...string) error {
		return nil // Simular éxito en la carga del archivo .env
	}

	// Restaurar la función original después de la prueba
	defer func() {
		config.LoadEnvFile = godotenv.Load
	}()

	// Restaurar variables de entorno originales después de las pruebas
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			pair := strings.SplitN(env, "=", 2)
			os.Setenv(pair[0], pair[1])
		}
	}()

	t.Run("Valid configuration", func(t *testing.T) {
		os.Setenv("ENV", "prod")
		os.Setenv("PORT", "8080")
		os.Setenv("PUBSUB_TOPICS", "topic1,topic2")
		os.Setenv("BUCKET_NAME", "my-bucket")
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", `{"type": "service_account"}`)

		config.GetGoogleCredentialsFromJSON = func(ctx context.Context, jsonData []byte, scopes ...string) (*google.Credentials, error) {
			return &google.Credentials{}, nil
		}

		cfg, err := config.Load()

		assert.Nil(t, err)
		assert.NotNil(t, cfg)
	})
	t.Run("Set environment variables and return nil", func(t *testing.T) {
		// Simular un contenido válido para ENVIRONMENTS
		os.Setenv("ENVIRONMENTS", "KEY1=VALUE1\nKEY2=VALUE2\nKEY3=VALUE3")

		// Ejecutar loadFromEnvironments
		err := config.LoadFromEnvironments()

		// Verificar que no haya errores
		assert.Nil(t, err)

		// Verificar que las variables de entorno se hayan establecido correctamente
		assert.Equal(t, "VALUE1", os.Getenv("KEY1"))
		assert.Equal(t, "VALUE2", os.Getenv("KEY2"))
		assert.Equal(t, "VALUE3", os.Getenv("KEY3"))
	})
	t.Run("Error loading .env file", func(t *testing.T) {
		config.LoadEnvFile = func(filenames ...string) error {
			return fmt.Errorf("failed to load .env")
		}

		defer func() {
			config.LoadEnvFile = godotenv.Load
		}()

		os.Clearenv()

		_, err := config.Load()

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error loading .env file")
	})

	t.Run("Error parsing GCP credentials JSON", func(t *testing.T) {
		// Mockear las credenciales de GCP para devolver un error de parseo
		config.GetGoogleCredentialsFromJSON = func(ctx context.Context, jsonData []byte, scopes ...string) (*google.Credentials, error) {
			return nil, fmt.Errorf("failed to parse GCP credentials")
		}

		// Restaurar la función original después de la prueba
		defer func() {
			config.GetGoogleCredentialsFromJSON = google.CredentialsFromJSON
		}()

		// Simular la variable de entorno de las credenciales de GCP
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", `{"type": "invalid_service_account"}`)

		// Simular la carga exitosa del archivo .env (aunque no debería llegar aquí)
		config.LoadEnvFile = func(filenames ...string) error {
			return nil
		}

		// Ejecutar la prueba
		_, err := config.Load()

		// Asegurarse de que el error devuelto es el esperado para el parseo de las credenciales
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error parsing GCP credentials JSON")
	})
	t.Run("Error reading ENVIRONMENTS", func(t *testing.T) {
		// Simular un contenido que provoque un error en bufio.Scanner
		longString := strings.Repeat("A", 65536) // Supera el tamaño del buffer predeterminado
		os.Setenv("ENVIRONMENTS", longString)

		// Restaurar la función original después de la prueba
		defer func() {
			os.Unsetenv("ENVIRONMENTS")
		}()

		err := config.LoadFromEnvironments()

		// Verificar que el error es el esperado
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error reading ENVIRONMENTS")
	})
	t.Run("Error loading from ENVIRONMENTS", func(t *testing.T) {
		// Simular un error diferente en loadFromEnvironments
		config.LoadFromEnvironments = func() error {
			return fmt.Errorf("some other error")
		}

		// Restaurar la función original después de la prueba
		defer func() {
			config.LoadFromEnvironments = config.LoadFromEnvironments
		}()

		os.Clearenv()

		_, err := config.Load()

		// Verificar que el error es el esperado
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error loading from ENVIRONMENTS")
	})
}
