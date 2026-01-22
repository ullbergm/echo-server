package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

// TLSService handles TLS certificate generation and loading
type TLSService struct {
	certPEM []byte
	keyPEM  []byte
}

// NewTLSService creates a new TLS service
func NewTLSService() *TLSService {
	return &TLSService{}
}

// GetOrGenerateCertificate loads certificates from files or generates self-signed certificates
func (t *TLSService) GetOrGenerateCertificate(certFile, keyFile string) (tls.Certificate, error) {
	// Try to load from files first
	if _, err := os.Stat(certFile); err == nil {
		if _, errKey := os.Stat(keyFile); errKey == nil {
			log.Printf("Loading TLS certificates from files: cert=%s, key=%s", certFile, keyFile)

			cert, loadErr := tls.LoadX509KeyPair(certFile, keyFile)
			if loadErr != nil {
				return tls.Certificate{}, fmt.Errorf("failed to load certificate files: %w", loadErr)
			}

			// Store PEM data for logging
			// #nosec G304 -- Certificate file path is from configuration, not user input
			t.certPEM, err = os.ReadFile(certFile)
			if err != nil {
				log.Printf("Warning: Failed to read certificate file for logging: %v", err)
			}
			// #nosec G304 -- Key file path is from configuration, not user input
			t.keyPEM, err = os.ReadFile(keyFile)
			if err != nil {
				log.Printf("Warning: Failed to read key file for logging: %v", err)
			}

			// Log certificate information
			t.logCertificateInfo(&cert)

			return cert, nil
		}
	}

	// Generate self-signed certificate
	log.Printf("Certificate files not found, generating self-signed certificate")
	return t.generateSelfSignedCertificate()
}

// generateSelfSignedCertificate creates a new self-signed TLS certificate
func (t *TLSService) generateSelfSignedCertificate() (tls.Certificate, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate serial number: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		log.Printf("Warning: Failed to get hostname: %v, using default", err)
		hostname = "echo-server"
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // 1 year

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Echo Server"},
			CommonName:   hostname,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{hostname, "localhost"},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate to PEM
	t.certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM
	t.keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Load certificate
	cert, err := tls.X509KeyPair(t.certPEM, t.keyPEM)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load generated certificate: %w", err)
	}

	log.Printf("Generated self-signed certificate: CN=%s, NotBefore=%s, NotAfter=%s",
		hostname, notBefore.Format(time.RFC3339), notAfter.Format(time.RFC3339))

	// Log certificate information
	t.logCertificateInfo(&cert)

	return cert, nil
}

// logCertificateInfo logs details about the certificate
func (t *TLSService) logCertificateInfo(cert *tls.Certificate) {
	if len(cert.Certificate) == 0 {
		return
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		log.Printf("Warning: Failed to parse certificate for logging: %v", err)
		return
	}

	log.Printf("TLS Certificate Info:")
	log.Printf("  Subject: %s", x509Cert.Subject.String())
	log.Printf("  Issuer: %s", x509Cert.Issuer.String())
	log.Printf("  NotBefore: %s", x509Cert.NotBefore.Format(time.RFC3339))
	log.Printf("  NotAfter: %s", x509Cert.NotAfter.Format(time.RFC3339))
	log.Printf("  SerialNumber: %s", x509Cert.SerialNumber.String())
	if len(x509Cert.DNSNames) > 0 {
		log.Printf("  DNSNames: %v", x509Cert.DNSNames)
	}
}

// ParseCertificate parses the X.509 certificate from the TLS certificate
func ParseCertificate(cert *tls.Certificate) (*x509.Certificate, error) {
	if len(cert.Certificate) == 0 {
		return nil, fmt.Errorf("no certificate found")
	}

	return x509.ParseCertificate(cert.Certificate[0])
}
