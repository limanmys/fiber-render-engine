package auth

import (
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestKeycloakHTTPClientRejectsUntrustedCertificate(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	response, err := newKeycloakHTTPClient(nil).R().Get(server.URL)
	if err == nil {
		response.RawResponse.Body.Close()
		t.Fatal("expected the Keycloak client to reject an untrusted TLS certificate")
	}
}

func TestKeycloakHTTPClientAcceptsTrustedCertificate(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(server.Certificate())

	response, err := newKeycloakHTTPClient(rootCAs).R().Get(server.URL)
	if err != nil {
		t.Fatalf("expected the Keycloak client to accept an explicitly trusted TLS certificate: %v", err)
	}
	t.Cleanup(func() { response.RawResponse.Body.Close() })

	if response.StatusCode() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode())
	}
}

func TestKeycloakHTTPClientRejectsHostnameMismatch(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(server.Certificate())

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse test server URL: %v", err)
	}
	_, port, err := net.SplitHostPort(serverURL.Host)
	if err != nil {
		t.Fatalf("failed to read test server port: %v", err)
	}
	serverURL.Host = net.JoinHostPort("localhost", port)

	response, err := newKeycloakHTTPClient(rootCAs).R().Get(serverURL.String())
	if err == nil {
		response.RawResponse.Body.Close()
		t.Fatal("expected the Keycloak client to reject a certificate for a different hostname")
	}

	var hostnameError x509.HostnameError
	if !errors.As(err, &hostnameError) {
		t.Fatalf("expected a hostname verification error, got: %v", err)
	}
}
