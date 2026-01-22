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

// TestLogCertificateInfo tests the logCertificateInfo method
func TestLogCertificateInfo(t *testing.T) {
	service := NewTLSService()

	// Generate a certificate first
	cert, err := service.generateSelfSignedCertificate()
	if err != nil {
		t.Fatalf("Failed to generate certificate: %v", err)
	}

	// This should not panic
	service.logCertificateInfo(&cert)
}

// TestLogCertificateInfo_EmptyCert tests logCertificateInfo with empty certificate
func TestLogCertificateInfo_EmptyCert(t *testing.T) {
	service := NewTLSService()

	emptyCert := &tls.Certificate{}

	// This should not panic, just return early
	service.logCertificateInfo(emptyCert)
}

// TestLogCertificateInfo_InvalidCert tests logCertificateInfo with invalid certificate data
func TestLogCertificateInfo_InvalidCert(t *testing.T) {
	service := NewTLSService()

	invalidCert := &tls.Certificate{
		Certificate: [][]byte{[]byte("invalid cert data")},
	}

	// This should not panic, just log a warning
	service.logCertificateInfo(invalidCert)
}

// TestGetOrGenerateCertificate_CertExistsKeyMissing tests when cert exists but key doesn't
func TestGetOrGenerateCertificate_CertExistsKeyMissing(t *testing.T) {
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "exists.crt")
	keyFile := filepath.Join(tmpDir, "missing.key")

	// Generate a certificate
	service := NewTLSService()
	_, genErr := service.generateSelfSignedCertificate()
	if genErr != nil {
		t.Fatalf("Failed to generate certificate: %v", genErr)
	}

	// Write only the cert file
	if err := os.WriteFile(certFile, service.certPEM, 0o644); err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	// Should generate a new certificate since key is missing
	service2 := NewTLSService()
	cert, err := service2.GetOrGenerateCertificate(certFile, keyFile)
	if err != nil {
		t.Fatalf("Failed when key file missing: %v", err)
	}

	if len(cert.Certificate) == 0 {
		t.Fatal("Expected certificate to be generated")
	}
}

// TestGenerateSelfSignedCertificate_MultipleCalls tests generating multiple certificates
func TestGenerateSelfSignedCertificate_MultipleCalls(t *testing.T) {
	service := NewTLSService()

	cert1, err := service.generateSelfSignedCertificate()
	if err != nil {
		t.Fatalf("Failed to generate first certificate: %v", err)
	}

	cert2, err := service.generateSelfSignedCertificate()
	if err != nil {
		t.Fatalf("Failed to generate second certificate: %v", err)
	}

	// Each certificate should have a unique serial number
	x509Cert1, _ := x509.ParseCertificate(cert1.Certificate[0])
	x509Cert2, _ := x509.ParseCertificate(cert2.Certificate[0])

	if x509Cert1.SerialNumber.Cmp(x509Cert2.SerialNumber) == 0 {
		t.Error("Expected different serial numbers for different certificates")
	}
}
