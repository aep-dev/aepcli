package service

import (
	"fmt"
	"strings"

	"github.com/aep-dev/aepcli/internal/openapi"
)

type ServiceDefinition struct {
	ServerURL string
	Resources map[string]Resource
}

func GetServiceDefinition(api *openapi.OpenAPI) (*ServiceDefinition, error) {
	resources := make(map[string]Resource)
	for _, r := range api.Components.Schemas {
		if r.XAEPResource != nil {
			addResourceToMap(r.XAEPResource, resources, api)
		}
	}
	// get the first serverURL url
	serverURL := ""
	for _, s := range api.Servers {
		serverURL = s.URL
	}
	if serverURL == "" {
		return nil, fmt.Errorf("no servers found in the OpenAPI definition. Cannot find a server to send a request to.")
	}

	return &ServiceDefinition{
		ServerURL: serverURL,
		Resources: resources,
	}, nil
}

func (s *ServiceDefinition) GetResource(resource string) (*Resource, error) {
	r, ok := (*s).Resources[resource]
	if !ok {
		return nil, fmt.Errorf("Resource %s not found. Resources available: %q", resource, (*s).Resources)
	}
	return &r, nil
}

func addResourceToMap(r *openapi.XAEPResource, resourceMap map[string]Resource, api *openapi.OpenAPI) (*Resource, error) {
	plural := strings.ToLower(r.Plural)
	if r, ok := resourceMap[plural]; ok {
		return &r, nil
	}
	parents := []*Resource{}
	for _, p := range r.Parents {
		s, ok := api.Components.Schemas[p]
		if !ok {
			return nil, fmt.Errorf("Resource %q parent %q not found", r.Singular, p)
		}
		if s.XAEPResource == nil {
			return nil, fmt.Errorf("Resource %q parent %q does not have the x-aep-resource annotation", r.Singular, p)
		}
		parentResource, err := addResourceToMap(s.XAEPResource, resourceMap, api)
		if err != nil {
			return nil, fmt.Errorf("Resource %q parent %q does not have the x-aep-resource annotation", r.Singular, p)
		}
		parents = append(parents, parentResource)
	}
	resource := Resource{
		Singular: r.Singular,
		Plural:   r.Plural,
		Parent:   parents,
	}
	resourceMap[strings.ToLower(r.Plural)] = resource
	return &resource, nil
}
