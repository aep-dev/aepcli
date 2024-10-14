package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func (s *Service) doRequest(r *http.Request) (string, error) {
	r.Header.Set("Content-Type", "application/json")
	for k, v := range s.Headers {
		r.Header.Set(k, v)
	}
	resp, err := s.Client.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %w", err)
	}
	return prettyJSON.String(), nil
}

func (s *Service) ListResource(resource string) (string, error) {
	r, err := s.ServiceDefinition.GetResource(resource)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/%s", s.ServerURL, r.Plural)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}
	return s.doRequest(req)
}

func (s *Service) GetResource(resource, id string) (string, error) {
	r, err := s.ServiceDefinition.GetResource(resource)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/%s/%s", s.ServerURL, r.Plural, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}
	return s.doRequest(req)
}

func (s *Service) CreateResource(resource, id string) (string, error) {
	r, err := s.ServiceDefinition.GetResource(resource)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/%s?id=%s", s.ServerURL, r.Plural, id)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}
	return s.doRequest(req)
}

func (s *Service) DeleteResource(resource, id string) (string, error) {
	r, err := s.ServiceDefinition.GetResource(resource)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/%s/%s", s.ServerURL, r.Plural, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}
	return s.doRequest(req)
}
