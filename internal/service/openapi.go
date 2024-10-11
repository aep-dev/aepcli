package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type OpenAPI struct {
	Openapi    string              `json:"openapi"`
	Info       Info                `json:"info"`
	Paths      map[string]PathItem `json:"paths"`
	Components Components          `json:"components"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

type Operation struct {
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	OperationID string              `json:"operationId"`
	Parameters  []Parameter         `json:"parameters"`
	Responses   map[string]Response `json:"responses"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content"`
}

type RequestBody struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Schema struct {
	Type         string            `json:"type"`
	Format       string            `json:"format,omitempty"`
	Items        *Schema           `json:"items,omitempty"`
	Properties   map[string]Schema `json:"properties,omitempty"`
	Ref          string            `json:"$ref,omitempty"`
	XAEPResource *XAEPResource     `json:"x-aep-resource,omitempty"`
}

type Components struct {
	Schemas map[string]Schema `json:"schemas"`
}

type XAEPResource struct {
	Singular string   `json:"singular,omitempty"`
	Plural   string   `json:"plural,omitempty"`
	Patterns []string `json:"patterns,omitempty"`
}

func FetchOpenAPI(url string) (*OpenAPI, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var api OpenAPI
	if err := json.Unmarshal(body, &api); err != nil {
		return nil, err
	}

	return &api, nil
}
