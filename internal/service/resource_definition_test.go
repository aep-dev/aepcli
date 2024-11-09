package service

import (
	"io"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

var projectResource = api.Resource{
	Singular:     "project",
	Plural:       "projects",
	PatternElems: []string{"projects", "{project}"},
	Parents:      []*api.Resource{},
	Schema: &openapi.Schema{
		Properties: map[string]openapi.Schema{
			"name": {
				Type: "string",
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
	GetMethod:    &api.GetMethod{},
	ListMethod:   &api.ListMethod{},
	CreateMethod: &api.CreateMethod{},
	UpdateMethod: &api.UpdateMethod{},
	DeleteMethod: &api.DeleteMethod{},
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name           string
		resource       api.Resource
		args           []string
		expectedQuery  string
		expectedPath   string
		expectedMethod string
		wantErr        bool
		body           string
	}{
		{
			name:           "simple resource no parents",
			resource:       projectResource,
			args:           []string{"list"},
			expectedPath:   "projects",
			expectedMethod: "GET",
			wantErr:        false,
			body:           "",
		},
		{
			name:           "create with tags",
			resource:       projectResource,
			args:           []string{"create", "myproject", "--name=test-project", "--tags=tag1,tag2,tag3"},
			expectedPath:   "projects",
			expectedMethod: "POST",
			expectedQuery:  "id=myproject",
			wantErr:        false,
			body:           `{"name":"test-project","tags":["tag1","tag2","tag3"]}`,
		},
		{
			name:           "create with tags quoted",
			resource:       projectResource,
			args:           []string{"create", "myproject", "--name=test-project", "--tags=\"tag1,\",tag2,tag3"},
			expectedPath:   "projects",
			expectedMethod: "POST",
			expectedQuery:  "id=myproject",
			wantErr:        false,
			body:           `{"name":"test-project","tags":["tag1,","tag2","tag3"]}`,
		},
		{
			name: "resource with parent",
			resource: api.Resource{
				Singular:     "dataset",
				Plural:       "datasets",
				PatternElems: []string{"projects", "{project}", "datasets", "{dataset}"},
				Parents:      []*api.Resource{&projectResource},
				Schema:       &openapi.Schema{},
				GetMethod:    &api.GetMethod{},
				ListMethod:   &api.ListMethod{},
				CreateMethod: &api.CreateMethod{},
				UpdateMethod: &api.UpdateMethod{},
				DeleteMethod: &api.DeleteMethod{},
			},
			args:           []string{"--project=foo", "get", "abc"},
			expectedPath:   "projects/foo/datasets/abc",
			expectedMethod: "GET",
			wantErr:        false,
			body:           "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _, err := ExecuteResourceCommand(&tt.resource, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && req == nil {
				t.Error("ExecuteCommand() returned nil request when no error expected")
			}
			if !tt.wantErr {
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
