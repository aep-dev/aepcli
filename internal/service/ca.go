package service

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
)

// loadCACertificate loads a CA certificate from the specified file path and returns a *x509.CertPool
// that includes the system CA certificates plus the custom CA certificate.
func loadCACertificate(caCertPath string) (*x509.CertPool, error) {
	if caCertPath == "" {
		// Return system CA pool when no custom CA is specified
		return x509.SystemCertPool()
	}

	slog.Debug("Loading custom CA certificate", "path", caCertPath)

	// Read the CA certificate file
	caCertData, err := os.ReadFile(caCertPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Failed to read CA certificate from %s: file does not exist\n\nTo fix this issue:\n  1. Verify the file path is correct\n  2. Ensure the file exists and is readable", caCertPath)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("Failed to read CA certificate from %s: permission denied\n\nTo fix this issue:\n  1. Verify you have read permissions for the file\n  2. Check file permissions with: ls -l %s", caCertPath, caCertPath)
		}
		return nil, fmt.Errorf("Failed to read CA certificate from %s: %v", caCertPath, err)
	}

	// Start with system CA certificates
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		slog.Warn("Failed to load system CA certificates, using empty pool", "error", err)
		caCertPool = x509.NewCertPool()
	} else {
		slog.Debug("System CA certificates loaded from system trust store")
	}

	// Parse the PEM block first to validate format
	block, _ := pem.Decode(caCertData)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("Failed to parse CA certificate from %s: not valid PEM format\n\nExpected format:\n  -----BEGIN CERTIFICATE-----\n  ...\n  -----END CERTIFICATE-----\n\nUse 'openssl x509 -in %s -text -noout' to verify the certificate.", caCertPath, caCertPath)
	}

	// Parse the certificate to ensure it's valid
	_, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse CA certificate from %s: invalid certificate data: %v\n\nUse 'openssl x509 -in %s -text -noout' to verify the certificate.", caCertPath, err, caCertPath)
	}

	// Add the custom CA certificate
	if !caCertPool.AppendCertsFromPEM(caCertData) {
		return nil, fmt.Errorf("Failed to add CA certificate from %s to certificate pool", caCertPath)
	}

	slog.Debug("Custom CA certificate loaded", "path", caCertPath)
	return caCertPool, nil
}