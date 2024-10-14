package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aep-dev/aepcli/internal/service"

	"github.com/spf13/cobra"
)

func main() {
	var openapiFile string
	var resource string
	var additionalArgs []string
	var s *service.Service

	rootCmd := &cobra.Command{
		Use: "aepcli",
		Run: func(cmd *cobra.Command, args []string) {
			resource = args[0]
			additionalArgs = args[1:]
		},
	}

	var rawHeaders []string
	rootCmd.PersistentFlags().StringArrayVar(&rawHeaders, "header", []string{}, "Specify headers in the format key=value")
	rootCmd.PersistentFlags().StringVar(&openapiFile, "openapi-file", "", "Specify the path to the openapi file to configure aepcli. Can be a local file path, or a URL")
	rootCmd.MarkPersistentFlagRequired("host")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	openapi, err := service.FetchOpenAPI(openapiFile)
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

	resourceCmd := &cobra.Command{Use: "aepcli-resource"}
	resourceCmd.SetArgs(additionalArgs)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Get a resource",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(s.ListResource(resource))
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a resource",
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			resp, err := s.GetResource(resource, id)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(resp)
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource",
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			fmt.Println(s.CreateResource(resource, id))
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a resource",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Update command executed")
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a resource",
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			fmt.Println(s.DeleteResource(resource, id))
		},
	}

	resourceCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, deleteCmd)

	if err := resourceCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
