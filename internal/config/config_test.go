package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigEndToEnd(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_config.toml")

	// First read should return a default configuration since file doesn't exist
	cfg, err := ReadConfigFromFile(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.APIs)

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
	cfg, err = ReadConfigFromFile(testFile)
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

func TestReadConfigFromFile_NoFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nonexistent_config.toml")

	// Attempt to read the config file
	cfg, err := ReadConfigFromFile(testFile)

	// Verify no error is returned and a default config is provided
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.APIs)
}
