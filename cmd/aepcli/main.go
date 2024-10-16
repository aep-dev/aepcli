package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aep-dev/aepcli/internal/config"
	"github.com/aep-dev/aepcli/internal/openapi"
	"github.com/aep-dev/aepcli/internal/service"

	"github.com/spf13/cobra"
)

func main() {
	var fileOrAlias string
	var resource string
	var additionalArgs []string
	var s *service.Service

	rootCmd := &cobra.Command{
		Use:  "aepcli",
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fileOrAlias = args[0]
			resource = args[1]
			additionalArgs = args[2:]
		},
	}

	var rawHeaders []string
	rootCmd.Flags().SetInterspersed(false) // allow sub parsers to parse subsequent flags after the resource
	rootCmd.PersistentFlags().StringArrayVar(&rawHeaders, "header", []string{}, "Specify headers in the format key=value")
	rootCmd.MarkPersistentFlagRequired("host")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c, err := config.ReadConfig()
	if err != nil {
		fmt.Println(fmt.Errorf("unable to read config: %v", err))
		os.Exit(1)
	}

	if api, ok := c.APIs[fileOrAlias]; ok {
		cd, err := config.ConfigDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fileOrAlias = filepath.Join(cd, api.OpenAPIPath)
		rawHeaders = append(rawHeaders, api.Headers...)
	}

	openapi, err := openapi.FetchOpenAPI(fileOrAlias)
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
