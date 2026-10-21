package bridge

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"net"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestRawSshHostKeyCachePinsForOneFixedHour(t *testing.T) {
	cache := &rawSshHostKeyCache{}
	first, changed := testSigner(t).PublicKey(), testSigner(t).PublicKey()
	now := time.Now()
	if err := cache.check("EXAMPLE.COM.", 22, first, now); err != nil {
		t.Fatal(err)
	}
	if err := cache.check("example.com", 22, first, now.Add(59*time.Minute)); err != nil {
		t.Fatal(err)
	}
	err := cache.check("example.com", 22, changed, now.Add(time.Hour-time.Nanosecond))
	var hostKeyErr *HostKeyError
	if !errors.As(err, &hostKeyErr) || hostKeyErr.Kind != hostKeyMismatch || hostKeyErr.Fingerprint != ssh.FingerprintSHA256(changed) {
		t.Fatalf("expected changed key to be rejected before expiry, got %v", err)
	}
	if len(cache.trustedKeys("example.com", 22, now.Add(time.Hour))) != 0 {
		t.Fatal("pin lifetime was extended by repeated use")
	}
	if err := cache.check("example.com", 22, changed, now.Add(time.Hour)); err != nil {
		t.Fatalf("new key must be accepted after the fixed hour: %v", err)
	}
	if err := cache.check("example.com", 22, first, now.Add(time.Hour+time.Second)); !IsHostKeyError(err) {
		t.Fatalf("replacement key was not pinned: %v", err)
	}
}

func TestRawSshHostKeyCacheSeparatesEndpointsAndCleansExpiredPins(t *testing.T) {
	cache := &rawSshHostKeyCache{}
	first, other := testSigner(t).PublicKey(), testSigner(t).PublicKey()
	now := time.Now()
	for _, endpoint := range []struct {
		host string
		port int
	}{
		{"example.com", 22}, {"example.com", 2222}, {"other.example.com", 22}, {"[2001:db8::1]", 22},
	} {
		if err := cache.check(endpoint.host, endpoint.port, first, now); err != nil {
			t.Fatal(err)
		}
	}
	if err := cache.check("2001:0db8::1", 22, other, now); !IsHostKeyError(err) {
		t.Fatalf("IPv6 spelling must not bypass the pin: %v", err)
	}
	if err := cache.check("example.com", 2200, other, now); err != nil {
		t.Fatalf("different port must have an independent pin: %v", err)
	}
	cache.trustedKeys("unused.example.com", 22, now.Add(time.Hour))
	if len(cache.keys) != 0 {
		t.Fatal("expired entries were not cleaned up")
	}
}

func TestRawSshHostKeyCacheConcurrentFirstKeyWins(t *testing.T) {
	cache := &rawSshHostKeyCache{}
	keys := []ssh.PublicKey{testSigner(t).PublicKey(), testSigner(t).PublicKey()}
	now := time.Now()
	type result struct {
		key int
		err error
	}
	results := make(chan result, 32)
	for i := range 32 {
		go func() {
			results <- result{i % 2, cache.check("example.com", 22, keys[i%2], now)}
		}()
	}
	winner := -1
	for range 32 {
		result := <-results
		if result.err == nil {
			if winner >= 0 && winner != result.key {
				t.Fatal("conflicting first keys were both accepted")
			}
			winner = result.key
		} else if !IsHostKeyError(result.err) {
			t.Fatal(result.err)
		}
	}
	if winner < 0 {
		t.Fatal("no first key was accepted")
	}
}

// Exercise CreateRaw itself so neither password nor private-key authentication
// accidentally falls back to the database-backed host-key policy.
func TestCreateRawUsesMemoryHostKeysBeforeAuthentication(t *testing.T) {
	for _, connectionType := range []string{"ssh", "ssh_certificate"} {
		t.Run(connectionType, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
			secret := "test-password"
			if connectionType == "ssh_certificate" {
				_, privateKey, err := ed25519.GenerateKey(rand.Reader)
				if err != nil {
					t.Fatal(err)
				}
				block, err := ssh.MarshalPrivateKey(privateKey, "test")
				if err != nil {
					t.Fatal(err)
				}
				secret = string(pem.EncodeToMemory(block))
			}
			first, changed := testSigner(t), testSigner(t)
			var authAttempts atomic.Int32
			done := make(chan struct{})
			go func() {
				defer close(done)
				for _, signer := range []ssh.Signer{first, first, changed} {
					conn, err := listener.Accept()
					if err != nil {
						return
					}
					conn.SetDeadline(time.Now().Add(5 * time.Second))
					config := &ssh.ServerConfig{
						PasswordCallback: func(_ ssh.ConnMetadata, _ []byte) (*ssh.Permissions, error) {
							authAttempts.Add(1)
							return nil, nil
						},
						PublicKeyCallback: func(_ ssh.ConnMetadata, _ ssh.PublicKey) (*ssh.Permissions, error) {
							authAttempts.Add(1)
							return nil, nil
						},
					}
					config.AddHostKey(signer)
					server, _, _, _ := ssh.NewServerConn(conn, config)
					if server != nil {
						server.Close()
					}
					conn.Close()
				}
			}()
			t.Cleanup(func() {
				listener.Close()
				<-done
			})
			for i, want := range []bool{true, true, false} {
				session := &Session{}
				got := session.CreateRaw(connectionType, "test", secret, "127.0.0.1", port)
				if session.SSH != nil {
					session.SSH.Close()
				}
				if got != want {
					t.Fatalf("connection %d: CreateRaw = %v, want %v", i+1, got, want)
				}
			}
			<-done
			if got := authAttempts.Load(); got != 2 {
				t.Fatalf("authentication attempts = %d, want 2; changed host key must fail before authentication", got)
			}
		})
	}
}
