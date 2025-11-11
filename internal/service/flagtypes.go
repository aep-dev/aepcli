package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
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

// DataFlag handles file references with @file syntax
type DataFlag struct {
	Target *map[string]interface{}
}

func (f *DataFlag) String() string {
	if f.Target == nil || *f.Target == nil {
		return ""
	}
	b, err := json.Marshal(*f.Target)
	if err != nil {
		return "failed to marshal object"
	}
	return string(b)
}

func (f *DataFlag) Set(v string) error {
	// The filename is provided directly (no @ prefix needed)
	filename := v
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("unable to read file '%s': no such file or directory", filename)
		}
		return fmt.Errorf("unable to read file '%s': %v", filename, err)
	}

	// Parse JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		// Try to provide line/column information if possible
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			// Calculate line and column from offset
			line := 1
			col := 1
			for i := int64(0); i < syntaxErr.Offset; i++ {
				if i < int64(len(data)) && data[i] == '\n' {
					line++
					col = 1
				} else {
					col++
				}
			}
			return fmt.Errorf("invalid JSON in '%s': %s at line %d, column %d", filename, syntaxErr.Error(), line, col)
		}
		return fmt.Errorf("invalid JSON in '%s': %v", filename, err)
	}

	*f.Target = jsonData
	return nil
}

func (f *DataFlag) Type() string {
	return "data"
}
