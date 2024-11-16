package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/aep-dev/aepcli/internal/config"
	"github.com/spf13/cobra"
)

func handleCoreCommand(additionalArgs []string, configFile string) error {
	coreCmd := &cobra.Command{
		Use:   "core",
		Args:  cobra.MinimumNArgs(1),
		Short: "Core API management commands",
	}

	coreCmd.AddCommand(openAPICommand())
	coreCmd.AddCommand(configCmd(configFile))

	coreCmd.SetArgs(additionalArgs)
	if err := coreCmd.Execute(); err != nil {
		return fmt.Errorf("error executing core command: %v", err)
	}
	return nil
}

func openAPICommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "openapi",
		Short: "OpenAPI commands",
	}
	var inputPath string
	var outputPath string
	var pathPrefix string

	convertCmd := &cobra.Command{
		Use:   "convert",
		Short: "Best effort conversion of OpenAPI specification to an AEP API",
		Run: func(cmd *cobra.Command, args []string) {
			if inputPath == "" {
				fmt.Println("Input path is required")
				os.Exit(1)
			}
			slog.Debug("Converting OpenAPI spec", "inputPath", inputPath, "outputPath", outputPath, "pathPrefix", pathPrefix)

			originalOAS, err := openapi.FetchOpenAPI(inputPath)
			if err != nil {
				fmt.Printf("Error fetching OpenAPI spec: %v\n", err)
				os.Exit(1)
			}

			api, err := api.GetAPI(originalOAS, "", pathPrefix)
			if err != nil {
				fmt.Printf("Error converting to AEP API: %v\n", err)
				os.Exit(1)
			}

			finalOAS, err := api.ConvertToOpenAPIBytes()
			if err != nil {
				fmt.Printf("Error converting to OpenAPI: %v\n", err)
				os.Exit(1)
			}

			if outputPath == "" {
				fmt.Println(string(finalOAS))
			} else {
				err = os.WriteFile(outputPath, []byte(finalOAS), 0644)
				if err != nil {
					fmt.Printf("Error writing output file: %v\n", err)
					os.Exit(1)
				}
			}

		},
	}

	convertCmd.Flags().StringVarP(&inputPath, "input", "i", "", "Input OpenAPI specification file path")
	convertCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path. If unset, print to stdout")
	convertCmd.Flags().StringVar(&pathPrefix, "path-prefix", "", "Path prefix to strip from paths when evaluating resource hierarchy")
	convertCmd.MarkFlagRequired("input")

	c.AddCommand(convertCmd)
	return c
}

func configCmd(configFile string) *cobra.Command {
	var openAPIPath string
	var overwrite bool
	var api config.API
	var serverURL string
	var headers []string
	var pathPrefix string

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage core API configurations",
	}

	addCmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new core API configuration",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			api = config.API{
				Name:        args[0],
				OpenAPIPath: openAPIPath,
				ServerURL:   serverURL,
				Headers:     headers,
				PathPrefix:  pathPrefix,
			}
			if err := config.WriteAPIWithName(configFile, api, overwrite); err != nil {
				fmt.Printf("Error writing API config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Core API configuration '%s' added successfully\n", args[0])
		},
	}

	addCmd.Flags().StringVar(&openAPIPath, "openapi-path", "", "Path to OpenAPI specification file")
	addCmd.Flags().StringArrayVar(&headers, "header", []string{}, "Headers in format key=value")
	addCmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL")
	addCmd.Flags().StringVar(&pathPrefix, "path-prefix", "", "Path prefix")
	addCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing configuration")

	readCmd := &cobra.Command{
		Use:   "get [name]",
		Short: "Get an API configuration",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.ReadConfigFromFile(configFile)
			if err != nil {
				fmt.Printf("Error reading config file: %v\n", err)
				os.Exit(1)
			}

			api, exists := cfg.APIs[args[0]]
			if !exists {
				fmt.Printf("No API configuration found with name '%s'\n", args[0])
				os.Exit(1)
			}

			fmt.Printf("Name: %s\n", api.Name)
			fmt.Printf("OpenAPI Path: %s\n", api.OpenAPIPath)
			fmt.Printf("Server URL: %s\n", api.ServerURL)
			fmt.Printf("Headers: %v\n", api.Headers)
			fmt.Printf("Path Prefix: %s\n", api.PathPrefix)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all API configurations",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			apis, err := config.ListAPIs(configFile)
			if err != nil {
				fmt.Printf("Error listing APIs: %v\n", err)
				os.Exit(1)
			}

			if len(apis) == 0 {
				fmt.Println("No API configurations found")
				return
			}

			for _, api := range apis {
				fmt.Printf("Name: %s\n", api.Name)
				fmt.Printf("OpenAPI Path: %s\n", api.OpenAPIPath)
				fmt.Printf("Server URL: %s\n", api.ServerURL)
				fmt.Printf("Headers: %v\n", api.Headers)
				fmt.Printf("Path Prefix: %s\n", api.PathPrefix)
				fmt.Println()
			}
		},
	}

	configCmd.AddCommand(addCmd)
	configCmd.AddCommand(readCmd)
	configCmd.AddCommand(listCmd)

	return configCmd
}
