package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/spf13/cobra"
)

func ExecuteResourceCommand(r *api.Resource, args []string) (*http.Request, string, error) {
	c := cobra.Command{Use: r.Singular}
	var err error
	var req *http.Request
	var parents []*string

	i := 1
	patternElems := r.PatternElems()
	for i < len(patternElems)-1 {
		p := patternElems[i]
		flagName := p[1 : len(p)-1] // extract content between braces
		// Strip "_id" suffix if present to make flag names cleaner
		if strings.HasSuffix(flagName, "_id") {
			flagName = strings.TrimSuffix(flagName, "_id")
		}
		var flagValue string
		parents = append(parents, &flagValue)
		c.PersistentFlags().StringVar(&flagValue, flagName, "", fmt.Sprintf("The %v of the resource", flagName))
		c.MarkPersistentFlagRequired(flagName)
		i += 2
	}

	withPrefix := func(path string) string {
		pElems := []string{}
		for i, p := range patternElems {
			// last element, we assume this was handled by the caller.
			if i == len(patternElems)-1 {
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

	if r.Methods.Create != nil {
		use := "create [id]"
		args := cobra.ExactArgs(1)
		if !r.Methods.Create.SupportsUserSettableCreate {
			use = "create"
			args = cobra.ExactArgs(0)
		}
		createArgs := map[string]interface{}{}
		var dataContent map[string]interface{}
		createArgs["data"] = &dataContent

		createCmd := &cobra.Command{
			Use:   use,
			Short: fmt.Sprintf("Create a %v", strings.ToLower(r.Singular)),
			Args:  args,
			Run: func(cmd *cobra.Command, args []string) {
				p := withPrefix("")
				if r.Methods.Create.SupportsUserSettableCreate {
					id := args[0]
					p = withPrefix(fmt.Sprintf("?id=%s", url.QueryEscape(id)))
				}
				jsonBody, err := generateJsonPayload(cmd, createArgs)
				if err != nil {
					slog.Error(fmt.Sprintf("unable to create json body for create: %v", err))
				}
				req, err = http.NewRequest("POST", p, strings.NewReader(string(jsonBody)))
				if err != nil {
					slog.Error(fmt.Sprintf("error creating post request: %v", err))
				}
			},
		}

		createCmd.Flags().Var(&DataFlag{&dataContent}, "@data", "Read resource data from JSON file")

		addSchemaFlags(createCmd, *r.Schema, createArgs)
		c.AddCommand(createCmd)
	}

	if r.Methods.Get != nil {
		getCmd := &cobra.Command{
			Use:   "get [id]",
			Short: fmt.Sprintf("Get a %v", strings.ToLower(r.Singular)),
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				id := args[0]
				p := withPrefix(fmt.Sprintf("/%s", id))
				req, err = http.NewRequest("GET", p, nil)
			},
		}
		c.AddCommand(getCmd)
	}

	if r.Methods.Update != nil {

		updateArgs := map[string]interface{}{}
		var updateDataContent map[string]interface{}
		updateArgs["data"] = &updateDataContent

		updateCmd := &cobra.Command{
			Use:   "update [id]",
			Short: fmt.Sprintf("Update a %v", strings.ToLower(r.Singular)),
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				id := args[0]
				p := withPrefix(fmt.Sprintf("/%s", id))
				jsonBody, err := generateJsonPayload(cmd, updateArgs)
				if err != nil {
					slog.Error(fmt.Sprintf("unable to create json body for update: %v", err))
				}
				req, err = http.NewRequest("PATCH", p, strings.NewReader(string(jsonBody)))
				if err != nil {
					slog.Error(fmt.Sprintf("error creating patch request: %v", err))
				}
			},
		}

		updateCmd.Flags().Var(&DataFlag{&updateDataContent}, "@data", "Read resource data from JSON file")

		addSchemaFlags(updateCmd, *r.Schema, updateArgs)
		c.AddCommand(updateCmd)
	}

	if r.Methods.Delete != nil {

		deleteCmd := &cobra.Command{
			Use:   "delete [id]",
			Short: fmt.Sprintf("Delete a %v", strings.ToLower(r.Singular)),
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				id := args[0]
				p := withPrefix(fmt.Sprintf("/%s", id))
				req, err = http.NewRequest("DELETE", p, nil)
			},
		}
		c.AddCommand(deleteCmd)
	}

	if r.Methods.List != nil {

		listCmd := &cobra.Command{
			Use:   "list",
			Short: fmt.Sprintf("List %v", strings.ToLower(r.Singular)),
			Run: func(cmd *cobra.Command, args []string) {
				p := withPrefix("")
				req, err = http.NewRequest("GET", p, nil)
			},
		}
		c.AddCommand(listCmd)
	}
	for _, cm := range r.CustomMethods {
		customArgs := map[string]interface{}{}
		var customDataContent map[string]interface{}

		customCmd := &cobra.Command{
			Use:   fmt.Sprintf(":%s [id]", cm.Name),
			Short: fmt.Sprintf("%v a %v", cm.Method, strings.ToLower(r.Singular)),
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				id := args[0]
				p := withPrefix(fmt.Sprintf("/%s:%s", id, cm.Name))
				if cm.Method == "POST" {
					jsonBody, inner_err := generateJsonPayload(cmd, customArgs)
					if inner_err != nil {
						slog.Error(fmt.Sprintf("unable to create json body for custom method: %v", inner_err))
					}
					req, err = http.NewRequest(cm.Method, p, strings.NewReader(string(jsonBody)))
				} else {
					req, err = http.NewRequest(cm.Method, p, nil)
				}
			},
		}

		if cm.Method == "POST" {
			customArgs["data"] = &customDataContent
			customCmd.Flags().Var(&DataFlag{&customDataContent}, "@data", "Read resource data from JSON file")
			addSchemaFlags(customCmd, *cm.Request, customArgs)
		}
		c.AddCommand(customCmd)
	}
	var stdout strings.Builder
	c.SetOut(&stdout)
	c.SetArgs(args)
	if err := c.Execute(); err != nil {
		return nil, stdout.String(), err
	}
	return req, stdout.String(), err
}

