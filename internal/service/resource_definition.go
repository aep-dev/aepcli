package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

type Resource struct {
	Singular string
	Plural   string
	Parent   []*Resource
}

func (r *Resource) ExecuteCommand(args []string) (*http.Request, error) {
	c := cobra.Command{Use: r.Plural}
	var parent string
	var err error
	var req *http.Request

	// TODO(yft): add support for multiple parents
	if len(r.Parent) > 0 {
		s := strings.ToLower(r.Parent[0].Singular)
		c.PersistentFlags().StringVar(
			&parent, s, "", fmt.Sprintf("The %v of the resource", s),
		)
	}

	withPrefix := func(path string) string {
		// TODO(yft): add support for multiple parents
		if len(r.Parent) > 0 {
			return fmt.Sprintf("%s/%s/%s", r.Parent[0].Plural, parent, path)
		}
		return path
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: fmt.Sprintf("Create a %v", strings.ToLower(r.Singular)),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			p := withPrefix(fmt.Sprintf("%s?id=%s", r.Plural, id))
			req, err = http.NewRequest("POST", p, nil)
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: fmt.Sprintf("Get a %v", strings.ToLower(r.Singular)),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			p := withPrefix(fmt.Sprintf("%s/%s", r.Plural, id))
			req, err = http.NewRequest("GET", p, nil)
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: fmt.Sprintf("Update a %v", strings.ToLower(r.Singular)),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: fmt.Sprintf("Delete a %v", strings.ToLower(r.Singular)),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			p := withPrefix(fmt.Sprintf("%s/%s", r.Plural, id))
			req, err = http.NewRequest("DELETE", p, nil)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %v", strings.ToLower(r.Plural)),
		Run: func(cmd *cobra.Command, args []string) {
			p := withPrefix(r.Plural)
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
