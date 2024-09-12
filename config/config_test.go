package config_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/sebasegovia01/base-template-go-gin/config"
	"github.com/sebasegovia01/base-template-go-gin/enums"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/google"
)

// ErrorReader es un io.Reader que siempre devuelve un error
type ErrorReader struct{}

func (er ErrorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("forced read error")
}

func TestLoad(t *testing.T) {
	// Mockear la función LoadEnvFile para evitar la carga de un archivo .env real
	originalLoadEnvFile := config.LoadEnvFile
	config.LoadEnvFile = func(filenames ...string) error {
		return nil // Simular éxito en la carga del archivo .env
	}
	defer func() {
		config.LoadEnvFile = originalLoadEnvFile
	}()

	// Mockear LoadFromEnvironments para simular un entorno controlado
	originalLoadFromEnvironments := config.LoadFromEnvironments
	config.LoadFromEnvironments = func() error {
		return fmt.Errorf("ENVIRONMENTS variable is empty")
	}
	defer func() {
		config.LoadFromEnvironments = originalLoadFromEnvironments
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

	t.Run("Load with all required environment variables", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENV", "dev")
		os.Setenv("PORT", "8080")
		os.Setenv("DATASTORE_DB_NAME", "test-db")
		os.Setenv("DATASTORE_NAMESPACE", "test-namespace")
		os.Setenv("DATASTORE_PRESENTIAL_CHANNELS_KIND", "channel-kind")
		os.Setenv("DATASTORE_AUTOMATIC_TELLER_MACHINES_KIND", "atm-kind")

		cfg, err := config.Load()

		assert.Nil(t, err)
		assert.NotNil(t, cfg)
		if cfg != nil {
			assert.Equal(t, "8080", cfg.ServerAddress)
			assert.Equal(t, enums.Dev, cfg.Environment)
			assert.Equal(t, "test-db", cfg.DataStoreDBName)
			assert.Equal(t, "test-namespace", cfg.DataStoreNamespace)
			assert.Equal(t, "channel-kind", cfg.DatastorePresentialChannelKind)
			assert.Equal(t, "atm-kind", cfg.DatastoreAutomaticTellerMachineKind)
		}
	})

	t.Run("Load with missing DATASTORE_NAMESPACE", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENV", "dev")
		os.Setenv("PORT", "8080")
		os.Setenv("DATASTORE_DB_NAME", "test-db")
		os.Setenv("DATASTORE_PRESENTIAL_CHANNELS_KIND", "channel-kind")
		os.Setenv("DATASTORE_AUTOMATIC_TELLER_MACHINES_KIND", "atm-kind")

		_, err := config.Load()

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "DATASTORE_NAMESPACE is empty")
	})

	t.Run("Load with missing DATASTORE_PRESENTIAL_CHANNELS_KIND", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENV", "dev")
		os.Setenv("PORT", "8080")
		os.Setenv("DATASTORE_DB_NAME", "test-db")
		os.Setenv("DATASTORE_NAMESPACE", "test-namespace")
		os.Setenv("DATASTORE_AUTOMATIC_TELLER_MACHINES_KIND", "atm-kind")

		_, err := config.Load()

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "DATASTORE_PRESENTIAL_CHANNELS_KIND is empty")
	})

	t.Run("Load with missing DATASTORE_AUTOMATIC_TELLER_MACHINES_KIND", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENV", "dev")
		os.Setenv("PORT", "8080")
		os.Setenv("DATASTORE_DB_NAME", "test-db")
		os.Setenv("DATASTORE_NAMESPACE", "test-namespace")
		os.Setenv("DATASTORE_PRESENTIAL_CHANNELS_KIND", "channel-kind")

		_, err := config.Load()

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "DATASTORE_AUTOMATIC_TELLER_MACHINES_KIND is empty")
	})

	t.Run("Unknown environment defaults to Dev", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENV", "unknown")
		os.Setenv("PORT", "8080")
		os.Setenv("DATASTORE_NAMESPACE", "test-namespace")
		os.Setenv("DATASTORE_PRESENTIAL_CHANNELS_KIND", "channel-kind")
		os.Setenv("DATASTORE_AUTOMATIC_TELLER_MACHINES_KIND", "atm-kind")

		cfg, err := config.Load()

		assert.Nil(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, enums.Dev, cfg.Environment)
	})

	t.Run("Error parsing GCP credentials JSON", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENV", "dev")
		os.Setenv("PORT", "8080")
		os.Setenv("DATASTORE_NAMESPACE", "test-namespace")
		os.Setenv("DATASTORE_PRESENTIAL_CHANNELS_KIND", "channel-kind")
		os.Setenv("DATASTORE_AUTOMATIC_TELLER_MACHINES_KIND", "atm-kind")
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", `{"invalid": "json"}`)

		originalGetGoogleCredentialsFromJSON := config.GetGoogleCredentialsFromJSON
		config.GetGoogleCredentialsFromJSON = func(ctx context.Context, jsonData []byte, scopes ...string) (*google.Credentials, error) {
			return nil, fmt.Errorf("invalid credentials")
		}
		defer func() {
			config.GetGoogleCredentialsFromJSON = originalGetGoogleCredentialsFromJSON
		}()

		_, err := config.Load()

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error parsing GCP credentials JSON")
	})

	t.Run("Error loading from ENVIRONMENTS (other error)", func(t *testing.T) {
		// Guardar la función original
		originalLoadFromEnvironments := config.LoadFromEnvironments
		defer func() {
			config.LoadFromEnvironments = originalLoadFromEnvironments
		}()

		// Mockear LoadFromEnvironments para que devuelva un error personalizado
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
	})
}

