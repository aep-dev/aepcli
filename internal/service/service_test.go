package service

import (
	"strings"
	"testing"
)

func TestService_ExecuteCommand_ListResources(t *testing.T) {
	// Test setup
	svc := NewServiceCommand(getTestAPI(), nil, false, false)

	tests := []struct {
		name          string
		args          []string
		expectAsError bool
		expected      string
	}{
		{
			name:     "no arguments",
			args:     []string{},
			expected: "Available resources:\n  - comment\n  - dataset\n  - project\n  - user\n",
		},
		{
			name:     "help flag",
			args:     []string{"--help"},
			expected: "Available resources:\n  - comment\n  - dataset\n  - project\n  - user\n",
		},
		{
			name:          "unknown resource",
			args:          []string{"users"},
			expectAsError: true,
			expected:      "Resource \"users\" not found",
		},
		{
			name:     "help for project",
			args:     []string{"project", "--help"},
			expected: "Available Commands:",
		},
		{
			name:     "help for project",
			args:     []string{"project"},
			expected: "Available Commands:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.Execute(tt.args)
			if err != nil {
				if !tt.expectAsError {
					t.Errorf("ExecuteCommand() error = %v, expected no error", err)
				} else if !strings.Contains(err.Error(), tt.expected) {
					t.Errorf("ExecuteCommand() error = %v, expected it to contain %v", err, tt.expected)
				}
			} else if !strings.Contains(result.Output, tt.expected) {
				t.Errorf("ExecuteCommand() = %q, expected it to contain %q", result.Output, tt.expected)
			}
		})
	}
}
