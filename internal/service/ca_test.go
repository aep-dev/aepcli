package service

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCACertificate(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string) error
		caCertPath    string
		expectError   bool
		errorContains string
	}{
		{
			name:        "empty path returns system certs",
			caCertPath:  "",
			expectError: false,
		},
		{
			name:          "nonexistent file",
			caCertPath:    "/nonexistent/file.pem",
			expectError:   true,
			errorContains: "file does not exist",
		},
		{
			name: "valid PEM certificate",
			setupFunc: func(path string) error {
				// Create a sample valid PEM certificate
				validPEM := `-----BEGIN CERTIFICATE-----
MIIDBTCCAe2gAwIBAgIUJkQOlSxfNroAhpQ9RvvaA+NpG5IwDQYJKoZIhvcNAQEL
BQAwEjEQMA4GA1UEAwwHVGVzdCBDQTAeFw0yNTExMTUyMDE5NTBaFw0yNjExMTUy
MDE5NTBaMBIxEDAOBgNVBAMMB1Rlc3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IB
DwAwggEKAoIBAQCkUKFcR5T8Wnf7sAfgHMNoVHxZAToufffFPP/UitdigSrokhDT
3SQ37dYJ/wrerGdBT5kCnfFnfyO7fE+0n4zKxFe4AAt198K+8lBi4/PyepRDYOtO
BkATnyu6idXvg5cFja5cg+qJ1Ccua8e56R8x4e5nOmVmKdKCP8hEE33cROhFkDqp
z7K7lVqjWOSK9nzKG6Rvsz02/1iAW6/LN3nqII65ju1uSIvweEr8uRvjv70zt1Mn
lajQTbjj18tqthaP2BVfNlw/OMG7ijzbc8N7bDgW/lUz41vb15uA3dHdZnv7OZGF
vG8KUbG9htBZOnU2oXV3qnzrRx8cz2DBpMcNAgMBAAGjUzBRMB0GA1UdDgQWBBSm
OgGLv48OGFBQrrbxrHOCYnV5lzAfBgNVHSMEGDAWgBSmOgGLv48OGFBQrrbxrHOC
YnV5lzAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBJwynG1SZ2
8R2shQmASL91t1fnFoZfAm2bEpUvw03BjjLW+wMbYx3xKER3ZZ57Rizagj/LFZPQ
2sASrpArJOXBVZHuKUwKnfp0uCpy5n+gdqPfEuGOp9wm5HQaK3JZaqLo6/3AM4Oh
6UAOxSzo6cf1PWGpxoHn51JjBlwNnQbxgtq4YhmTuuoouXGPqTD8QrOa5Yu5qUh5
jfqo/7/4VI0vahto6nF0q369a+hRuo+sHqOdX0i343lthZoa5SaiOuR0yihZbOhx
2l4rT3C3MuThnsar+axNQTXm+HEQuuc/eDqzdGUkY9I14LuKz3C+DwEwmbrcrGRM
yUOQDT7vpIja
-----END CERTIFICATE-----`
				return os.WriteFile(path, []byte(validPEM), 0644)
			},
			expectError: false,
		},
		{
			name:       "invalid PEM format",
			caCertPath: "temp", // This will trigger temp file creation
			setupFunc: func(path string) error {
				// Create an invalid PEM file that looks like PEM with proper Base64 encoding but invalid cert data
				invalidPEM := `-----BEGIN CERTIFICATE-----
VGhpcyBpcyBub3QgYSB2YWxpZCBjZXJ0aWZpY2F0ZSBjb250ZW50
-----END CERTIFICATE-----`
				return os.WriteFile(path, []byte(invalidPEM), 0644)
			},
			expectError:   true,
			errorContains: "invalid certificate data",
		},
		{
			name:       "completely invalid format",
			caCertPath: "temp", // This will trigger temp file creation
			setupFunc: func(path string) error {
				// Create a file without proper PEM headers
				invalidFormat := `This is not a PEM file at all`
				return os.WriteFile(path, []byte(invalidFormat), 0644)
			},
			expectError:   true,
			errorContains: "not valid PEM format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testFilePath string

			if tt.caCertPath == "" {
				// Empty path test
				testFilePath = tt.caCertPath
			} else if tt.caCertPath[0] == '/' && filepath.Dir(tt.caCertPath) == "/nonexistent" {
				// Non-existent file test - use the path as is
				testFilePath = tt.caCertPath
			} else {
				// Create a temporary file for other tests
				tmpFile, err := os.CreateTemp("", "test-ca-*.pem")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.Close()
				testFilePath = tmpFile.Name()
				defer os.Remove(testFilePath)

				if tt.setupFunc != nil {
					err := tt.setupFunc(testFilePath)
					if err != nil {
						t.Fatalf("Setup function failed: %v", err)
					}
				}
			}

			certPool, err := loadCACertificate(testFilePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if certPool == nil {
					t.Error("Expected cert pool but got nil")
				}
			}
		})
	}
}

