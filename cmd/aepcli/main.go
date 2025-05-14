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

const (
	CODE_OK                  = 0
	CODE_ERR                 = 1
	CODE_HTTP_ERROR_RESPONSE = 2
)

func main() {
	code, err := aepcli(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(code)
}

func aepcli(args []string) (int, error) {
	var dryRun bool
	var logHTTP bool
	var logLevel string
	var fileAliasOrCore string
	var additionalArgs []string
	var headers []string
	var pathPrefix string
	var serverURL string
	var configFileVar string
	var s *service.ServiceCommand

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
		return CODE_OK, fmt.Errorf("unable to get default config file: %w", err)
	}

	rootCmd.Flags().SetInterspersed(false) // allow sub parsers to parse subsequent flags after the resource
	rootCmd.PersistentFlags().StringArrayVar(&headers, "header", []string{}, "Specify headers in the format key=value")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the logging level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolVar(&logHTTP, "log-http", false, "Set to true to log HTTP requests. This can be helpful when attempting to write your own code or debug.")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Set to true to not make any changes. This can be helpful when paired with log-http to just view http requests instead of perform them.")
	rootCmd.PersistentFlags().StringVar(&pathPrefix, "path-prefix", "", "Specify a path prefix that is prepended to all paths in the openapi schema. This will strip them when evaluating the resource hierarchy paths.")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server-url", "", "Specify a URL to use for the server. If not specified, the first server URL in the OpenAPI definition will be used.")
	rootCmd.PersistentFlags().StringVar(&configFileVar, "config", "", "Path to config file")
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		return CODE_OK, err
	}

	if configFileVar != "" {
		configFile = configFileVar
	}

	if err := setLogLevel(logLevel); err != nil {
		return CODE_ERR, fmt.Errorf("unable to set log level: %w", err)
	}

	c, err := config.ReadConfigFromFile(configFile)
	if err != nil {
		return CODE_ERR, fmt.Errorf("unable to read config: %v", err)
	}

	if fileAliasOrCore == "core" {
		return CODE_OK, handleCoreCommand(additionalArgs, configFile)
	}

	if api, ok := c.APIs[fileAliasOrCore]; ok {
		cd, err := config.ConfigDir()
		if err != nil {
			return CODE_ERR, fmt.Errorf("unable to get config directory: %w", err)
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
		return CODE_ERR, fmt.Errorf("unable to fetch openapi: %w", err)
	}
	api, err := api.GetAPI(oas, serverURL, pathPrefix)
	if err != nil {
		return CODE_ERR, fmt.Errorf("unable to get api: %w", err)
	}
	headersMap, err := parseHeaders(headers)
	if err != nil {
		return CODE_ERR, fmt.Errorf("unable to parse headers: %w", err)
	}

	s = service.NewServiceCommand(api, headersMap, dryRun, logHTTP)

	result, err := s.Execute(additionalArgs)
	returnCode := CODE_OK
	output := ""
	if result != nil {
		output = result.Output
		if result.StatusCode != 0 && result.StatusCode/100 != 2 {
			returnCode = CODE_HTTP_ERROR_RESPONSE
		}
	}
	fmt.Println(output)
	if err != nil {
		return CODE_ERR, err
	}
	return returnCode, nil
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
