// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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

type mockClient struct {
	mock.Mock
}

func (m *mockClient) Do(r *http.Request) (*http.Response, error) {
	args := m.Mock.Called(r)

	if resp, ok := args.Get(0).(*http.Response); ok {
		return resp, args.Error(1)
	}

	return nil, args.Error(1)
}

const localCert = `-----BEGIN CERTIFICATE-----
MIICnDCCAkOgAwIBAgIRAJ7vRKfNUfgTzPf3A2usN5MwCgYIKoZIzj0EAwIwgbkx
CzAJBgNVBAYTAlVTMQswCQYDVQQIEwJDQTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNj
bzEaMBgGA1UECRMRMTAxIFNlY29uZCBTdHJlZXQxDjAMBgNVBBETBTk0MTA1MRcw
FQYDVQQKEw5IYXNoaUNvcnAgSW5jLjFAMD4GA1UEAxM3Q29uc3VsIEFnZW50IENB
IDE3OTE0MzkwMDM4OTUwMjI2MjM2Njc1OTk3NzcwNTA5NjcxNjY5MzAeFw0yNTAy
MjcxMzMxMDVaFw0yNjAyMjcxMzMxMDVaMBwxGjAYBgNVBAMTEXNlcnZlci5kYzEu
Y29uc3VsMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEjLj2Ay/hhLhJ1cC5Rp7/
bkucDS+MrS8Te7HpXmQJAQt4DsMWbP9KJ9dc0LcE8rTwitkoLiTtjMl/y9J+I6jq
H6OBxzCBxDAOBgNVHQ8BAf8EBAMCBaAwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwKQYDVR0OBCIEIC31ZCjRSd188aHLNsUi6z25
Dm0CsyHqBhCZ/1Xak9+RMCsGA1UdIwQkMCKAIKEBxmHeTfADtdpbF1Sww30JXIln
rYqEyg+0PDbZW6yXMC0GA1UdEQQmMCSCEXNlcnZlci5kYzEuY29uc3Vsgglsb2Nh
bGhvc3SHBH8AAAEwCgYIKoZIzj0EAwIDRwAwRAIgI3d9t7SOR9RaTrnFWGh+igXE
4bZYvsUcWL2V9mA5T3MCIFH7XfGUEwuviYHt6Py1X9yaI5lcRxjgSOkFMMIsoY01
-----END CERTIFICATE-----`

// Mock private key for TLS server
const localKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHa08Nf/lf7KXSMcRwnhNOI5rpJsykbo4ZGImsZndeHYoAoGCCqGSM49
AwEHoUQDQgAEjLj2Ay/hhLhJ1cC5Rp7/bkucDS+MrS8Te7HpXmQJAQt4DsMWbP9K
J9dc0LcE8rTwitkoLiTtjMl/y9J+I6jqHw==
-----END EC PRIVATE KEY-----`

const localCaCert = `-----BEGIN CERTIFICATE-----
MIIC7zCCApSgAwIBAgIRAIbF3DbRPw3xEog94qQ5oZUwCgYIKoZIzj0EAwIwgbkx
CzAJBgNVBAYTAlVTMQswCQYDVQQIEwJDQTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNj
bzEaMBgGA1UECRMRMTAxIFNlY29uZCBTdHJlZXQxDjAMBgNVBBETBTk0MTA1MRcw
FQYDVQQKEw5IYXNoaUNvcnAgSW5jLjFAMD4GA1UEAxM3Q29uc3VsIEFnZW50IENB
IDE3OTE0MzkwMDM4OTUwMjI2MjM2Njc1OTk3NzcwNTA5NjcxNjY5MzAeFw0yNTAy
MjcxMzMxMDJaFw0zMDAyMjYxMzMxMDJaMIG5MQswCQYDVQQGEwJVUzELMAkGA1UE
CBMCQ0ExFjAUBgNVBAcTDVNhbiBGcmFuY2lzY28xGjAYBgNVBAkTETEwMSBTZWNv
bmQgU3RyZWV0MQ4wDAYDVQQREwU5NDEwNTEXMBUGA1UEChMOSGFzaGlDb3JwIElu
Yy4xQDA+BgNVBAMTN0NvbnN1bCBBZ2VudCBDQSAxNzkxNDM5MDAzODk1MDIyNjIz
NjY3NTk5Nzc3MDUwOTY3MTY2OTMwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQ/
UUzw4oVoTwlfz3LyOcEwjkgnDTkRAswgg3uX8uMHnXuAAowtMWZn1YoZBtSRzkk8
XwO/3Chhwicg6P1lww97o3sweTAOBgNVHQ8BAf8EBAMCAYYwDwYDVR0TAQH/BAUw
AwEB/zApBgNVHQ4EIgQgoQHGYd5N8AO12lsXVLDDfQlciWetioTKD7Q8NtlbrJcw
KwYDVR0jBCQwIoAgoQHGYd5N8AO12lsXVLDDfQlciWetioTKD7Q8NtlbrJcwCgYI
KoZIzj0EAwIDSQAwRgIhAKryzgZzELmfEMhP+vMVldV1qZg20WW+WQvlHCXrbB5b
AiEAuQ3n4H00QArzgaj+57yW7KCJu1Pi1WBXsfRStDUte9E=
-----END CERTIFICATE-----`

const localCaKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIMn8/saUQ0mFuujfsN/wWW7scKbzoPaN0O4aaEkLL5e+oAoGCCqGSM49
AwEHoUQDQgAEP1FM8OKFaE8JX89y8jnBMI5IJw05EQLMIIN7l/LjB517gAKMLTFm
Z9WKGQbUkc5JPF8Dv9woYcInIOj9ZcMPew==
-----END EC PRIVATE KEY-----`

func createTLSServer() (*httptest.Server, string, string, error) {
	log.Println("createTLSServer() called...")

	// Save cert & key to temp files
	certFile, err := saveTempFile([]byte(localCert))
	if err != nil {
		log.Printf("Failed to create cert file: %v\n", err)
		return nil, "", "", err
	}

	keyFile, err := saveTempFile([]byte(localKey))
	if err != nil {
		log.Printf("Failed to create key file: %v\n", err)
		return nil, "", "", err
	}

	// Read and print file contents
	certContent, err := ioutil.ReadFile(certFile)
	if err != nil {
		log.Printf("Failed to read cert file: %v\n", err)
		return nil, "", "", err
	}

	keyContent, err := ioutil.ReadFile(keyFile)
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

	// Load the TLS key pair
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
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

func saveTempFile(data []byte) (string, error) {
	data = bytes.TrimSpace(data) // Ensure no trailing spaces or newlines

	tmpFile, err := ioutil.TempFile("", "tls-test-*.pem")
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

	// Fake certificate data
	testCert := `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAOiL+Fc8m4n9MA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
...
-----END CERTIFICATE-----`

	// Attempt to save the file
	filePath, err := saveTempFile([]byte(testCert))
	if err != nil {
		t.Fatalf("saveTempFile() failed: %v", err)
	}

	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("saveTempFile() did not create a file: %s", filePath)
	}

	// Read the file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created temp file: %v", err)
	}

	// Ensure content is correct
	if string(content) != testCert {
		t.Fatalf("File content does not match expected cert data")
	}

	// Cleanup
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