func TestNewServiceCommand_CACertificates(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(string) error
		caCertPath  string
		expectError bool
		checkTLS    func(*testing.T, *http.Client)
	}{
		{
			name:        "no ca cert path",
			caCertPath:  "",
			expectError: false,
			checkTLS: func(t *testing.T, client *http.Client) {
				// Should not have custom transport
				if client.Transport != nil {
					t.Error("Expected no custom transport for empty CA cert path")
				}
			},
		},
		{
			name:       "valid ca cert path",
			caCertPath: "temp", // This will trigger temp file creation
			setupFunc: func(path string) error {
				validPEM := `-----BEGIN CERTIFICATE-----
MIIDBTCCAe2gAwIBAgIUJkQOlSxfNroAhpQ9RvvaA+NpG5IwDQYJKoZIhvcNAQEL
BQAwEjEQMA4GA1UEAwwHVGVzdCBDQTAeFw0yNTExMTUyMDE5NTBaFw0yNjExMTUy
MDE5NTBaMBIxEDAOBgNVBAMMB1Rlc3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IB
DwAwggEKAoIBAQCkUKFcR5T8Wnf7sAfgHMNoVHxZAToufffFPP/UitdigSrokhDT
3SQ37dYJ/wrerGdBT5kCnfFnfyO7fE+0n4zKxFe4AAt198K+8lBi4/PyepRDYOtO
BkATnyu6idXvg5cFja5cg+qJ1Ccua8e56R8x4e5nOmVmKdKCP8hEE33cROhFkDqp
z7K7lVqjWOSK9nzKG6Rvsz02/1iAW6/LN3nqII65ju1uSIvweEr8uRvjv70zt1Mn
lajQTbjj18tqthaP2BVfNlw/OMG7ijzbc8N7bDgW/lUz41vb15uA3dHdZnv7OZGF
vG8KUbG9htBZOnU2oXV3qnzrRx8cz2DBpMcNAgMBAAGjUzBRMB0GA1UdDgQWBBSm
OgGLv48OGFBQrrbxrHOCYnV5lzAfBgNVHSMEGDAWgBSmOgGLv48OGFBQrrbxrHOC
YnV5lzAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBJwynG1SZ2
8R2shQmASL91t1fnFoZfAm2bEpUvw03BjjLW+wMbYx3xKER3ZZ57Rizagj/LFZPQ
2sASrpArJOXBVZHuKUwKnfp0uCpy5n+gdqPfEuGOp9wm5HQaK3JZaqLo6/3AM4Oh
6UAOxSzo6cf1PWGpxoHn51JjBlwNnQbxgtq4YhmTuuoouXGPqTD8QrOa5Yu5qUh5
jfqo/7/4VI0vahto6nF0q369a+hRuo+sHqOdX0i343lthZoa5SaiOuR0yihZbOhx
2l4rT3C3MuThnsar+axNQTXm+HEQuuc/eDqzdGUkY9I14LuKz3C+DwEwmbrcrGRM
yUOQDT7vpIja
-----END CERTIFICATE-----`
				return os.WriteFile(path, []byte(validPEM), 0644)
			},
			expectError: false,
			checkTLS: func(t *testing.T, client *http.Client) {
				if client.Transport == nil {
					t.Error("Expected HTTP transport to be set for CA cert")
					return
				}
				transport, ok := client.Transport.(*http.Transport)
				if !ok {
					t.Error("Expected HTTP transport to be *http.Transport type")
					return
				}
				if transport.TLSClientConfig == nil {
					t.Error("Expected TLS config to be set for CA cert")
					return
				}
				if transport.TLSClientConfig.RootCAs == nil {
					t.Error("Expected RootCAs to be set for CA cert")
				}
				if transport.TLSClientConfig.InsecureSkipVerify {
					t.Error("Expected InsecureSkipVerify to be false when using CA cert")
				}
			},
		},
		{
			name:        "invalid ca cert path",
			caCertPath:  "/nonexistent/file.pem",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testFilePath string

			if tt.caCertPath == "" {
				testFilePath = ""
			} else if tt.caCertPath[0] == '/' && filepath.Dir(tt.caCertPath) == "/nonexistent" {
				testFilePath = tt.caCertPath
			} else {
				tmpFile, err := os.CreateTemp("", "test-ca-*.pem")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.Close()
				testFilePath = tmpFile.Name()
				defer os.Remove(testFilePath)

				if tt.setupFunc != nil {
					err := tt.setupFunc(testFilePath)
					if err != nil {
						t.Fatalf("Setup function failed: %v", err)
					}
				}
			}

			svc, err := NewServiceCommand(getTestAPI(), nil, false, false, false, testFilePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if svc == nil {
					t.Error("Expected service command but got nil")
					return
				}
				if svc.CACertPath != testFilePath {
					t.Errorf("Expected CACertPath to be %s, got %s", testFilePath, svc.CACertPath)
				}
				if tt.checkTLS != nil {
					tt.checkTLS(t, svc.Client)
				}
			}
		})
	}
}

