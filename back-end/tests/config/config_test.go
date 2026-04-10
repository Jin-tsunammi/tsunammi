package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mm/config"
)

func TestMustLoadConfig_PanicOnMissingEnvFile(t *testing.T) {
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	assert.Panics(t, func() {
		config.NewConfig()
	}, "MustLoadConfig should panic when .env file is missing")
}

func TestMustLoadConfig_PanicOnMissingRequiredEnvVars(t *testing.T) {
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	envFilePath := filepath.Join(tempDir, ".env")
	envContent := ``

	err = os.WriteFile(envFilePath, []byte(envContent), 0644)
	require.NoError(t, err)

	defer func() {
		if _, err := os.Stat(envFilePath); err == nil {
			os.Remove(envFilePath)
		}
	}()

	assert.Panics(t, func() {
		config.NewConfig()
	}, "MustLoadConfig should panic when required environment variables are missing")
}

func TestMustLoadConfig_Success(t *testing.T) {
	envVars := map[string]string{
		"ENVIRONMENT":            "stage",
		"HTTP_PORT":              "8080",
		"HTTP_HOST":              "localhost",
		"POSTGRES_USER":          "user",
		"POSTGRES_PASSWORD":      "password",
		"POSTGRES_HOST":          "localhost",
		"POSTGRES_PORT":          "5432",
		"POSTGRES_DATABASE":      "db",
		"POSTGRES_SSLMODE":       "disable",
		"DB_MAX_CONNS":           "10",
		"DB_MIN_CONNS":           "1",
		"WALLET_ENCRYPTION_KEY":  "key",
		"ACCOUNT_ENCRYPTION_KEY": "key",
		"SOLANA_RPC_URL":         "rpcurl",
		"JWT_SECRET":             "secret",
	}

	for k, v := range envVars {
		err := os.Setenv(k, v)
		if err != nil {
			require.NoError(t, err)
		}
	}

	defer func() {
		for k := range envVars {
			err := os.Unsetenv(k)
			if err != nil {
				require.NoError(t, err)
			}
		}
	}()

	assert.NotPanics(t, func() {
		config.NewConfig()
	}, "MustLoadConfig should not panic when all required environment variables are set")
}
