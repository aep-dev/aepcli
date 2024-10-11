package service

import (
	"fmt"
	"strings"
)

type Resource struct {
	Plural string
}

type ServiceDefinition struct {
	Resources map[string]Resource
}

func GetServiceDefinition(api *OpenAPI) (*ServiceDefinition, error) {
	resources := make(map[string]Resource)
	for _, r := range api.Components.Schemas {
		if r.XAEPResource != nil {
			resources[strings.ToLower(r.XAEPResource.Plural)] = Resource{Plural: r.XAEPResource.Plural}
		}
	}
	return &ServiceDefinition{Resources: resources}, nil
}

func (s *ServiceDefinition) GetResource(resource string) (*Resource, error) {
	r, ok := (*s).Resources[resource]
	if !ok {
		return nil, fmt.Errorf("Resource %s not found. Resources available: %q", resource, (*s).Resources)
	}
	return &r, nil
}