func addSchemaFlags(c *cobra.Command, schema openapi.Schema, args map[string]interface{}) error {
	for name, prop := range schema.Properties {
		description := prop.Description
		if description == "" {
			description = fmt.Sprintf("The %v of the resource", name)
		}

		// Check if the field is required
		isRequired := false
		for _, required := range schema.Required {
			if required == name {
				isRequired = true
				break
			}
		}
		if isRequired {
			description = fmt.Sprintf("%s (required)", description)
		}

		if prop.ReadOnly {
			continue
		}
		switch prop.Type {
		case "string":
			var value string
			args[name] = &value
			c.Flags().StringVar(&value, name, "", description)
		case "integer":
			var value int
			args[name] = &value
			c.Flags().IntVar(&value, name, 0, description)
		case "number":
			var value float64
			args[name] = &value
			c.Flags().Float64Var(&value, name, 0, description)
		case "boolean":
			var value bool
			args[name] = &value
			c.Flags().BoolVar(&value, name, false, description)
		case "array":
			if prop.Items == nil {
				return fmt.Errorf("items is required for array type, not found for field %v", name)
			}
			var value []interface{}
			args[name] = &value
			c.Flags().Var(&ArrayFlag{&value, prop.Items.Type}, name, description)
		case "object":
			var parsedValue map[string]interface{}
			args[name] = &parsedValue
			c.Flags().Var(&JSONFlag{&parsedValue}, name, description)
		default:
			var parsedValue map[string]interface{}
			args[name] = &parsedValue
			c.Flags().Var(&JSONFlag{&parsedValue}, name, description)
		}
	}
	for _, f := range schema.Required {
		c.MarkFlagRequired(f)
	}
	return nil
}

func generateJsonPayload(c *cobra.Command, args map[string]interface{}) (string, error) {
	// Check if --@data flag was used
	dataFlag := c.Flags().Lookup("@data")
	if dataFlag != nil && dataFlag.Changed {
		// Check for conflicts with other flags
		var conflictingFlags []string
		for key := range args {
			if key == "data" {
				continue // Skip the internal data key
			}
			if flag := c.Flags().Lookup(key); flag != nil && flag.Changed {
				conflictingFlags = append(conflictingFlags, "--"+key)
			}
		}

		if len(conflictingFlags) > 0 {
			return "", fmt.Errorf("--@data flag cannot be used with individual field flags (%s)", strings.Join(conflictingFlags, ", "))
		}

		// Get the data from the --@data flag
		if dataValue, ok := args["data"]; ok {
			if dataMap, ok := dataValue.(*map[string]interface{}); ok && *dataMap != nil {
				jsonBody, err := json.Marshal(*dataMap)
				if err != nil {
					return "", fmt.Errorf("error marshalling JSON from --@data: %v", err)
				}
				return string(jsonBody), nil
			}
		}

		// If --@data flag was used but no data was set, return empty object
		return "{}", nil
	}

	// Original logic for individual flags
	body := map[string]interface{}{}
	for key, value := range args {
		if key == "data" {
			continue // Skip the data field when building from individual flags
		}
		if c.Flags().Lookup(key).Changed {
			body[key] = value
		}
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %v", err)
	}
	return string(jsonBody), nil
}
