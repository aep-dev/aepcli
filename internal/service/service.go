package service

import (
	"fmt"
	"io"
	"net/http"
)

type Service struct {
	BaseURL           string
	ServiceDefinition *ServiceDefinition
	Client            *http.Client
}

func NewService(baseURL string, serviceDefinition *ServiceDefinition) *Service {
	return &Service{
		BaseURL:           baseURL,
		ServiceDefinition: serviceDefinition,
		Client:            &http.Client{},
	}
}

func (s *Service) ListResource(resource string) (string, error) {
	r, err := s.ServiceDefinition.GetResource(resource)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/%s", s.BaseURL, r.Plural)
	resp, err := s.Client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (s *Service) GetResource(resource, id string) (string, error) {
	r, err := s.ServiceDefinition.GetResource(resource)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/%s/%s", s.BaseURL, r.Plural, id)
	resp, err := s.Client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (s *Service) CreateResource(resource, id string) (string, error) {
	r, err := s.ServiceDefinition.GetResource(resource)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/%s?id=%s", s.BaseURL, r.Plural, id)
	resp, err := s.Client.Post(url, "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (s *Service) DeleteResource(resource, id string) (string, error) {
	r, err := s.ServiceDefinition.GetResource(resource)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/%s/%s", s.BaseURL, r.Plural, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return "", err
	}
	s.Client.Do(req)
	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
