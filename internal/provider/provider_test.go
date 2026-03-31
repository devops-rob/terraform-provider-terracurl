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
	"encoding/pem"
	"fmt"
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
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
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

// generateSelfSignedCert creates a fresh self-signed ECDSA certificate and key
// valid for 1 hour, with SANs for localhost and 127.0.0.1. Returns PEM-encoded
// cert and key bytes.
func generateSelfSignedCert() ([]byte, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "terracurl-test",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM, nil
}

func createTLSServer() (*httptest.Server, string, string, error) {
	log.Println("createTLSServer() called...")

	certPEM, keyPEM, err := generateSelfSignedCert()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate self-signed cert: %w", err)
	}

	// Save cert & key to temp files.
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

	fmt.Printf("CertFile: %s, KeyFile: %s\n", certFile, keyFile)

	// Load the TLS key pair.
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Printf("Failed to load X509KeyPair: %v\n", err)
		return nil, "", "", err
	}

	log.Println("TLS key pair loaded successfully!")

	// Create test TLS server.
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
