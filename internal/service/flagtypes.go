package service

import (
	"encoding/csv"
	"encoding/json"
	"strings"
)

type JSONFlag struct {
	Target interface{}
}

func (f *JSONFlag) String() string {
	b, err := json.Marshal(f.Target)
	if err != nil {
		return "failed to marshal object"
	}
	return string(b)
}

func (f *JSONFlag) Set(v string) error {
	return json.Unmarshal([]byte(v), f.Target)
}

func (f *JSONFlag) Type() string {
	return "json"
}

type ArrayFlag struct {
	Target   *[]interface{}
	ItemType string
}

func (f *ArrayFlag) String() string {
	b, err := json.Marshal(f.Target)
	if err != nil {
		return "failed to marshal object"
	}
	return string(b)
}

func (f *ArrayFlag) Set(v string) error {
	r := csv.NewReader(strings.NewReader(v))
	record, err := r.Read()
	if err != nil {
		return err
	}
	result := []interface{}{}
	for _, t := range record {
		result = append(result, t)
	}
	*f.Target = result
	return nil
}

func (f *ArrayFlag) Type() string {
	return "array"
}
