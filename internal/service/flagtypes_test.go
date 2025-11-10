package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDataFlag(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Test data
	validJSON := map[string]interface{}{
		"title":  "Test Book",
		"author": "Test Author",
		"metadata": map[string]interface{}{
			"isbn":  "123-456-789",
			"pages": float64(300), // JSON numbers are float64
		},
	}

	t.Run("valid JSON file", func(t *testing.T) {
		// Create a temporary JSON file
		jsonData, _ := json.Marshal(validJSON)
		testFile := filepath.Join(tempDir, "valid.json")
		err := os.WriteFile(testFile, jsonData, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Test the flag
		var target map[string]interface{}
		flag := &DataFlag{Target: &target}

		err = flag.Set(testFile)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Check that the data was parsed correctly
		if target["title"] != "Test Book" {
			t.Errorf("Expected title 'Test Book', got: %v", target["title"])
		}
		if target["author"] != "Test Author" {
			t.Errorf("Expected author 'Test Author', got: %v", target["author"])
		}
	})

	t.Run("empty filename", func(t *testing.T) {
		var target map[string]interface{}
		flag := &DataFlag{Target: &target}

		err := flag.Set("")
		if err == nil {
			t.Fatal("Expected error for empty filename")
		}

		expectedError := "filename cannot be empty"
		if err.Error() != expectedError {
			t.Errorf("Expected error: %s, got: %s", expectedError, err.Error())
		}
	})

	t.Run("file not found", func(t *testing.T) {
		var target map[string]interface{}
		flag := &DataFlag{Target: &target}

		err := flag.Set("nonexistent.json")
		if err == nil {
			t.Fatal("Expected error for nonexistent file")
		}

		if !contains(err.Error(), "unable to read file 'nonexistent.json': no such file or directory") {
			t.Errorf("Expected file not found error, got: %s", err.Error())
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		// Create a file with invalid JSON
		invalidJSON := `{"title": "Test", "missing": "closing brace"`
		testFile := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(testFile, []byte(invalidJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		var target map[string]interface{}
		flag := &DataFlag{Target: &target}

		err = flag.Set(testFile)
		if err == nil {
			t.Fatal("Expected error for invalid JSON")
		}

		if !contains(err.Error(), "invalid JSON in") {
			t.Errorf("Expected invalid JSON error, got: %s", err.Error())
		}
	})

	t.Run("string representation", func(t *testing.T) {
		target := map[string]interface{}{
			"title": "Test Book",
		}
		flag := &DataFlag{Target: &target}

		str := flag.String()
		expected := `{"title":"Test Book"}`
		if str != expected {
			t.Errorf("Expected string: %s, got: %s", expected, str)
		}
	})

	t.Run("type", func(t *testing.T) {
		flag := &DataFlag{}
		if flag.Type() != "data" {
			t.Errorf("Expected type 'data', got: %s", flag.Type())
		}
	})
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		(len(str) > len(substr) &&
			(str[:len(substr)] == substr ||
				str[len(str)-len(substr):] == substr ||
				containsInMiddle(str, substr))))
}

func containsInMiddle(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
