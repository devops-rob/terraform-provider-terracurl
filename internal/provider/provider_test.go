// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0.

package provider

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"terracurl": providerserver.NewProtocol6WithError(New("test")()),
}

var testAccProtoV6ProviderFactoriesWithEcho = map[string]func() (tfprotov6.ProviderServer, error){
	"terracurl": providerserver.NewProtocol6WithError(New("test")()),
	"echo":      echoprovider.NewProviderServer(),
}

// generateTestCert generates a self-signed TLS certificate and private key dynamically
// for testing purposes. This eliminates certificate expiration issues and removes the
// need to store test certificates in the codebase.
//
// The certificate is valid for 24 hours from generation time and includes:
//   - Subject: server.dc1.consul
//   - SANs: server.dc1.consul, localhost, 127.0.0.1
//   - Key: ECDSA P-256 (faster than RSA for tests)
//
// Returns PEM-encoded certificate and key as byte slices.
func generateTestCert() (certPEM, keyPEM []byte, err error) {
	// Generate ECDSA private key (P-256 curve)
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	notBefore := time.Now().Add(-1 * time.Hour) // Handle clock skew
	notAfter := notBefore.Add(24 * time.Hour)   // Valid for 24 hours

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{"HashiCorp Inc."},
			Country:       []string{"US"},
			Province:      []string{"CA"},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"101 Second Street"},
			PostalCode:    []string{"94105"},
			CommonName:    "server.dc1.consul",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"server.dc1.consul", "localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// PEM encode certificate
	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// PEM encode private key
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})

	return certPEM, keyPEM, nil
}

func createTLSServer() (*httptest.Server, string, string, error) {
	log.Println("createTLSServer() called...")

	// Generate fresh TLS certificate and key programmatically
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		log.Printf("Failed to generate test certificate: %v\n", err)
		return nil, "", "", err
	}

	// Save cert & key to temp files (still needed for tests that expect file paths)
	certFile, err := saveTempFile(certPEM)
	if err != nil {
		log.Printf("Failed to create cert file: %v\n", err)
		return nil, "", "", err
	}

	keyFile, err := saveTempFile(keyPEM)
	if err != nil {
		log.Printf("Failed to create key file: %v\n", err)
		return nil, "", "", err
	}

	log.Printf("Generated certificate saved to: %s\n", certFile)
	log.Printf("Generated private key saved to: %s\n", keyFile)

	// Log certificate details for debugging
	certHex := hex.EncodeToString(certPEM)
	keyHex := hex.EncodeToString(keyPEM)
	log.Printf("Cert hex (first 100 chars): %s...\n", certHex[:min(100, len(certHex))])
	log.Printf("Key hex (first 100 chars): %s...\n", keyHex[:min(100, len(keyHex))])

	// Parse the PEM-encoded certificate and key directly
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Printf("Failed to load X509KeyPair: %v\n", err)
		return nil, "", "", err
	}

	log.Println("TLS key pair loaded successfully!")

	// Create test TLS server
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message": "TLS test successful"}`))
		if err != nil {
			return
		}
	}))

	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	server.StartTLS()

	log.Println("TLS server started successfully!")

	return server, certFile, keyFile, nil
}

// min returns the minimum of two integers (helper for Go versions < 1.21)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func saveTempFile(data []byte) (string, error) {
	data = bytes.TrimSpace(data)
	tmpFile, err := os.CreateTemp("", "tls-test-*.pem")
	if err != nil {
		log.Printf("Error creating temp file: %v\n", err)
		return "", err
	}

	_, err = tmpFile.Write(data)
	if err != nil {
		log.Printf("Error writing to temp file %s: %v\n", tmpFile.Name(), err)
		return "", err
	}

	err = tmpFile.Close()
	if err != nil {
		log.Printf("Error closing temp file %s: %v\n", tmpFile.Name(), err)
		return "", err
	}

	log.Printf("Temp file written successfully: %s\n", tmpFile.Name())
	return tmpFile.Name(), nil
}

func TestSaveTempFile(t *testing.T) {
	log.Println("Running TestSaveTempFile...")

	// Fake certificate data.
	testCert := `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAOiL+Fc8m4n9MA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
...
-----END CERTIFICATE-----`

	// Attempt to save the file.
	filePath, err := saveTempFile([]byte(testCert))
	if err != nil {
		t.Fatalf("saveTempFile() failed: %v", err)
	}

	// Ensure the file exists.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("saveTempFile() did not create a file: %s", filePath)
	}

	// Read the file content.
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created temp file: %v", err)
	}

	// Ensure content is correct.
	if string(content) != testCert {
		t.Fatalf("File content does not match expected cert data")
	}

	// Cleanup.
	err = os.Remove(filePath)
	if err != nil {
		return
	}

	log.Printf("TestSaveTempFile passed! File created: %s\n", filePath)
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

func getTerraformVersion() (string, error) {
	cmd := exec.Command("terraform", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		versionParts := strings.Fields(lines[0])
		if len(versionParts) > 1 {
			return strings.TrimPrefix(versionParts[1], "v"), nil
		}
	}

	return "", fmt.Errorf("could not determine Terraform version")
}

// Run this function inside tests that should be skipped for Terraform <1.10
func skipIfTerraformIsLegacy(t *testing.T) {
	tfVersion, err := getTerraformVersion()
	if err != nil {
		t.Fatalf("Failed to get Terraform version: %s", err)
	}

	// Check if Terraform is <1.10
	if strings.HasPrefix(tfVersion, "1.0.") ||
		strings.HasPrefix(tfVersion, "1.1.") ||
		strings.HasPrefix(tfVersion, "1.2.") ||
		strings.HasPrefix(tfVersion, "1.3.") ||
		strings.HasPrefix(tfVersion, "1.4.") ||
		strings.HasPrefix(tfVersion, "1.5.") ||
		strings.HasPrefix(tfVersion, "1.6.") ||
		strings.HasPrefix(tfVersion, "1.7.") ||
		strings.HasPrefix(tfVersion, "1.8.") ||
		strings.HasPrefix(tfVersion, "1.9.") {
		t.Skip("Skipping test: Terraform version <1.10 detected")
	}
}
