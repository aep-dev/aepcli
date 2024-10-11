package main

import (
	"fmt"
	"os"

	"github.com/aep-dev/aepcli/internal/service"

	"github.com/spf13/cobra"
)

func main() {
	var host string
	var resource string
	var additionalArgs []string
	var s *service.Service

	rootCmd := &cobra.Command{
		Use: "aepcli",
		Run: func(cmd *cobra.Command, args []string) {
			// fmt.Printf("args: %q\n", args)
			resource = args[0]
			additionalArgs = args[1:]
		},
	}

	rootCmd.PersistentFlags().StringVar(&host, "host", "", "Specify the host address of the service")
	rootCmd.MarkPersistentFlagRequired("host")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	openapi, err := service.FetchOpenAPI(fmt.Sprintf("%s/openapi.json", host))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	serviceDefinition, err := service.GetServiceDefinition(openapi)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s = service.NewService(host, serviceDefinition)

	// fmt.Printf("additionalArgs: %q\n", additionalArgs)
	resourceCmd := &cobra.Command{Use: "aepcli-resource"}
	resourceCmd.SetArgs(additionalArgs)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Get a resource",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(s.ListResource(resource))
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a resource",
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			fmt.Print(s.GetResource(resource, id))
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource",
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			fmt.Print(s.CreateResource(resource, id))
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
			fmt.Print(s.DeleteResource(resource, id))
		},
	}

	resourceCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, deleteCmd)

	if err := resourceCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
