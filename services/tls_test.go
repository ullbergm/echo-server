package services

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewTLSService(t *testing.T) {
	service := NewTLSService()
	if service == nil {
		t.Fatal("Expected NewTLSService to return non-nil service")
	}
}

func TestGenerateSelfSignedCertificate(t *testing.T) {
	service := NewTLSService()

	cert, err := service.generateSelfSignedCertificate()
	if err != nil {
		t.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	// Verify certificate is valid
	if len(cert.Certificate) == 0 {
		t.Fatal("Expected certificate to have at least one certificate")
	}

	// Parse the certificate
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Verify certificate properties
	if len(x509Cert.Subject.Organization) == 0 {
		t.Error("Expected certificate to have organization")
	} else if x509Cert.Subject.Organization[0] != "Echo Server" {
		t.Errorf("Expected organization to be 'Echo Server', got '%s'", x509Cert.Subject.Organization[0])
	}

	// Verify validity period
	now := time.Now()
	if x509Cert.NotBefore.After(now) {
		t.Error("Certificate NotBefore is in the future")
	}

	expectedExpiry := now.Add(365 * 24 * time.Hour)
	if x509Cert.NotAfter.Before(expectedExpiry.Add(-time.Minute)) || x509Cert.NotAfter.After(expectedExpiry.Add(time.Minute)) {
		t.Errorf("Certificate expiry is not approximately 1 year from now")
	}

	// Verify DNS names
	if len(x509Cert.DNSNames) < 1 {
		t.Error("Expected at least one DNS name")
	}
	hasLocalhost := false
	for _, name := range x509Cert.DNSNames {
		if name == "localhost" {
			hasLocalhost = true
			break
		}
	}
	if !hasLocalhost {
		t.Error("Expected 'localhost' to be in DNS names")
	}

	// Verify key usage
	if x509Cert.KeyUsage&x509.KeyUsageKeyEncipherment == 0 {
		t.Error("Expected KeyUsageKeyEncipherment to be set")
	}
	if x509Cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		t.Error("Expected KeyUsageDigitalSignature to be set")
	}

	// Verify extended key usage
	hasServerAuth := false
	for _, usage := range x509Cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth {
			hasServerAuth = true
			break
		}
	}
	if !hasServerAuth {
		t.Error("Expected ExtKeyUsageServerAuth to be set")
	}
}

func TestGetOrGenerateCertificate_Generate(t *testing.T) {
	service := NewTLSService()

	// Use non-existent paths to force generation
	cert, err := service.GetOrGenerateCertificate("/nonexistent/cert.pem", "/nonexistent/key.pem")
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	if len(cert.Certificate) == 0 {
		t.Fatal("Expected certificate to have at least one certificate")
	}

	// Verify PEM data was stored
	if len(service.certPEM) == 0 {
		t.Error("Expected certPEM to be populated")
	}
	if len(service.keyPEM) == 0 {
		t.Error("Expected keyPEM to be populated")
	}
}

func TestGetOrGenerateCertificate_LoadFromFiles(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "test.crt")
	keyFile := filepath.Join(tmpDir, "test.key")

	// Generate a certificate first
	service := NewTLSService()
	cert, err := service.generateSelfSignedCertificate()
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	// Write certificate and key to files
	if writeErr := os.WriteFile(certFile, service.certPEM, 0o644); writeErr != nil {
		t.Fatalf("Failed to write cert file: %v", writeErr)
	}
	if writeErr := os.WriteFile(keyFile, service.keyPEM, 0o600); writeErr != nil {
		t.Fatalf("Failed to write key file: %v", writeErr)
	}

	// Create new service and load from files
	service2 := NewTLSService()
	loadedCert, err := service2.GetOrGenerateCertificate(certFile, keyFile)
	if err != nil {
		t.Fatalf("Failed to load certificate from files: %v", err)
	}

	if len(loadedCert.Certificate) == 0 {
		t.Fatal("Expected loaded certificate to have at least one certificate")
	}

	// Verify it's the same certificate
	originalX509, _ := x509.ParseCertificate(cert.Certificate[0])
	loadedX509, _ := x509.ParseCertificate(loadedCert.Certificate[0])

	if originalX509.SerialNumber.Cmp(loadedX509.SerialNumber) != 0 {
		t.Error("Loaded certificate serial number doesn't match original")
	}
}

func TestParseCertificate(t *testing.T) {
	service := NewTLSService()
	cert, err := service.generateSelfSignedCertificate()
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	x509Cert, err := ParseCertificate(&cert)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	if x509Cert == nil {
		t.Fatal("Expected non-nil x509 certificate")
	}

	if len(x509Cert.Subject.Organization) == 0 {
		t.Error("Expected certificate to have organization")
	} else if x509Cert.Subject.Organization[0] != "Echo Server" {
		t.Errorf("Expected organization to be 'Echo Server', got '%s'", x509Cert.Subject.Organization[0])
	}
}

func TestParseCertificate_EmptyCertificate(t *testing.T) {
	emptyCert := tls.Certificate{}

	_, err := ParseCertificate(&emptyCert)
	if err == nil {
		t.Error("Expected error when parsing empty certificate")
	}
}

func TestGetOrGenerateCertificate_InvalidFiles(t *testing.T) {
	// Create temporary directory with invalid files
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "invalid.crt")
	keyFile := filepath.Join(tmpDir, "invalid.key")

	// Write invalid data
	if err := os.WriteFile(certFile, []byte("invalid cert data"), 0o644); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte("invalid key data"), 0o600); err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	service := NewTLSService()
	_, err := service.GetOrGenerateCertificate(certFile, keyFile)
	if err == nil {
		t.Error("Expected error when loading invalid certificate files")
	}
}
