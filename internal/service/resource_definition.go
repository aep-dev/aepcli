package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aep-dev/aepcli/internal/openapi"
	"github.com/spf13/cobra"
)

type Resource struct {
	Singular string
	Plural   string
	Parents  []*Resource
	Pattern  []string // TOO(yft): support multiple patterns
	Schema   openapi.Schema
}

func (r *Resource) ExecuteCommand(args []string) (*http.Request, error) {
	c := cobra.Command{Use: r.Plural}
	var err error
	var req *http.Request
	var parents []*string

	i := 1
	for i < len(r.Pattern)-1 {
		p := r.Pattern[i]
		flagName := p[1 : len(p)-1]
		var flagValue string
		parents = append(parents, &flagValue)
		c.PersistentFlags().StringVar(&flagValue, flagName, "", fmt.Sprintf("The %v of the resource", flagName))
		i += 2
	}

	withPrefix := func(path string) string {
		pElems := []string{}
		for i, p := range r.Pattern {
			// last element, we assume this was handled by the caller.
			if i == len(r.Pattern)-1 {
				continue
			}
			if i%2 == 0 {
				pElems = append(pElems, p)
			} else {
				pElems = append(pElems, *parents[i/2])
			}
		}
		prefix := strings.Join(pElems, "/")
		return fmt.Sprintf("%s%s", prefix, path)
	}

	createArgs := map[string]*string{}
	createCmd := &cobra.Command{
		Use:   "create",
		Short: fmt.Sprintf("Create a %v", strings.ToLower(r.Singular)),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			p := withPrefix(fmt.Sprintf("?id=%s", id))
			jsonBody, err := generateJsonPayload(createArgs)
			if err != nil {
				slog.Error(fmt.Sprintf("unable to create json body for update: %v", err))
			}
			req, err = http.NewRequest("POST", p, strings.NewReader(string(jsonBody)))
			if err != nil {
				slog.Error(fmt.Sprintf("error creating post request: %v", err))
			}
		},
	}
	addSchemaFlags(createCmd, r.Schema, createArgs)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: fmt.Sprintf("Get a %v", strings.ToLower(r.Singular)),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			p := withPrefix(fmt.Sprintf("/%s", id))
			req, err = http.NewRequest("GET", p, nil)
		},
	}

	updateArgs := map[string]*string{}
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: fmt.Sprintf("Update a %v", strings.ToLower(r.Singular)),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			p := withPrefix(fmt.Sprintf("/%s", id))
			jsonBody, err := generateJsonPayload(updateArgs)
			if err != nil {
				slog.Error(fmt.Sprintf("unable to create json body for update: %v", err))
			}
			req, err = http.NewRequest("PATCH", p, strings.NewReader(string(jsonBody)))
			if err != nil {
				slog.Error(fmt.Sprintf("error creating patch request: %v", err))
			}
		},
	}
	addSchemaFlags(updateCmd, r.Schema, updateArgs)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: fmt.Sprintf("Delete a %v", strings.ToLower(r.Singular)),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			p := withPrefix(fmt.Sprintf("/%s", id))
			req, err = http.NewRequest("DELETE", p, nil)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %v", strings.ToLower(r.Plural)),
		Run: func(cmd *cobra.Command, args []string) {
			p := withPrefix("")
			req, err = http.NewRequest("GET", p, nil)
		},
	}

	c.AddCommand(createCmd, getCmd, updateCmd, deleteCmd, listCmd)
	c.SetArgs(args)
	if err := c.Execute(); err != nil {
		return nil, err
	}
	return req, err
}

// TODO(yft): add support for more types
func addSchemaFlags(c *cobra.Command, schema openapi.Schema, args map[string]*string) error {
	for name, prop := range schema.Properties {
		if prop.ReadOnly {
			continue
		}
		// slog.Debug(fmt.Sprintf("Adding flag %v", name))
		var value string
		args[name] = &value
		switch prop.Type {
		case "string":
			c.Flags().StringVar(&value, name, "", fmt.Sprintf("The %v of the resource", name))
		// case "integer":
		// c.Flags().IntVar(&value, name, 0, fmt.Sprintf("The %v of the resource", name))
		// case "boolean":
		// c.Flags().BoolVar(&value, name, false, fmt.Sprintf("The %v of the resource", name))
		default:
			fmt.Printf("Unsupported type: %v\n", prop.Type)
		}
	}
	for _, f := range schema.Required {
		c.MarkFlagRequired(f)
	}
	return nil
}

func generateJsonPayload(args map[string]*string) (string, error) {
	body := map[string]interface{}{}
	for key, value := range args {
		body[key] = *value
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %v", err)
	}
	return string(jsonBody), nil
}
