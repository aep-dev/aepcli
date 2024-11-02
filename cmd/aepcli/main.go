package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aep-dev/aepcli/internal/config"
	"github.com/aep-dev/aepcli/internal/openapi"
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
	var fileOrAlias string
	var additionalArgs []string
	var s *service.Service

	rootCmd := &cobra.Command{
		Use:  "aepcli [host or api alias] [resource or --help]",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fileOrAlias = args[0]
			if len(args) > 1 {
				additionalArgs = args[1:]
			}
		},
	}

	var rawHeaders []string
	var pathPrefix string
	var serverURL string
	rootCmd.Flags().SetInterspersed(false) // allow sub parsers to parse subsequent flags after the resource
	rootCmd.PersistentFlags().StringArrayVar(&rawHeaders, "header", []string{}, "Specify headers in the format key=value")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the logging level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&pathPrefix, "path-prefix", "", "Specify a path prefix that is prepended to all paths in the openapi schema. This will strip them when evaluating the resource hierarchy paths.")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server-url", "", "Specify a URL to use for the server. If not specified, the first server URL in the OpenAPI definition will be used.")

	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		return err
	}

	if err := setLogLevel(logLevel); err != nil {
		return fmt.Errorf("unable to set log level: %w", err)
	}

	c, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("unable to read config: %v", err)
	}

	if api, ok := c.APIs[fileOrAlias]; ok {
		cd, err := config.ConfigDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fileOrAlias = filepath.Join(cd, api.OpenAPIPath)
		if pathPrefix == "" {
			pathPrefix = api.PathPrefix
		}
		rawHeaders = append(rawHeaders, api.Headers...)
		serverURL = api.ServerURL
	}

	openapi, err := openapi.FetchOpenAPI(fileOrAlias)
	if err != nil {
		return fmt.Errorf("unable to fetch openapi: %w", err)
	}
	serviceDefinition, err := service.GetServiceDefinition(openapi, serverURL, pathPrefix)
	if err != nil {
		return fmt.Errorf("unable to get service definition: %w", err)
	}

	headers, err := parseHeaders(rawHeaders)
	if err != nil {
		return fmt.Errorf("unable to parse headers: %w", err)
	}

	s = service.NewService(*serviceDefinition, headers)

	result, err := s.ExecuteCommand(additionalArgs)
	if err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}
	fmt.Println(result)
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
