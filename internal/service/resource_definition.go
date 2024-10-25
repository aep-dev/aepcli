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
	Parents  []*Resource
	Pattern  []string // TOO(yft): support multiple patterns
}

func (r *Resource) ExecuteCommand(args []string) (*http.Request, error) {
	c := cobra.Command{Use: r.Plural}
	var err error
	var req *http.Request
	var parents []*string

	i := 1
	for i < len(r.Pattern) {
		p := r.Pattern[i]
		flagName := p[1 : len(p)-1]
		var flagValue string
		parents = append(parents, &flagValue)
		c.PersistentFlags().StringVar(&flagValue, flagName, "", fmt.Sprintf("The %v of the resource", flagName))
		i += 2
	}

	withPrefix := func(path string) string {
		pElems := []string{}
		for i, p := range r.Pattern {
			// last element, we assume this was handled by the caller.
			if i == len(r.Pattern)-1 {
				continue
			}
			if i%2 == 0 {
				pElems = append(pElems, p)
			} else {
				pElems = append(pElems, *parents[i/2])
			}
		}
		prefix := strings.Join(pElems, "/")
		return fmt.Sprintf("%s/%s", prefix, path)
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: fmt.Sprintf("Create a %v", strings.ToLower(r.Singular)),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			p := withPrefix(fmt.Sprintf("?id=%s", id))
			req, err = http.NewRequest("POST", p, nil)
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: fmt.Sprintf("Get a %v", strings.ToLower(r.Singular)),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			p := withPrefix(id)
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
			p := withPrefix(id)
			req, err = http.NewRequest("DELETE", p, nil)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %v", strings.ToLower(r.Plural)),
		Run: func(cmd *cobra.Command, args []string) {
			p := withPrefix("")
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
