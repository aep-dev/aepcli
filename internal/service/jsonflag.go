package service

import (
	"encoding/json"
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
