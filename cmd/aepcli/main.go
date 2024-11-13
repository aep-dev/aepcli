package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/aep-dev/aepcli/internal/config"
	"github.com/aep-dev/aepcli/internal/service"

	"github.com/spf13/cobra"
)

func main() {
	err := aepcli(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func aepcli(args []string) error {
	var logLevel string
	var fileAliasOrCore string
	var additionalArgs []string
	var headers []string
	var pathPrefix string
	var serverURL string
	var configFileVar string
	var s *service.Service

	rootCmd := &cobra.Command{
		Use:  "aepcli [host or api alias] [resource or --help]",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fileAliasOrCore = args[0]
			if len(args) > 1 {
				additionalArgs = args[1:]
			}
		},
	}

	configFile, err := config.DefaultConfigFile()
	if err != nil {
		return fmt.Errorf("unable to get default config file: %w", err)
	}

	rootCmd.Flags().SetInterspersed(false) // allow sub parsers to parse subsequent flags after the resource
	rootCmd.PersistentFlags().StringArrayVar(&headers, "header", []string{}, "Specify headers in the format key=value")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the logging level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&pathPrefix, "path-prefix", "", "Specify a path prefix that is prepended to all paths in the openapi schema. This will strip them when evaluating the resource hierarchy paths.")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server-url", "", "Specify a URL to use for the server. If not specified, the first server URL in the OpenAPI definition will be used.")
	rootCmd.PersistentFlags().StringVar(&configFileVar, "config", "", "Path to config file")
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		return err
	}

	if configFileVar != "" {
		configFile = configFileVar
	}

	if err := setLogLevel(logLevel); err != nil {
		return fmt.Errorf("unable to set log level: %w", err)
	}

	c, err := config.ReadConfigFromFile(configFile)
	if err != nil {
		return fmt.Errorf("unable to read config: %v", err)
	}

	if fileAliasOrCore == "core" {
		return handleCoreCommand(additionalArgs, configFile)
	}

	if api, ok := c.APIs[fileAliasOrCore]; ok {
		cd, err := config.ConfigDir()
		if err != nil {
			return fmt.Errorf("unable to get config directory: %w", err)
		}
		if filepath.IsAbs(api.OpenAPIPath) || strings.HasPrefix(api.OpenAPIPath, "http") {
			fileAliasOrCore = api.OpenAPIPath
		} else {
			fileAliasOrCore = filepath.Join(cd, api.OpenAPIPath)
		}
		if pathPrefix == "" {
			pathPrefix = api.PathPrefix
		}
		headers = append(headers, api.Headers...)
		serverURL = api.ServerURL
	}

	oas, err := openapi.FetchOpenAPI(fileAliasOrCore)
	if err != nil {
		return fmt.Errorf("unable to fetch openapi: %w", err)
	}
	api, err := api.GetAPI(oas, serverURL, pathPrefix)
	if err != nil {
		return fmt.Errorf("unable to get api: %w", err)
	}
	headersMap, err := parseHeaders(headers)
	if err != nil {
		return fmt.Errorf("unable to parse headers: %w", err)
	}

	s = service.NewService(api, headersMap)

	result, err := s.ExecuteCommand(additionalArgs)
	fmt.Println(result)
	if err != nil {
		return err
	}
	return nil
}

func parseHeaders(headers []string) (map[string]string, error) {
	parsedHeaders := map[string]string{}
	for _, header := range headers {
		parts := strings.SplitN(header, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format: %s", header)
		}
		parsedHeaders[parts[0]] = parts[1]
	}
	return parsedHeaders, nil
}

func setLogLevel(levelAsString string) error {
	level := slog.LevelInfo
	switch levelAsString {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return fmt.Errorf("invalid log level: %v", levelAsString)
	}
	slog.SetLogLoggerLevel(level)
	return nil
}

func handleCoreCommand(additionalArgs []string, configFile string) error {
	var openAPIPath string
	var overwrite bool
	var api config.API
	var serverURL string
	var headers []string
	var pathPrefix string

	coreCmd := &cobra.Command{
		Use:   "core",
		Short: "Core API management commands",
	}

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
	coreCmd.AddCommand(configCmd)

	coreCmd.SetArgs(additionalArgs)
	if err := coreCmd.Execute(); err != nil {
		return fmt.Errorf("error executing core command: %v", err)
	}
	return nil
}
