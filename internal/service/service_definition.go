package service

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/aep-dev/aepcli/internal/openapi"
	"github.com/aep-dev/aepcli/internal/utils"
)

type ServiceDefinition struct {
	ServerURL string
	Resources map[string]*Resource
}

func GetServiceDefinition(api *openapi.OpenAPI, serverURL, pathPrefix string) (*ServiceDefinition, error) {
	slog.Debug("parsing openapi", "pathPrefix", pathPrefix)
	resourceBySingular := make(map[string]*Resource)
	customMethodsByPattern := make(map[string][]*CustomMethod)
	// we try to parse the paths to find possible resources, since
	// they may not always be annotated as such.
	for path, pathItem := range api.Paths {
		path = path[len(pathPrefix):]
		slog.Debug("path", "path", path)
		var r Resource
		var sRef *openapi.Schema
		p := getPatternInfo(path)
		if p == nil { // not a resource pattern
			slog.Debug("path is not a resource", "path", path)
			continue
		}
		slog.Debug("parsing path for resource", "path", path)
		if p.CustomMethodName != "" && p.IsResourcePattern {
			// strip the leading slash and the custom method suffix
			pattern := strings.Split(path, ":")[0][1:]
			if _, ok := customMethodsByPattern[pattern]; !ok {
				customMethodsByPattern[pattern] = []*CustomMethod{}
			}
			if pathItem.Post != nil {
				if resp, ok := pathItem.Post.Responses["200"]; ok {
					schema := api.GetSchemaFromResponse(resp)
					responseSchema := &openapi.Schema{}
					if schema != nil {
						var err error
						responseSchema, err = api.DereferenceSchema(*schema)
						if err != nil {
							return nil, fmt.Errorf("error dereferencing schema %v: %v", schema, err)
						}
					}
					schema = api.GetSchemaFromRequestBody(*pathItem.Post.RequestBody)
					requestSchema, err := api.DereferenceSchema(*schema)
					if err != nil {
						return nil, fmt.Errorf("error dereferencing schema %q: %v", schema.Ref, err)
					}
					customMethodsByPattern[pattern] = append(customMethodsByPattern[pattern], &CustomMethod{
						Name:     p.CustomMethodName,
						Method:   "POST",
						Request:  requestSchema,
						Response: responseSchema,
					})
				}
			}
			if pathItem.Get != nil {
				if resp, ok := pathItem.Get.Responses["200"]; ok {
					schema := api.GetSchemaFromResponse(resp)
					responseSchema := &openapi.Schema{}
					if schema != nil {
						var err error
						responseSchema, err = api.DereferenceSchema(*schema)
						if err != nil {
							return nil, fmt.Errorf("error dereferencing schema %v: %v", schema.Ref, err)
						}
					}
					customMethodsByPattern[pattern] = append(r.CustomMethods, &CustomMethod{
						Name:     p.CustomMethodName,
						Method:   "GET",
						Response: responseSchema,
					})
				}
			}
		} else if p.IsResourcePattern {
			// treat it like a collection pattern (update, delete, get)
			if pathItem.Delete != nil {
				r.DeleteMethod = &DeleteMethod{}
			}
			if pathItem.Get != nil {
				if resp, ok := pathItem.Get.Responses["200"]; ok {
					sRef = api.GetSchemaFromResponse(resp)
					r.GetMethod = &GetMethod{}
				}
			}
			if pathItem.Patch != nil {
				if resp, ok := pathItem.Patch.Responses["200"]; ok {
					sRef = api.GetSchemaFromResponse(resp)
					r.UpdateMethod = &UpdateMethod{}
				}
			}
		} else {
			// create method
			if pathItem.Post != nil {
				// check if there is a query parameter "id"
				if resp, ok := pathItem.Post.Responses["200"]; ok {
					sRef = api.GetSchemaFromResponse(resp)
					supportsUserSettableCreate := false
					for _, param := range pathItem.Post.Parameters {
						if param.Name == "id" {
							supportsUserSettableCreate = true
							break
						}
					}
					r.CreateMethod = &CreateMethod{SupportsUserSettableCreate: supportsUserSettableCreate}
				}
			}
			// list method
			if pathItem.Get != nil {
				if resp, ok := pathItem.Get.Responses["200"]; ok {
					respSchema := api.GetSchemaFromResponse(resp)
					if respSchema == nil {
						slog.Warn(fmt.Sprintf("resource %q has a LIST method with a response schema, but the response schema is nil.", path))
					} else {
						resolvedSchema, err := api.DereferenceSchema(*respSchema)
						if err != nil {
							return nil, fmt.Errorf("error dereferencing schema %q: %v", respSchema.Ref, err)
						}
						found := false
						for _, property := range resolvedSchema.Properties {
							if property.Type == "array" {
								sRef = property.Items
								r.ListMethod = &ListMethod{}
								found = true
								break
							}
						}
						if !found {
							slog.Warn(fmt.Sprintf("resource %q has a LIST method with a response schema, but the items field is not present or is not an array.", path))
						}
					}
				}
			}
		}
		if sRef != nil {
			// s should always be a reference to a schema in the components section.
			parts := strings.Split(sRef.Ref, "/")
			key := parts[len(parts)-1]
			dereferencedSchema, err := api.DereferenceSchema(*sRef)
			if err != nil {
				return nil, fmt.Errorf("error dereferencing schema %q: %v", sRef.Ref, err)
			}
			singular := utils.PascalCaseToKebabCase(key)
			pattern := strings.Split(path, "/")[1:]
			if !p.IsResourcePattern {
				// deduplicate the singular, if applicable
				finalSingular := singular
				parent := ""
				if len(pattern) >= 3 {
					parent = pattern[len(pattern)-3]
					parent = parent[0 : len(parent)-1] // strip curly surrounding
					if strings.HasPrefix(singular, parent) {
						finalSingular = strings.TrimPrefix(singular, parent+"-")
					}
				}
				pattern = append(pattern, fmt.Sprintf("{%s}", finalSingular))
			}
			r2, err := getOrPopulateResource(singular, pattern, dereferencedSchema, resourceBySingular, api)
			if err != nil {
				return nil, fmt.Errorf("error populating resource %q: %v", r.Singular, err)
			}
			foldResourceMethods(&r, r2)
		}
	}
	// the custom methods are trickier - because they may not respond with the schema of the resource
	// (which would allow us to map the resource via looking at it's reference), we instead will have to
	// map it by the pattern.
	// we also have to do this by longest pattern match - this helps account for situations where
	// the custom method doesn't match the resource pattern exactly with things like deduping.
	for pattern, customMethods := range customMethodsByPattern {
		found := false
		for _, r := range resourceBySingular {
			if r.GetPattern() == pattern {
				r.CustomMethods = customMethods
				found = true
				break
			}
		}
		if !found {
			slog.Debug(fmt.Sprintf("custom methods with pattern %q have no resource associated with it", pattern))
		}
	}
	if serverURL == "" {
		for _, s := range api.Servers {
			serverURL = s.URL + pathPrefix
		}
	}

	if serverURL == "" {
		return nil, fmt.Errorf("no server URL found in openapi, and none was provided")
	}

	return &ServiceDefinition{
		ServerURL: serverURL,
		Resources: resourceBySingular,
	}, nil
}

