package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/api"
)

type ServiceCommand struct {
	API        api.API
	Headers    map[string]string
	DryRun     bool
	LogHTTP    bool
	Insecure   bool
	CACertPath string
	Client     *http.Client
}

func NewServiceCommand(api *api.API, headers map[string]string, dryRun bool, logHTTP bool, insecure bool, caCertPath string) (*ServiceCommand, error) {
	client := &http.Client{}

	if insecure {
		slog.Debug("Loading TLS configuration")
		slog.Debug("TLS certificate verification disabled (insecure mode)")
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	} else if caCertPath != "" {
		slog.Debug("Loading TLS configuration")
		// Load custom CA certificate
		caCertPool, err := loadCACertificate(caCertPath)
		if err != nil {
			return nil, err
		}
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: caCertPool},
		}
		client.Transport = tr
	} else {
		// Use default system CA certificates
		slog.Debug("Using system CA certificates")
	}

	return &ServiceCommand{
		API:        *api,
		Headers:    headers,
		DryRun:     dryRun,
		LogHTTP:    logHTTP,
		Insecure:   insecure,
		CACertPath: caCertPath,
		Client:     client,
	}, nil
}

func (s *ServiceCommand) Execute(args []string) (*Result, error) {
	if len(args) == 0 || args[0] == "--help" {
		return &Result{s.PrintHelp(), 0}, nil
	}
	resource := args[0]
	r, err := s.API.GetResource(resource)
	if err != nil {
		return nil, fmt.Errorf("%v\n%v", err, s.PrintHelp())
	}
	req, output, err := ExecuteResourceCommand(r, args[1:])
	if err != nil {
		return &Result{output, 0}, err
	}
	if req == nil {
		return &Result{output, 0}, nil
	}
	url, err := url.Parse(fmt.Sprintf("%s/%s", s.API.ServerURL, req.URL.String()))
	if err != nil {
		return nil, fmt.Errorf("unable to create url: %v", err)
	}
	req.URL = url
	resp, err := s.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("unable to execute request: %v", err)
	}
	if output != "" {
		resp.Output = output + "\n" + resp.Output
	}
	return resp, nil
}

func (s *ServiceCommand) doRequest(r *http.Request) (*Result, error) {
	contentType := "application/json"
	if r.Method == http.MethodPatch {
		contentType = "application/merge-patch+json"
	}
	r.Header.Set("Content-Type", contentType)
	for k, v := range s.Headers {
		r.Header.Set(k, v)
	}
	body := ""
	if r.Body != nil {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to read request body: %v", err)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(b))
		body = string(b)
	}
	requestLog := fmt.Sprintf("Request: %s %s\n%s", r.Method, r.URL.String(), string(body))
	slog.Debug(requestLog)
	if s.LogHTTP {
		fmt.Println(requestLog)
	}
	if s.DryRun {
		slog.Debug("Dry run: not making request")
		return nil, nil
	}
	resp, err := s.Client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("unable to execute request: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %v", err)
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, respBody, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to format JSON: %w", err)
	}
	return &Result{prettyJSON.String(), resp.StatusCode}, nil
}

func (s *ServiceCommand) PrintHelp() string {
	var resources []string
	for singular := range s.API.Resources {
		resources = append(resources, singular)
	}
	sort.Strings(resources)

	var output strings.Builder
	output.WriteString("Usage: [resource] [method] [flags]\n\n")
	output.WriteString("Command group for " + s.API.ServerURL + "\n\n")
	output.WriteString("Available resources:\n")
	for _, r := range resources {
		output.WriteString(fmt.Sprintf("  - %s\n", r))
	}
	return output.String()
}
