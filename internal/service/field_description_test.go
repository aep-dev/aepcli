package service

import (
	"strings"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

func TestRequiredFieldDescription(t *testing.T) {
	resource := &api.Resource{
		Singular: "book",
		Plural:   "books",
		Schema: &openapi.Schema{
			Properties: map[string]openapi.Schema{
				"title": {
					Type:        "string",
					Description: "The title of the book",
				},
				"author": {
					Type:        "string",
					Description: "The author of the book",
				},
			},
			Required: []string{"title"},
		},
		Methods: api.Methods{
			Create: &api.CreateMethod{},
		},
	}

	args := []string{"create", "--help"}
	_, output, _ := ExecuteResourceCommand(resource, args)

	if !strings.Contains(output, "The title of the book (required)") {
		t.Errorf("Expected description for required field 'title' to contain '(required)', but got:\n%s", output)
	}

	if strings.Contains(output, "The author of the book (required)") {
		t.Errorf("Expected description for optional field 'author' NOT to contain '(required)', but got:\n%s", output)
	}
}
