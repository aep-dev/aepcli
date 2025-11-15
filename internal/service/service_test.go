package service

import (
	"net/http"
	"strings"
	"testing"
)

func TestService_ExecuteCommand_ListResources(t *testing.T) {
	// Test setup
	svc, err := NewServiceCommand(getTestAPI(), nil, false, false, false, "")
	if err != nil {
		t.Fatalf("Failed to create service command: %v", err)
	}

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

func TestNewServiceCommand_Insecure(t *testing.T) {
	tests := []struct {
		name     string
		insecure bool
	}{
		{
			name:     "secure client",
			insecure: false,
		},
		{
			name:     "insecure client",
			insecure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewServiceCommand(getTestAPI(), nil, false, false, tt.insecure, "")
			if err != nil {
				t.Fatalf("Failed to create service command: %v", err)
			}

			if svc.Insecure != tt.insecure {
				t.Errorf("NewServiceCommand() insecure = %v, want %v", svc.Insecure, tt.insecure)
			}

			// Check if the client has the correct TLS configuration
			if tt.insecure {
				transport, ok := svc.Client.Transport.(*http.Transport)
				if !ok {
					t.Error("Expected HTTP transport to be set for insecure client")
					return
				}
				if transport.TLSClientConfig == nil {
					t.Error("Expected TLS config to be set for insecure client")
					return
				}
				if !transport.TLSClientConfig.InsecureSkipVerify {
					t.Error("Expected InsecureSkipVerify to be true for insecure client")
				}
			} else {
				// For secure clients, we should have the default transport or no custom transport
				if svc.Client.Transport != nil {
					transport, ok := svc.Client.Transport.(*http.Transport)
					if ok && transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify {
						t.Error("Expected InsecureSkipVerify to be false for secure client")
					}
				}
			}
		})
	}
}