func TestNewServiceCommand_InsecureOverridesCA(t *testing.T) {
	// Create a temporary CA certificate file
	tmpFile, err := os.CreateTemp("", "test-ca-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	validPEM := `-----BEGIN CERTIFICATE-----
MIIDBTCCAe2gAwIBAgIUJkQOlSxfNroAhpQ9RvvaA+NpG5IwDQYJKoZIhvcNAQEL
BQAwEjEQMA4GA1UEAwwHVGVzdCBDQTAeFw0yNTExMTUyMDE5NTBaFw0yNjExMTUy
MDE5NTBaMBIxEDAOBgNVBAMMB1Rlc3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IB
DwAwggEKAoIBAQCkUKFcR5T8Wnf7sAfgHMNoVHxZAToufffFPP/UitdigSrokhDT
3SQ37dYJ/wrerGdBT5kCnfFnfyO7fE+0n4zKxFe4AAt198K+8lBi4/PyepRDYOtO
BkATnyu6idXvg5cFja5cg+qJ1Ccua8e56R8x4e5nOmVmKdKCP8hEE33cROhFkDqp
z7K7lVqjWOSK9nzKG6Rvsz02/1iAW6/LN3nqII65ju1uSIvweEr8uRvjv70zt1Mn
lajQTbjj18tqthaP2BVfNlw/OMG7ijzbc8N7bDgW/lUz41vb15uA3dHdZnv7OZGF
vG8KUbG9htBZOnU2oXV3qnzrRx8cz2DBpMcNAgMBAAGjUzBRMB0GA1UdDgQWBBSm
OgGLv48OGFBQrrbxrHOCYnV5lzAfBgNVHSMEGDAWgBSmOgGLv48OGFBQrrbxrHOC
YnV5lzAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBJwynG1SZ2
8R2shQmASL91t1fnFoZfAm2bEpUvw03BjjLW+wMbYx3xKER3ZZ57Rizagj/LFZPQ
2sASrpArJOXBVZHuKUwKnfp0uCpy5n+gdqPfEuGOp9wm5HQaK3JZaqLo6/3AM4Oh
6UAOxSzo6cf1PWGpxoHn51JjBlwNnQbxgtq4YhmTuuoouXGPqTD8QrOa5Yu5qUh5
jfqo/7/4VI0vahto6nF0q369a+hRuo+sHqOdX0i343lthZoa5SaiOuR0yihZbOhx
2l4rT3C3MuThnsar+axNQTXm+HEQuuc/eDqzdGUkY9I14LuKz3C+DwEwmbrcrGRM
yUOQDT7vpIja
-----END CERTIFICATE-----`
	err = os.WriteFile(tmpFile.Name(), []byte(validPEM), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Test that insecure flag overrides CA certificate
	svc, err := NewServiceCommand(getTestAPI(), nil, false, false, true, tmpFile.Name())
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	transport, ok := svc.Client.Transport.(*http.Transport)
	if !ok {
		t.Error("Expected HTTP transport to be set")
		return
	}

	if transport.TLSClientConfig == nil {
		t.Error("Expected TLS config to be set")
		return
	}

	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify to be true when insecure flag is set")
	}

	// RootCAs should not be set when insecure is true
	if transport.TLSClientConfig.RootCAs != nil {
		t.Error("Expected RootCAs to be nil when insecure flag is set")
	}
}
