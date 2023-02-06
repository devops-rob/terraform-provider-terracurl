package provider

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

func startMockServer() *httptest.Server {
	// Load certificate and key files
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		panic(err)
	}

	// Load the CA file
	caCert, err := ioutil.ReadFile("ca.pem")
	if err != nil {
		panic(err)
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		panic("Failed to add CA cert to pool")
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("TLS Server"))
	}))
	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	ts.StartTLS()

	return ts
}