func (s *ServiceDefinition) GetResource(resource string) (*Resource, error) {
	r, ok := (*s).Resources[resource]
	if !ok {
		return nil, fmt.Errorf("Resource %q not found", resource)
	}
	return r, nil
}

type PatternInfo struct {
	// if true, the pattern represents an individual resource,
	// otherwise it represents a path to a collection of resources
	IsResourcePattern bool
	CustomMethodName  string
}

// getPatternInfo returns true if the path is an alternating pairing of collection and id,
// and returns the collection names if so.
func getPatternInfo(path string) *PatternInfo {
	customMethodName := ""
	if strings.Contains(path, ":") {
		parts := strings.Split(path, ":")
		path = parts[0]
		customMethodName = parts[1]
	}
	// we ignore the first segment, which is empty.
	pattern := strings.Split(path, "/")[1:]
	for i, segment := range pattern {
		// check if segment is wrapped in curly brackets
		wrapped := strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}")
		wantWrapped := i%2 == 1
		if wrapped != wantWrapped {
			return nil
		}
	}
	return &PatternInfo{
		IsResourcePattern: len(pattern)%2 == 0,
		CustomMethodName:  customMethodName,
	}
}

// getOrPopulateResource populates the resource via a variety of means:
// - if the resource already exists in the map, it returns it
// - if the schema has the x-aep-resource annotation, it parses the resource
// - otherwise, it attempts to infer the resource from the schema and name.
func getOrPopulateResource(singular string, pattern []string, s *openapi.Schema, resourceBySingular map[string]*Resource, api *openapi.OpenAPI) (*Resource, error) {
	if r, ok := resourceBySingular[singular]; ok {
		return r, nil
	}
	var r *Resource
	// use the X-AEP-Resource annotation to populate the resource,
	// if it exists.
	if s.XAEPResource != nil {
		parents := []*Resource{}
		for _, parentSingular := range s.XAEPResource.Parents {
			parentSchema, ok := api.Components.Schemas[parentSingular]
			if !ok {
				return nil, fmt.Errorf("resource %q parent %q not found", singular, parentSingular)
			}
			parentResource, err := getOrPopulateResource(parentSingular, []string{}, &parentSchema, resourceBySingular, api)
			if err != nil {
				return nil, fmt.Errorf("error parsing resource %q parent %q: %v", singular, parentSingular, err)
			}
			parents = append(parents, parentResource)
		}
		r = &Resource{
			Singular:     s.XAEPResource.Singular,
			Plural:       s.XAEPResource.Plural,
			Parents:      parents,
			PatternElems: strings.Split(s.XAEPResource.Patterns[0], "/")[1:],
			Schema:       s,
		}
	} else {
		// best effort otherwise
		r = &Resource{
			Schema:       s,
			PatternElems: pattern,
			Singular:     singular,
		}
	}
	resourceBySingular[singular] = r
	return r, nil
}

func foldResourceMethods(from, into *Resource) {
	if from.GetMethod != nil {
		into.GetMethod = from.GetMethod
	}
	if from.ListMethod != nil {
		into.ListMethod = from.ListMethod
	}
	if from.CreateMethod != nil {
		into.CreateMethod = from.CreateMethod
	}
	if from.UpdateMethod != nil {
		into.UpdateMethod = from.UpdateMethod
	}
	if from.DeleteMethod != nil {
		into.DeleteMethod = from.DeleteMethod
	}
}
