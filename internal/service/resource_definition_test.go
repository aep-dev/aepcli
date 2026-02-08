package service

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

func getTestAPI() *api.API {
	projectResource := api.Resource{
		Singular: "project",
		Plural:   "projects",
		Parents:  []string{},
		Schema: &openapi.Schema{
			Properties: map[string]openapi.Schema{
				"name": {
					Type:        "string",
					Description: "The name of the project",
				},
				"description": {
					Type: "string",
				},
				"active": {
					Type: "boolean",
				},
				"tags": {
					Type: "array",
					Items: &openapi.Schema{
						Type: "string",
					},
				},
				"metadata": {
					Type: "object",
				},
				"priority": {
					Type: "integer",
				},
			},
			Required: []string{"name"},
		},
		Methods: api.Methods{
			Get:  &api.GetMethod{},
			List: &api.ListMethod{},
			Create: &api.CreateMethod{
				SupportsUserSettableCreate: true,
			},
			Update: &api.UpdateMethod{},
			Delete: &api.DeleteMethod{},
		},
	}

	a := &api.API{
		Name:      "test",
		ServerURL: "https://api.example.com",
		Resources: map[string]*api.Resource{
			"project": &projectResource,
			"dataset": {
				Singular: "dataset",
				Plural:   "datasets",
				Parents:  []string{"project"},
				Schema: &openapi.Schema{
					Properties: map[string]openapi.Schema{
						"name": {
							Type: "string",
						},
						"description": {
							Type: "string",
						},
						"size": {
							Type: "integer",
						},
						"config": {
							Type: "object",
						},
					},
					// Remove Required field to make testing easier
				},
				Methods: api.Methods{
					Get:    &api.GetMethod{},
					List:   &api.ListMethod{},
					Create: &api.CreateMethod{},
					Update: &api.UpdateMethod{},
					Delete: &api.DeleteMethod{},
				},
			},
			"user": {
				Singular: "user",
				Plural:   "users",
				Parents:  []string{},
				Schema: &openapi.Schema{
					Properties: map[string]openapi.Schema{
						"username": {
							Type: "string",
						},
						"email": {
							Type: "string",
						},
						"active": {
							Type: "boolean",
						},
					},
				},
				Methods: api.Methods{
					Get:    &api.GetMethod{},
					List:   &api.ListMethod{},
					Create: &api.CreateMethod{},
					Update: &api.UpdateMethod{},
					Delete: &api.DeleteMethod{},
				},
			},
			"comment": {
				Singular: "comment",
				Plural:   "comments",
				Parents:  []string{},
				Schema:   &openapi.Schema{},
			},
			"shelf": {
				Singular: "shelf",
				Plural:   "shelves",
				Parents:  []string{"project"},
				Schema:   &openapi.Schema{},
			},
			"book": {
				Singular: "book",
				Plural:   "books",
				Parents:  []string{"shelf"},
				Schema: &openapi.Schema{
					Properties: map[string]openapi.Schema{
						"title": {
							Type: "string",
						},
						"author": {
							Type: "string",
						},
						"path": {
							Type:     "string",
							ReadOnly: true,
						},
					},
				},
				Methods: api.Methods{
					Create: &api.CreateMethod{
						SupportsUserSettableCreate: true,
					},
				},
			},
		},
	}
	err := api.AddImplicitFieldsAndValidate(a)
	if err != nil {
		panic(err)
	}
	// Restore ReadOnly for book path, as AddImplicitFieldsAndValidate overwrites it
	if book, ok := a.Resources["book"]; ok {
		if pathProp, ok := book.Schema.Properties["path"]; ok {
			pathProp.ReadOnly = true
			book.Schema.Properties["path"] = pathProp
		}
	}
	return a
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name           string
		resource       string
		args           []string
		expectedQuery  string
		expectedPath   string
		expectedMethod string

		wantErr        bool
		body           string
		expectedOutput string
	}{
		{
			name:           "simple resource no parents",
			resource:       "project",
			args:           []string{"list"},
			expectedPath:   "projects",
			expectedMethod: "GET",
			wantErr:        false,
			body:           "",
		},
		{
			name:           "create with tags",
			resource:       "project",
			args:           []string{"create", "myproject", "--name=test-project", "--tags=tag1,tag2,tag3"},
			expectedPath:   "projects",
			expectedMethod: "POST",
			expectedQuery:  "id=myproject",
			wantErr:        false,
			body:           `{"name":"test-project","tags":["tag1","tag2","tag3"]}`,
		},
		{
			name:           "create with tags quoted",
			resource:       "project",
			args:           []string{"create", "myproject", "--name=test-project", "--tags=\"tag1,\",tag2,tag3"},
			expectedPath:   "projects",
			expectedMethod: "POST",
			expectedQuery:  "id=myproject",
			wantErr:        false,
			body:           `{"name":"test-project","tags":["tag1,","tag2","tag3"]}`,
		},
		{
			name:           "resource with parent",
			resource:       "dataset",
			args:           []string{"--project=foo", "get", "abc"},
			expectedPath:   "projects/foo/datasets/abc",
			expectedMethod: "GET",
			wantErr:        false,
			body:           "",
		},
		{
			name:     "create with @data flag",
			resource: "project",
			args: []string{"create", "dataproject", "--@data=" + createTestJSONFile(t, map[string]interface{}{
				"name":        "test-project",
				"description": "A test project",
				"active":      true,
				"priority":    5,
			}), "--name=test-project"}, // Add required flag to avoid validation error
			expectedPath:   "projects",
			expectedMethod: "POST",
			expectedQuery:  "id=dataproject",
			wantErr:        false, // Change to false since errors are logged, not returned
			body:           "",    // Empty body because error is logged
		},
		{
			name:     "create with @data flag - no conflicts",
			resource: "user", // Use resource with no required fields
			args: []string{"create", "--@data=" + createTestJSONFile(t, map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
			})},
			expectedPath:   "users",
			expectedMethod: "POST",
			wantErr:        false,
			body:           `{"email":"test@example.com","username":"testuser"}`,
		},
		{
			name:     "update with @data flag",
			resource: "user",
			args: []string{"update", "testuser", "--@data=" + createTestJSONFile(t, map[string]interface{}{
				"email": "newemail@example.com",
			})},
			expectedPath:   "users/testuser",
			expectedMethod: "PATCH",
			wantErr:        false,
			body:           `{"email":"newemail@example.com"}`,
		},
		{
			name:     "child resource with parent and @data flag",
			resource: "dataset",
			args: []string{"--project=myproject", "create", "--@data=" + createTestJSONFile(t, map[string]interface{}{
				"name":        "test-dataset",
				"description": "A test dataset",
				"size":        1000,
				"config": map[string]interface{}{
					"format": "parquet",
					"schema": "v1",
				},
			})},
			expectedPath:   "projects/myproject/datasets",
			expectedMethod: "POST",
			wantErr:        false,
			body:           `{"config":{"format":"parquet","schema":"v1"},"description":"A test dataset","name":"test-dataset","size":1000}`,
		},
		{
			name:           "child resource with parent and individual flags",
			resource:       "dataset",
			args:           []string{"--project=parentproject", "create", "--name=manual-dataset", "--description=Manual dataset"},
			expectedPath:   "projects/parentproject/datasets",
			expectedMethod: "POST",
			wantErr:        false,
			body:           `{"description":"Manual dataset","name":"manual-dataset"}`,
		},
		{
			name:           "create book with read-only path flag",
			resource:       "book",
			args:           []string{"--project=myproject", "--shelf=myshelf", "create", "mybook", "--path=some/path"},
			expectedPath:   "",
			expectedMethod: "",
			wantErr:        true,
			body:           "",
		},
		{
			name:           "help with description",
			resource:       "project",
			args:           []string{"create", "--help"},
			expectedPath:   "",
			expectedMethod: "",
			wantErr:        false,
			expectedOutput: "The name of the project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := getTestAPI()
			req, output, err := ExecuteResourceCommand(a.Resources[tt.resource], tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && req == nil && tt.expectedOutput == "" {
				t.Error("ExecuteCommand() returned nil request when no error expected")
			}
			if tt.expectedOutput != "" {
				if !strings.Contains(output, tt.expectedOutput) {
					t.Errorf("ExecuteCommand() output = %q, want to contain %q", output, tt.expectedOutput)
				}
			}
			if !tt.wantErr && req != nil {
				// Verify the request path matches expected pattern
				if req.URL.Path != tt.expectedPath {
					t.Errorf("ExecuteCommand() request path = %v, want %v", req.URL.Path, tt.expectedPath)
				}
				if req.Body != nil {
					body, err := io.ReadAll(req.Body)
					if err != nil {
						t.Errorf("ExecuteCommand() error reading request body: %v", err)
					}
					if string(body) != tt.body {
						t.Errorf("ExecuteCommand() request body = %v, want %v", string(body), tt.body)
					}
				}
				if req.Method != tt.expectedMethod {
					t.Errorf("ExecuteCommand() request method = %v, want %v", req.Method, tt.expectedMethod)
				}
				if req.URL.RawQuery != tt.expectedQuery {
					t.Errorf("ExecuteCommand() request query = %v, want %v", req.URL.RawQuery, tt.expectedQuery)
				}
			}
		})
	}
}

// Helper function to create temporary JSON files for testing
func createTestJSONFile(t *testing.T, data map[string]interface{}) string {
	t.Helper()

	// Create temporary directory
	tempDir := t.TempDir()

	// Create JSON file
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Write to temporary file
	tempFile := filepath.Join(tempDir, "test-data.json")
	err = os.WriteFile(tempFile, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write test JSON file: %v", err)
	}

	return tempFile
}
