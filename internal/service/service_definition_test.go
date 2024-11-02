package service

import (
	"testing"

	"github.com/aep-dev/aepcli/internal/openapi"
	"github.com/stretchr/testify/assert"
)

var basicOpenAPI = &openapi.OpenAPI{
	Servers: []openapi.Server{{URL: "https://api.example.com"}},
	Paths: map[string]openapi.PathItem{
		"/widgets": {
			Get: &openapi.Operation{
				Responses: map[string]openapi.Response{
					"200": {
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{
									Properties: map[string]openapi.Schema{
										"results": {
											Type: "array",
											Items: &openapi.Schema{
												Ref: "#/components/schemas/Widget",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Post: &openapi.Operation{
				Responses: map[string]openapi.Response{
					"200": {
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{
									Ref: "#/components/schemas/Widget",
								},
							},
						},
					},
				},
			},
		},
		"/widgets/{widget}": {
			Get: &openapi.Operation{
				Responses: map[string]openapi.Response{
					"200": {
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{
									Ref: "#/components/schemas/Widget",
								},
							},
						},
					},
				},
			},
			Delete: &openapi.Operation{},
			Patch: &openapi.Operation{
				Responses: map[string]openapi.Response{
					"200": {
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{
									Ref: "#/components/schemas/Widget",
								},
							},
						},
					},
				},
			},
		},
	},
	Components: openapi.Components{
		Schemas: map[string]openapi.Schema{
			"Widget": {
				Type: "object",
				Properties: map[string]openapi.Schema{
					"name": {Type: "string"},
				},
			},
		},
	},
}

func TestGetServiceDefinition(t *testing.T) {
	tests := []struct {
		name           string
		api            *openapi.OpenAPI
		serverURL      string
		expectedError  string
		validateResult func(*testing.T, *ServiceDefinition)
	}{
		{
			name: "basic resource with CRUD operations",
			api:  basicOpenAPI,
			validateResult: func(t *testing.T, sd *ServiceDefinition) {
				assert.Equal(t, "https://api.example.com", sd.ServerURL)

				widget, ok := sd.Resources["widget"]
				assert.True(t, ok, "widget resource should exist")
				assert.Equal(t, widget.Pattern, []string{"widgets", "{widget}"})
				assert.Equal(t, sd.ServerURL, "https://api.example.com")
				assert.NotNil(t, widget.GetMethod, "should have GET method")
				assert.NotNil(t, widget.ListMethod, "should have LIST method")
				assert.NotNil(t, widget.CreateMethod, "should have CREATE method")
				if widget.CreateMethod != nil {
					assert.False(t, widget.CreateMethod.SupportsUserSettableCreate, "should not support user-settable create")
				}
				assert.NotNil(t, widget.UpdateMethod, "should have UPDATE method")
				assert.NotNil(t, widget.DeleteMethod, "should have DELETE method")
			},
		},
		{
			name:      "empty openapi with server url override",
			api:       basicOpenAPI,
			serverURL: "https://override.example.com",
			validateResult: func(t *testing.T, sd *ServiceDefinition) {
				assert.Equal(t, "https://override.example.com", sd.ServerURL)
			},
		},
		{
			name: "resource with x-aep-resource annotation",
			api: &openapi.OpenAPI{
				Paths: map[string]openapi.PathItem{
					"/widgets/{widget}": {
						Get: &openapi.Operation{
							Responses: map[string]openapi.Response{
								"200": {
									Content: map[string]openapi.MediaType{
										"application/json": {
											Schema: &openapi.Schema{
												Ref: "#/components/schemas/widget",
											},
										},
									},
								},
							},
						},
					},
				},
				Servers: []openapi.Server{{URL: "https://api.example.com"}},
				Components: openapi.Components{
					Schemas: map[string]openapi.Schema{
						"widget": {
							Type: "object",
							Properties: map[string]openapi.Schema{
								"name": {Type: "string"},
							},
							XAEPResource: &openapi.XAEPResource{
								Singular: "widget",
								Plural:   "widgets",
								Patterns: []string{"/widgets/{widget}"},
							},
						},
					},
				},
			},
			validateResult: func(t *testing.T, sd *ServiceDefinition) {
				widget, ok := sd.Resources["widget"]
				assert.True(t, ok, "widget resource should exist")
				assert.Equal(t, "widget", widget.Singular)
				assert.Equal(t, "widgets", widget.Plural)
				assert.Equal(t, []string{"widgets", "{widget}"}, widget.Pattern)
			},
		},
		{
			name: "missing server URL",
			api: &openapi.OpenAPI{
				Servers: []openapi.Server{},
			},
			expectedError: "no server URL found in openapi, and none was provided",
		},
		{
			name: "resource with user-settable create ID",
			api: &openapi.OpenAPI{
				Servers: []openapi.Server{{URL: "https://api.example.com"}},
				Paths: map[string]openapi.PathItem{
					"/widgets": {
						Post: &openapi.Operation{
							Parameters: []openapi.Parameter{
								{Name: "id"},
							},
							Responses: map[string]openapi.Response{
								"200": {
									Content: map[string]openapi.MediaType{
										"application/json": {
											Schema: &openapi.Schema{
												Ref: "#/components/schemas/Widget",
											},
										},
									},
								},
							},
						},
					},
				},
				Components: openapi.Components{
					Schemas: map[string]openapi.Schema{
						"Widget": {
							Type: "object",
						},
					},
				},
			},
			validateResult: func(t *testing.T, sd *ServiceDefinition) {
				widget, ok := sd.Resources["widget"]
				assert.True(t, ok, "widget resource should exist")
				assert.True(t, widget.CreateMethod.SupportsUserSettableCreate,
					"should support user-settable create")
			},
		},
		{
			name: "OAS 2.0 style schema in response",
			api: &openapi.OpenAPI{
				Info:    openapi.Info{Version: "2.0"},
				Servers: []openapi.Server{{URL: "https://api.example.com"}},
				Paths: map[string]openapi.PathItem{
					"/widgets/{widget}": {
						Get: &openapi.Operation{
							Responses: map[string]openapi.Response{
								"200": {
									Schema: &openapi.Schema{
										Ref: "#/components/schemas/Widget",
									},
								},
							},
						},
					},
				},
				Components: openapi.Components{
					Schemas: map[string]openapi.Schema{
						"Widget": {
							Type: "object",
							Properties: map[string]openapi.Schema{
								"name": {Type: "string"},
							},
						},
					},
				},
			},
			validateResult: func(t *testing.T, sd *ServiceDefinition) {
				widget, ok := sd.Resources["widget"]
				assert.True(t, ok, "widget resource should exist")
				assert.NotNil(t, widget.GetMethod, "should have GET method")
				assert.Equal(t, []string{"widgets", "{widget}"}, widget.Pattern)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetServiceDefinition(tt.api, tt.serverURL, "")

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}
