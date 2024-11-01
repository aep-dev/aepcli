package service

import (
	"strings"
	"testing"
)

func TestService_ExecuteCommand_ListResources(t *testing.T) {
	// Test setup
	svc := NewService(ServiceDefinition{
		ServerURL: "http://test.com",
		Resources: map[string]*Resource{
			"user":    {},
			"post":    {},
			"comment": {},
		},
	}, nil)

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "no arguments",
			args:     []string{},
			expected: "Available resources:\n  - comment\n  - post\n  - user\n",
		},
		{
			name:     "help flag",
			args:     []string{"--help"},
			expected: "Available resources:\n  - comment\n  - post\n  - user\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.ExecuteCommand(tt.args)
			if err != nil {
				t.Errorf("ExecuteCommand() error = %v, expected no error", err)
			}
			if !strings.Contains(result, tt.expected) {
				t.Errorf("ExecuteCommand() = %v, expected it to contain %v", result, tt.expected)
			}
		})
	}
}
