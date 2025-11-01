package main

import (
	"fmt"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

func main() {
	a := &api.API{
		Name:      "test",
		ServerURL: "https://api.example.com",
		Resources: map[string]*api.Resource{
			"project": {
				Singular: "project",
				Plural:   "projects",
				Parents:  []string{},
				Schema:   &openapi.Schema{},
			},
			"dataset": {
				Singular: "dataset",
				Plural:   "datasets",
				Parents:  []string{"project"},
				Schema:   &openapi.Schema{},
			},
		},
	}

	err := api.AddImplicitFieldsAndValidate(a)
	if err != nil {
		panic(err)
	}

	dataset := a.Resources["dataset"]
	patternElems := dataset.PatternElems()
	fmt.Printf("Dataset PatternElems: %+v\n", patternElems)

	// Simulate the flag creation logic
	i := 1
	for i < len(patternElems)-1 {
		p := patternElems[i]
		flagName := p[1 : len(p)-1]
		fmt.Printf("Would create flag: --%s\n", flagName)
		i += 2
	}
}
