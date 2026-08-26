package bridge

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/limanmys/render-engine/app/models"
	"golang.org/x/crypto/ssh"
)

func testSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	return signer
}

func testECDSASigner(t *testing.T) ssh.Signer {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	return signer
}

func testRSASigner(t *testing.T) ssh.Signer {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	return signer
}

func trustedKey(key ssh.PublicKey) models.SshHostKey {
	return models.SshHostKey{
		PublicKey: strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key))),
	}
}

func TestHostKeyCallbackAcceptsApprovedKey(t *testing.T) {
	signer := testSigner(t)
	callback := newHostKeyCallback("EXAMPLE.COM.", 22, func(host string, port int) ([]models.SshHostKey, error) {
		if host != "example.com" || port != 22 {
			t.Fatalf("unexpected endpoint: %s:%d", host, port)
		}
		return []models.SshHostKey{trustedKey(signer.PublicKey())}, nil
	})

	if err := callback("ignored", nil, signer.PublicKey()); err != nil {
		t.Fatalf("approved key was rejected: %v", err)
	}
}

func TestHostKeyCallbackRejectsUnknownAndChangedKeys(t *testing.T) {
	presented := testSigner(t)
	trusted := testSigner(t)

	tests := []struct {
		name string
		keys []models.SshHostKey
	}{
		{name: "unknown endpoint", keys: nil},
		{name: "changed key", keys: []models.SshHostKey{trustedKey(trusted.PublicKey())}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			callback := newHostKeyCallback("192.0.2.1", 2222, func(string, int) ([]models.SshHostKey, error) {
				return test.keys, nil
			})
			err := callback("ignored", nil, presented.PublicKey())
			if !IsHostKeyError(err) {
				t.Fatalf("expected HostKeyError, got %v", err)
			}
		})
	}
}

func TestHostKeyCallbackFailsClosedOnStoreErrorsAndMalformedKeys(t *testing.T) {
	presented := testSigner(t)
	storeErr := errors.New("database unavailable")

	callback := newHostKeyCallback("example.com", 22, func(string, int) ([]models.SshHostKey, error) {
		return nil, storeErr
	})
	if err := callback("ignored", nil, presented.PublicKey()); !errors.Is(err, storeErr) {
		t.Fatalf("expected database error, got %v", err)
	}

	callback = newHostKeyCallback("example.com", 22, func(string, int) ([]models.SshHostKey, error) {
		return []models.SshHostKey{{PublicKey: "not-a-public-key"}}, nil
	})
	if err := callback("ignored", nil, presented.PublicKey()); err == nil {
		t.Fatal("malformed trusted key must fail closed")
	}
}

