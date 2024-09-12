package config_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/google"
)

func TestLoad(t *testing.T) {
	// Mockear la función LoadEnvFile para evitar la carga de un archivo .env real
	config.LoadEnvFile = func(filenames ...string) error {
		return nil // Simular éxito en la carga del archivo .env
	}
	defer func() {
		config.LoadEnvFile = godotenv.Load
	}()

	// Guardar y restaurar las variables de entorno originales
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			pair := strings.SplitN(env, "=", 2)
			os.Setenv(pair[0], pair[1])
		}
	}()

	t.Run("Valid configuration", func(t *testing.T) {
		// Configurar las variables de entorno necesarias
		os.Clearenv()
		os.Setenv("ENV", "prod")
		os.Setenv("PORT", "8080")
		os.Setenv("URL_MS_PRESENTIAL_CHANELS", "valid-url")
		os.Setenv("URL_MS_AUTOMATIC_TELLER_MACHINES", "valid-url-atm")
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", `{"type": "service_account"}`)

		// Mockear las credenciales de GCP
		config.GetGoogleCredentialsFromJSON = func(ctx context.Context, jsonData []byte, scopes ...string) (*google.Credentials, error) {
			return &google.Credentials{}, nil
		}
		defer func() {
			config.GetGoogleCredentialsFromJSON = google.CredentialsFromJSON
		}()

		// Ejecutar la función Load
		cfg, err := config.Load()

		// Validar los resultados
		assert.Nil(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "8080", cfg.ServerAddress)
		assert.NotNil(t, cfg.GCPCredentials)
	})

	t.Run("Missing environment variable for URLs", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENV", "dev")
		os.Setenv("PORT", "8080")

		// Probar cuando falta la variable de entorno `URL_MS_PRESENTIAL_CHANELS`
		_, err := config.Load()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "URL_MS_PRESENTIAL_CHANELS is empty")

		// Probar cuando falta la variable de entorno `URL_MS_AUTOMATIC_TELLER_MACHINES`
		os.Setenv("URL_MS_PRESENTIAL_CHANELS", "valid-url")
		_, err = config.Load()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "URL_MS_AUTOMATIC_TELLER_MACHINES is empty")
	})

	t.Run("Invalid GCP credentials", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENV", "prod")
		os.Setenv("PORT", "8080")
		os.Setenv("URL_MS_PRESENTIAL_CHANELS", "valid-url")
		os.Setenv("URL_MS_AUTOMATIC_TELLER_MACHINES", "valid-url-atm")
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", `{"invalid": "json"}`)

		// Mockear las credenciales de GCP para devolver un error
		config.GetGoogleCredentialsFromJSON = func(ctx context.Context, jsonData []byte, scopes ...string) (*google.Credentials, error) {
			return nil, fmt.Errorf("failed to parse GCP credentials")
		}
		defer func() {
			config.GetGoogleCredentialsFromJSON = google.CredentialsFromJSON
		}()

		_, err := config.Load()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error parsing GCP credentials JSON")
	})

	t.Run("ENVIRONMENTS error reading", func(t *testing.T) {
		os.Clearenv()
		longString := strings.Repeat("A", 65536) // Crear un string demasiado largo para bufio.Scanner
		os.Setenv("ENVIRONMENTS", longString)

		// Probar la carga fallida
		err := config.LoadFromEnvironments()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error reading ENVIRONMENTS")
	})

	t.Run("Load with missing environment variable", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENV", "dev")
		os.Setenv("PORT", "8080")

		// Probar cuando falta la variable de entorno `URL_MS_PRESENTIAL_CHANELS`
		_, err := config.Load()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "URL_MS_PRESENTIAL_CHANELS is empty")
	})

	t.Run("Error loading .env file", func(t *testing.T) {
		// Simular un error al cargar el archivo .env
		config.LoadEnvFile = func(filenames ...string) error {
			return fmt.Errorf("failed to load .env")
		}

		defer func() {
			config.LoadEnvFile = godotenv.Load
		}()

		// Limpiar las variables de entorno para asegurarnos de que se intente cargar el .env
		os.Clearenv()
		os.Setenv("ENV", "dev") // Esto forzará la carga de un archivo .env

		// Ejecutar la función Load y verificar el error
		_, err := config.Load()

		// Validar que se devuelva el error correcto
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error loading .env file")
		assert.Contains(t, err.Error(), "failed to load .env")
	})

}

func TestLoadFromEnvironments(t *testing.T) {
	// Guardar y restaurar las variables de entorno originales
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			pair := strings.SplitN(env, "=", 2)
			os.Setenv(pair[0], pair[1])
		}
	}()

	t.Run("Valid ENVIRONMENTS", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENVIRONMENTS", "KEY1=VALUE1\nKEY2=VALUE2\nKEY3=VALUE3")
		err := config.LoadFromEnvironments()
		assert.Nil(t, err)
		assert.Equal(t, "VALUE1", os.Getenv("KEY1"))
		assert.Equal(t, "VALUE2", os.Getenv("KEY2"))
		assert.Equal(t, "VALUE3", os.Getenv("KEY3"))
	})
}

func TestLoad_OtherEnvironmentError(t *testing.T) {
	// Guardar la función original y restaurarla después
	originalLoadFromEnvironments := config.LoadFromEnvironments
	defer func() {
		config.LoadFromEnvironments = originalLoadFromEnvironments
	}()

	// Mockear LoadFromEnvironments para que devuelva un error diferente
	config.LoadFromEnvironments = func() error {
		return fmt.Errorf("some other error")
	}

	// Mockear LoadEnvFile para que no interfiera
	originalLoadEnvFile := config.LoadEnvFile
	config.LoadEnvFile = func(filenames ...string) error {
		return nil
	}
	defer func() {
		config.LoadEnvFile = originalLoadEnvFile
	}()

	// Limpiar variables de entorno
	os.Clearenv()

	// Ejecutar Load()
	_, err := config.Load()

	// Verificar que se devuelve el error esperado
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error loading from ENVIRONMENTS: some other error")
}

func TestLoad_UnknownEnvironment(t *testing.T) {
	// Guardar las funciones originales
	originalLoadEnvFile := config.LoadEnvFile
	originalLoadFromEnvironments := config.LoadFromEnvironments
	defer func() {
		config.LoadEnvFile = originalLoadEnvFile
		config.LoadFromEnvironments = originalLoadFromEnvironments
	}()

	// Mockear LoadEnvFile y LoadFromEnvironments
	config.LoadEnvFile = func(filenames ...string) error {
		return nil // Simular éxito en la carga del archivo .env
	}
	config.LoadFromEnvironments = func() error {
		return fmt.Errorf("ENVIRONMENTS variable is empty")
	}

	// Guardar las variables de entorno originales
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			pair := strings.SplitN(env, "=", 2)
			os.Setenv(pair[0], pair[1])
		}
	}()

	// Limpiar y configurar las variables de entorno
	os.Clearenv()
	os.Setenv("ENV", "unknown")
	os.Setenv("PORT", "8080")
	os.Setenv("URL_MS_PRESENTIAL_CHANELS", "http://example.com/channels")
	os.Setenv("URL_MS_AUTOMATIC_TELLER_MACHINES", "http://example.com/atms")

	// Capturar la salida estándar
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Ejecutar Load()
	cfg, err := config.Load()

	// Restaurar la salida estándar
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Verificaciones
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, enums.Dev, cfg.Environment)
	assert.Contains(t, buf.String(), "Warning: Unknown environment 'unknown' specified, defaulting to Dev")
}
