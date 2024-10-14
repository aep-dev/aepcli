package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aep-dev/aepcli/internal/openapi"
	"github.com/aep-dev/aepcli/internal/service"

	"github.com/spf13/cobra"
)

func main() {
	var openapiFile string
	var resource string
	var additionalArgs []string
	var s *service.Service

	rootCmd := &cobra.Command{
		Use:  "aepcli",
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			resource = args[0]
			additionalArgs = args[1:]
		},
	}

	var rawHeaders []string
	rootCmd.Flags().SetInterspersed(false) // allow sub parsers to parse subsequent flags after the resource
	rootCmd.PersistentFlags().StringArrayVar(&rawHeaders, "header", []string{}, "Specify headers in the format key=value")
	rootCmd.PersistentFlags().StringVar(&openapiFile, "openapi-file", "", "Specify the path to the openapi file to configure aepcli. Can be a local file path, or a URL")
	rootCmd.MarkPersistentFlagRequired("host")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	openapi, err := openapi.FetchOpenAPI(openapiFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	serviceDefinition, err := service.GetServiceDefinition(openapi)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	headers, err := parseHeaders(rawHeaders)
	if err != nil {
		fmt.Println(fmt.Errorf("unable to parse headers: %w", err))
		os.Exit(1)
	}

	s = service.NewService(*serviceDefinition, headers)

	result, err := s.ExecuteCommand(resource, additionalArgs)
	if err != nil {
		fmt.Println("an error occurred: %v", err)
		os.Exit(1)
	}
	fmt.Println(result)
	os.Exit(0)
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
