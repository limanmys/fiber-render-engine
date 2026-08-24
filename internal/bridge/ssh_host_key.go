package bridge

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/helpers"
	"golang.org/x/crypto/ssh"
)

const (
	hostKeyUnknown  = "unknown"
	hostKeyMismatch = "mismatch"
)

type hostKeyLoader func(host string, port int) ([]models.SshHostKey, error)

// HostKeyError indicates that the remote identity is absent from or conflicts with the trust store.
type HostKeyError struct {
	Kind        string
	Host        string
	Port        int
	Fingerprint string
}

func (e *HostKeyError) Error() string {
	if e.Kind == hostKeyUnknown {
		return fmt.Sprintf("ssh host key is not trusted for %s", net.JoinHostPort(e.Host, strconv.Itoa(e.Port)))
	}

	return fmt.Sprintf("ssh host key mismatch for %s", net.JoinHostPort(e.Host, strconv.Itoa(e.Port)))
}

// IsHostKeyError reports whether an SSH handshake failed because host identity validation failed.
func IsHostKeyError(err error) bool {
	var hostKeyErr *HostKeyError
	return errors.As(err, &hostKeyErr)
}

func normalizeSshHost(host string) string {
	normalized := strings.Trim(strings.TrimSpace(host), "[]")
	if ip := net.ParseIP(normalized); ip != nil {
		return ip.String()
	}

	return strings.TrimSuffix(strings.ToLower(normalized), ".")
}

func parseSshPort(port string) (int, error) {
	parsed, err := strconv.Atoi(port)
	if err != nil || parsed < 1 || parsed > 65535 {
		return 0, errors.New("invalid ssh port")
	}

	return parsed, nil
}

func newHostKeyCallback(host string, port int, loader hostKeyLoader) ssh.HostKeyCallback {
	normalizedHost := normalizeSshHost(host)

	return func(_ string, _ net.Addr, presented ssh.PublicKey) error {
		trustedKeys, err := loader(normalizedHost, port)
		if err != nil {
			return fmt.Errorf("cannot load trusted ssh host keys: %w", err)
		}

		if len(trustedKeys) == 0 {
			return &HostKeyError{
				Kind:        hostKeyUnknown,
				Host:        normalizedHost,
				Port:        port,
				Fingerprint: ssh.FingerprintSHA256(presented),
			}
		}

		for _, trusted := range trustedKeys {
			parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(trusted.PublicKey))
			if err != nil {
				return fmt.Errorf("trusted ssh host key is invalid: %w", err)
			}

			if bytes.Equal(parsed.Marshal(), presented.Marshal()) {
				return nil
			}
		}

		return &HostKeyError{
			Kind:        hostKeyMismatch,
			Host:        normalizedHost,
			Port:        port,
			Fingerprint: ssh.FingerprintSHA256(presented),
		}
	}
}

func dialSsh(host, port string, config *ssh.ClientConfig) (*ssh.Client, error) {
	return dialSshWithHostKeyLoader(host, port, config, liman.GetTrustedSshHostKeys)
}

func dialSshWithHostKeyLoader(host, port string, config *ssh.ClientConfig, loader hostKeyLoader) (*ssh.Client, error) {
	parsedPort, err := parseSshPort(port)
	if err != nil {
		return nil, err
	}
	config.HostKeyCallback = newHostKeyCallback(host, parsedPort, loader)

	normalizedHost := normalizeSshHost(host)
	resolvedIP, err := helpers.ResolveIP(normalizedHost)
	if err != nil {
		return nil, err
	}

	rawConnection, err := net.DialTimeout("tcp", net.JoinHostPort(resolvedIP, port), config.Timeout)
	if err != nil {
		return nil, err
	}

	clientConnection, channels, requests, err := ssh.NewClientConn(
		rawConnection,
		net.JoinHostPort(normalizedHost, port),
		config,
	)
	if err != nil {
		rawConnection.Close()
		return nil, err
	}

	return ssh.NewClient(clientConnection, channels, requests), nil
}
