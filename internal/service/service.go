package service

import (
	"bytes"
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

type Service struct {
	API     api.API
	Headers map[string]string
	Client  *http.Client
}

func NewService(api *api.API, headers map[string]string) *Service {
	return &Service{
		API:     *api,
		Headers: headers,
		Client:  &http.Client{},
	}
}

func (s *Service) ExecuteCommand(args []string) (string, error) {
	if len(args) == 0 || args[0] == "--help" {
		return s.PrintHelp(), nil
	}
	resource := args[0]
	r, err := s.API.GetResource(resource)
	if err != nil {
		return "", fmt.Errorf("%v\n%v", err, s.PrintHelp())
	}
	req, output, err := ExecuteResourceCommand(r, args[1:])
	if err != nil {
		return "", fmt.Errorf("unable to execute command: %v", err)
	}
	if req == nil {
		return output, nil
	}
	url, err := url.Parse(fmt.Sprintf("%s/%s", s.API.ServerURL, req.URL.String()))
	if err != nil {
		return "", fmt.Errorf("unable to create url: %v", err)
	}
	req.URL = url
	reqOutput, err := s.doRequest(req)
	if err != nil {
		return "", fmt.Errorf("unable to execute request: %v", err)
	}
	return strings.Join([]string{output, reqOutput}, "\n"), nil
}

func (s *Service) doRequest(r *http.Request) (string, error) {
	r.Header.Set("Content-Type", "application/json")
	for k, v := range s.Headers {
		r.Header.Set(k, v)
	}
	body := ""
	if r.Body != nil {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return "", fmt.Errorf("unable to read request body: %v", err)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(b))
		body = string(b)
	}
	slog.Debug(fmt.Sprintf("Request: %s %s\n%s", r.Method, r.URL.String(), string(body)))
	resp, err := s.Client.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body: %v", err)
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, respBody, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %w", err)
	}
	return prettyJSON.String(), nil
}

func (s *Service) PrintHelp() string {
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
