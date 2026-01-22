package services

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
)

// TestGetOrGenerateCertificate_LoadError tests error handling when loading corrupt certificate files
func TestGetOrGenerateCertificate_LoadError(t *testing.T) {
	tempDir := t.TempDir()
	certFile := filepath.Join(tempDir, "corrupt.crt")
	keyFile := filepath.Join(tempDir, "corrupt.key")

	// Write corrupt certificate and key files
	if err := os.WriteFile(certFile, []byte("not a certificate"), 0o644); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte("not a key"), 0o644); err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	service := NewTLSService()
	_, err := service.GetOrGenerateCertificate(certFile, keyFile)

	// Should return error for corrupt files
	if err == nil {
		t.Error("Expected error when loading corrupt certificate files")
	}
}

// TestGetOrGenerateCertificate_ReadError tests error handling when certificate file cannot be read after loading
func TestGetOrGenerateCertificate_ReadError(t *testing.T) {
	tempDir := t.TempDir()
	certFile := filepath.Join(tempDir, "test.crt")
	keyFile := filepath.Join(tempDir, "test.key")

	// Generate a valid certificate first
	service := NewTLSService()
	_, err := service.generateSelfSignedCertificate()
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	// Write certificate files
	err = os.WriteFile(certFile, service.certPEM, 0o644)
	if err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}
	err = os.WriteFile(keyFile, service.keyPEM, 0o600)
	if err != nil {
		t.Fatalf("Failed to write key file: %v", err)
	}

	// Now try to load it - should succeed
	service2 := NewTLSService()
	loadedCert, err := service2.GetOrGenerateCertificate(certFile, keyFile)
	if err != nil {
		t.Fatalf("Failed to load certificate: %v", err)
	}

	// Verify certificate was loaded
	if len(loadedCert.Certificate) == 0 {
		t.Error("Expected certificate to be loaded")
	}
}

// TestGenerateSelfSignedCertificate_EmptyHostname tests certificate generation when hostname is empty
func TestGenerateSelfSignedCertificate_EmptyHostname(t *testing.T) {
	// This test ensures that even if hostname retrieval fails,
	// certificate generation still works
	service := NewTLSService()

	cert, err := service.generateSelfSignedCertificate()
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	// Parse certificate
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Should have some common name
	if x509Cert.Subject.CommonName == "" {
		t.Error("Expected certificate to have a common name")
	}
}

// TestParseCertificate_ValidCertificate tests parsing a valid certificate
func TestParseCertificate_ValidCertificate(t *testing.T) {
	service := NewTLSService()
	cert, err := service.generateSelfSignedCertificate()
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	// Now test ParseCertificate function
	x509Cert, err := ParseCertificate(&cert)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	if x509Cert == nil {
		t.Fatal("Expected non-nil certificate")
	}

	if x509Cert.Subject.CommonName == "" {
		t.Error("Expected subject common name to be set")
	}

	if len(x509Cert.Subject.Organization) == 0 {
		t.Error("Expected subject organization to be set")
	}

	if x509Cert.NotBefore.IsZero() {
		t.Error("Expected NotBefore to be set")
	}

	if x509Cert.NotAfter.IsZero() {
		t.Error("Expected NotAfter to be set")
	}

	if x509Cert.SerialNumber == nil {
		t.Error("Expected SerialNumber to be set")
	}
}

// TestGetOrGenerateCertificate_GenerateWhenFilesNotExist tests certificate generation when files don't exist
func TestGetOrGenerateCertificate_GenerateWhenFilesNotExist(t *testing.T) {
	service := NewTLSService()

	// Use non-existent file paths
	certFile := "/tmp/nonexistent-cert-file.crt"
	keyFile := "/tmp/nonexistent-key-file.key"

	cert, err := service.GetOrGenerateCertificate(certFile, keyFile)
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	// Should generate a new certificate
	if len(cert.Certificate) == 0 {
		t.Error("Expected certificate to be generated")
	}

	// Verify it's a valid TLS certificate
	_, err = tls.X509KeyPair(service.certPEM, service.keyPEM)
	if err != nil {
		t.Errorf("Generated certificate is not valid: %v", err)
	}
}

// TestLogCertificateInfo_NoPanic tests certificate logging doesn't panic
func TestLogCertificateInfo_NoPanic(t *testing.T) {
	service := NewTLSService()
	cert, err := service.generateSelfSignedCertificate()
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	// This should not panic
	service.logCertificateInfo(&cert)
}