func TestLoadFromEnvironments(t *testing.T) {
	// Restaurar variables de entorno originales después de las pruebas
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			pair := strings.SplitN(env, "=", 2)
			os.Setenv(pair[0], pair[1])
		}
	}()

	t.Run("Empty ENVIRONMENTS variable", func(t *testing.T) {
		os.Clearenv()
		err := config.LoadFromEnvironments()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "ENVIRONMENTS variable is empty")
	})

	t.Run("Valid ENVIRONMENTS content", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENVIRONMENTS", "KEY1=VALUE1\nKEY2=VALUE2\nKEY3=VALUE3")
		err := config.LoadFromEnvironments()
		assert.Nil(t, err)
		assert.Equal(t, "VALUE1", os.Getenv("KEY1"))
		assert.Equal(t, "VALUE2", os.Getenv("KEY2"))
		assert.Equal(t, "VALUE3", os.Getenv("KEY3"))
	})

	t.Run("ENVIRONMENTS with empty lines and spaces", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENVIRONMENTS", "KEY1 = VALUE1\n\nKEY2=  VALUE2  \n  KEY3  =VALUE3")
		err := config.LoadFromEnvironments()
		assert.Nil(t, err)
		assert.Equal(t, "VALUE1", os.Getenv("KEY1"))
		assert.Equal(t, "VALUE2", os.Getenv("KEY2"))
		assert.Equal(t, "VALUE3", os.Getenv("KEY3"))
	})

	t.Run("ENVIRONMENTS with invalid format", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ENVIRONMENTS", "KEY1=VALUE1\nINVALID_LINE\nKEY2=VALUE2")
		err := config.LoadFromEnvironments()
		assert.Nil(t, err)
		assert.Equal(t, "VALUE1", os.Getenv("KEY1"))
		assert.Equal(t, "VALUE2", os.Getenv("KEY2"))
		assert.Empty(t, os.Getenv("INVALID_LINE"))
	})

	t.Run("Error loading .env file", func(t *testing.T) {
		os.Clearenv()

		// Mockear LoadFromEnvironments para que devuelva un error indicando que ENVIRONMENTS está vacío
		originalLoadFromEnvironments := config.LoadFromEnvironments
		config.LoadFromEnvironments = func() error {
			return fmt.Errorf("ENVIRONMENTS variable is empty")
		}
		defer func() {
			config.LoadFromEnvironments = originalLoadFromEnvironments
		}()

		// Mockear LoadEnvFile para que devuelva un error
		originalLoadEnvFile := config.LoadEnvFile
		config.LoadEnvFile = func(filenames ...string) error {
			return fmt.Errorf("failed to load .env file")
		}
		defer func() {
			config.LoadEnvFile = originalLoadEnvFile
		}()

		_, err := config.Load()

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error loading .env file")
	})

	t.Run("Error reading ENVIRONMENTS", func(t *testing.T) {
		// Guardar el valor original de ENVIRONMENTS
		originalEnvironments := os.Getenv("ENVIRONMENTS")
		defer os.Setenv("ENVIRONMENTS", originalEnvironments)

		// Crear una cadena muy larga que exceda el tamaño máximo de buffer de bufio.Scanner
		longString := strings.Repeat("a", 1<<20) // 1 MB de 'a's
		os.Setenv("ENVIRONMENTS", "KEY="+longString)

		err := config.LoadFromEnvironments()

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error reading ENVIRONMENTS")
	})

}
