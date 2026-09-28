package bridge

import (
	"errors"
	"net"
	"strings"
	"syscall"
)

// SSHDiagnosticCode exposes a bounded public category, never raw credential errors.
func SSHDiagnosticCode(err error) string {
	var hostKeyErr *HostKeyError
	if errors.As(err, &hostKeyErr) {
		if hostKeyErr.Kind == hostKeyUnknown {
			return "SSH_HOST_KEY_UNKNOWN"
		}
		return "SSH_HOST_KEY_MISMATCH"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "SSH_CONNECTION_TIMEOUT"
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "SSH_DNS_FAILED"
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return "SSH_CONNECTION_REFUSED"
	}
	message := err.Error()
	if strings.Contains(message, "unable to authenticate") {
		return "SSH_AUTHENTICATION_FAILED"
	}
	if strings.Contains(message, "no common algorithm") {
		return "SSH_ALGORITHM_UNSUPPORTED"
	}
	return "SSH_HANDSHAKE_FAILED"
}
