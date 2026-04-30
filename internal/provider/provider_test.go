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

// Mock private key for TLS server.
const localKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHa08Nf/lf7KXSMcRwnhNOI5rpJsykbo4ZGImsZndeHYoAoGCCqGSM49
AwEHoUQDQgAEjLj2Ay/hhLhJ1cC5Rp7/bkucDS+MrS8Te7HpXmQJAQt4DsMWbP9K
J9dc0LcE8rTwitkoLiTtjMl/y9J+I6jqHw==
-----END EC PRIVATE KEY-----`

func createTLSServer() (*httptest.Server, string, string, error) {
	log.Println("createTLSServer() called...")

	// Save cert & key to temp files.

	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		return nil, "", "", err
	}

	certFile, err := saveTempFile(certPEM)
	if err != nil {
		return nil, "", "", err
	}

	keyFile, err := saveTempFile(keyPEM)
	if err != nil {
		return nil, "", "", err
	}

	// Read and print file contents.
	certContent, err := os.ReadFile(certFile)
	if err != nil {
		log.Printf("Failed to read cert file: %v\n", err)
		return nil, "", "", err
	}

	keyContent, err := os.ReadFile(keyFile)
	if err != nil {
		log.Printf("Failed to read key file: %v\n", err)
		return nil, "", "", err
	}

	log.Printf("Cert file content:\n%s", string(certContent))
	log.Printf("Key file content:\n%s", string(keyContent))

	certHex := hex.EncodeToString(certContent)
	keyHex := hex.EncodeToString(keyContent)

	log.Printf("Cert file hex:\n%s", certHex)
	log.Printf("Key file hex:\n%s", keyHex)

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

func generateTestCert() ([]byte, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),

		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},

		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	return certPEM, keyPEM, nil
}
