package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type Service struct {
	ServiceDefinition
	Headers map[string]string
	Client  *http.Client
}

func NewService(serviceDefinition ServiceDefinition, headers map[string]string) *Service {
	return &Service{
		ServiceDefinition: serviceDefinition,
		Headers:           headers,
		Client:            &http.Client{},
	}
}

func (s *Service) ExecuteCommand(resource string, args []string) (string, error) {
	r, err := s.GetResource(resource)
	if err != nil {
		return "", err
	}
	req, err := r.ExecuteCommand(args)
	if err != nil {
		return "", fmt.Errorf("unable to execute command: %v", err)
	}
	if req == nil {
		return "", nil
	}
	url, err := url.Parse(fmt.Sprintf("%s/%s", s.ServerURL, req.URL.String()))
	if err != nil {
		return "", fmt.Errorf("unable to create url: %v", err)
	}
	req.URL = url
	return s.doRequest(req)
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
