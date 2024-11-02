package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigEndToEnd(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_config.toml")

	// First read should fail since file doesn't exist
	_, err := ReadConfigFromFile(testFile)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, os.ErrNotExist))

	// Create test API config
	testAPI := API{
		Name:        "test-api",
		OpenAPIPath: "/path/to/openapi.yaml",
		ServerURL:   "https://api.example.com",
		Headers:     []string{"Authorization=Bearer token"},
		PathPrefix:  "/v1",
	}

	// Write API config to file
	err = WriteAPIWithName(testFile, testAPI, false)
	assert.NoError(t, err)

	// Read config back and verify contents
	cfg, err := ReadConfigFromFile(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Len(t, cfg.APIs, 1)

	// Verify API config matches what we wrote
	savedAPI, exists := cfg.APIs[testAPI.Name]
	assert.True(t, exists)
	assert.Equal(t, testAPI, savedAPI)
}

func TestWriteAPIWithEmptyName(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_config.toml")

	testAPI := API{
		Name:        "", // Empty name should cause an error
		OpenAPIPath: "/path/to/openapi.yaml",
		ServerURL:   "https://api.example.com",
		Headers:     []string{"Authorization=Bearer token"},
		PathPrefix:  "/v1",
	}

	err := WriteAPIWithName(testFile, testAPI, false)
	assert.Error(t, err)
	assert.Equal(t, "api name cannot be empty", err.Error())
}
