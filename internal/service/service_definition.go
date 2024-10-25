package service

import (
	"errors"
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
	for name, s := range api.Components.Schemas {
		if s.XAEPResource != nil {
			_, err := addResourceToMap(s, resources, api)
			if err != nil {
				return nil, fmt.Errorf("error adding resource %q to map: %v", name, err)
			}
		}
	}
	// get the first serverURL url
	serverURL := ""
	for _, s := range api.Servers {
		serverURL = s.URL
	}
	if serverURL == "" {
		return nil, errors.New("no servers found in the OpenAPI definition. Cannot find a server to send a request to")
	}

	return &ServiceDefinition{
		ServerURL: serverURL,
		Resources: resources,
	}, nil
}

func (s *ServiceDefinition) GetResource(resource string) (*Resource, error) {
	r, ok := (*s).Resources[resource]
	if !ok {
		return nil, fmt.Errorf("Resource %s not found. Resources available: %v", resource, (*s).Resources)
	}
	return &r, nil
}

func addResourceToMap(s openapi.Schema, resourceMap map[string]Resource, api *openapi.OpenAPI) (*Resource, error) {
	r := s.XAEPResource
	if r == nil {
		return nil, fmt.Errorf("schema does not have the x-aep-resource annotation")
	}
	plural := strings.ToLower(r.Plural)
	if r, ok := resourceMap[plural]; ok {
		return &r, nil
	}
	parents := []*Resource{}
	for _, p := range r.Parents {
		s, ok := api.Components.Schemas[p]
		if !ok {
			return nil, fmt.Errorf("resource %q parent %q not found", r.Singular, p)
		}
		parentResource, err := addResourceToMap(s, resourceMap, api)
		if err != nil {
			return nil, fmt.Errorf("error parsing resource %q parent %q: %v", r.Singular, p, err)
		}
		parents = append(parents, parentResource)
	}

	resource := Resource{
		Singular: r.Singular,
		Plural:   r.Plural,
		Parents:  parents,
		Pattern:  strings.Split(r.Patterns[0], "/")[1:],
		Schema:   s,
	}
	resourceMap[strings.ToLower(r.Plural)] = resource
	return &resource, nil
}
