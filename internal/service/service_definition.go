package service

import (
	"fmt"
	"strings"
)

type Resource struct {
	Plural string
}

type ServiceDefinition struct {
	ServerURL string
	Resources map[string]Resource
}

func GetServiceDefinition(api *OpenAPI) (*ServiceDefinition, error) {
	resources := make(map[string]Resource)
	for _, r := range api.Components.Schemas {
		if r.XAEPResource != nil {
			resources[strings.ToLower(r.XAEPResource.Plural)] = Resource{Plural: r.XAEPResource.Plural}
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