func TestTrustedHostKeyAlgorithmsExpandRSAAndPreserveCallerRestrictions(t *testing.T) {
	ed25519Signer := testSigner(t)
	ecdsaSigner := testECDSASigner(t)
	rsaSigner := testRSASigner(t)

	algorithms, err := trustedHostKeyAlgorithms([]models.SshHostKey{
		trustedKey(ed25519Signer.PublicKey()),
		trustedKey(ecdsaSigner.PublicKey()),
		trustedKey(rsaSigner.PublicKey()),
		trustedKey(ed25519Signer.PublicKey()),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		ssh.KeyAlgoED25519,
		ssh.KeyAlgoECDSA256,
		ssh.KeyAlgoRSASHA256,
		ssh.KeyAlgoRSASHA512,
		ssh.KeyAlgoRSA,
	}
	if strings.Join(algorithms, ",") != strings.Join(want, ",") {
		t.Fatalf("algorithms = %v, want %v", algorithms, want)
	}

	configured := []string{ssh.KeyAlgoRSASHA512, ssh.KeyAlgoED25519, "unsupported"}
	intersection := intersectHostKeyAlgorithms(algorithms, configured)
	wantIntersection := []string{ssh.KeyAlgoRSASHA512, ssh.KeyAlgoED25519}
	if strings.Join(intersection, ",") != strings.Join(wantIntersection, ",") {
		t.Fatalf("intersection = %v, want %v", intersection, wantIntersection)
	}
}

func TestDialSshAuthenticatesOnlyAfterHostKeyApproval(t *testing.T) {
	serverSigner := testSigner(t)
	otherSigner := testSigner(t)

	tests := []struct {
		name             string
		trusted          ssh.PublicKey
		wantConnection   bool
		wantAuthAttempt  bool
		wantHostKeyError bool
	}{
		{
			name:            "approved key",
			trusted:         serverSigner.PublicKey(),
			wantConnection:  true,
			wantAuthAttempt: true,
		},
		{
			name:             "changed key",
			trusted:          otherSigner.PublicKey(),
			wantAuthAttempt:  false,
			wantHostKeyError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()

			var authAttempted atomic.Bool
			serverConfig := &ssh.ServerConfig{
				PasswordCallback: func(_ ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
					authAttempted.Store(true)
					if string(password) != "secret" {
						return nil, errors.New("invalid password")
					}
					return nil, nil
				},
			}
			serverConfig.AddHostKey(serverSigner)

			serverDone := make(chan struct{})
			go func() {
				defer close(serverDone)
				connection, acceptErr := listener.Accept()
				if acceptErr != nil {
					return
				}
				defer connection.Close()
				serverConnection, _, _, _ := ssh.NewServerConn(connection, serverConfig)
				if serverConnection != nil {
					serverConnection.Close()
				}
			}()

			port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
			client, err := dialSshWithHostKeyLoader(
				"127.0.0.1",
				port,
				&ssh.ClientConfig{
					User: "test",
					Auth: []ssh.AuthMethod{ssh.Password("secret")},
				},
				func(string, int) ([]models.SshHostKey, error) {
					return []models.SshHostKey{trustedKey(test.trusted)}, nil
				},
			)
			if client != nil {
				client.Close()
			}
			<-serverDone

			if (err == nil) != test.wantConnection {
				t.Fatalf("unexpected connection result: %v", err)
			}
			if authAttempted.Load() != test.wantAuthAttempt {
				t.Fatalf("authentication attempted = %v, want %v", authAttempted.Load(), test.wantAuthAttempt)
			}
			if test.wantHostKeyError && !IsHostKeyError(err) {
				t.Fatalf("expected host key error, got %v", err)
			}
		})
	}
}

func TestDialSshNegotiatesTheApprovedHostKeyAlgorithm(t *testing.T) {
	ecdsaSigner := testECDSASigner(t)
	ed25519Signer := testSigner(t)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	serverConfig := &ssh.ServerConfig{
		PasswordCallback: func(_ ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			if string(password) != "secret" {
				return nil, errors.New("invalid password")
			}
			return nil, nil
		},
	}
	serverConfig.AddHostKey(ecdsaSigner)
	serverConfig.AddHostKey(ed25519Signer)

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		for range 2 {
			connection, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}
			serverConnection, _, _, _ := ssh.NewServerConn(connection, serverConfig)
			if serverConnection != nil {
				serverConnection.Close()
			}
			connection.Close()
		}
	}()

	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	clientConfig := func() *ssh.ClientConfig {
		return &ssh.ClientConfig{
			User: "test",
			Auth: []ssh.AuthMethod{ssh.Password("secret")},
		}
	}

	client, err := dialSshWithHostKeyLoader(
		"127.0.0.1",
		port,
		clientConfig(),
		func(string, int) ([]models.SshHostKey, error) { return nil, nil },
	)
	if client != nil {
		client.Close()
	}
	var hostKeyErr *HostKeyError
	if !errors.As(err, &hostKeyErr) {
		t.Fatalf("expected untrusted default algorithm to fail, got %v", err)
	}
	if hostKeyErr.Fingerprint != ssh.FingerprintSHA256(ecdsaSigner.PublicKey()) {
		t.Fatalf("default negotiation did not select ECDSA: %s", hostKeyErr.Fingerprint)
	}

	var loaderCalls atomic.Int32
	client, err = dialSshWithHostKeyLoader(
		"127.0.0.1",
		port,
		clientConfig(),
		func(string, int) ([]models.SshHostKey, error) {
			loaderCalls.Add(1)
			return []models.SshHostKey{trustedKey(ed25519Signer.PublicKey())}, nil
		},
	)
	if err != nil {
		t.Fatalf("approved ED25519 key was not negotiated: %v", err)
	}
	client.Close()
	if loaderCalls.Load() != 1 {
		t.Fatalf("trusted key loader called %d times, want 1", loaderCalls.Load())
	}
	<-serverDone
}
