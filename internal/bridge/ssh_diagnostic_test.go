package bridge

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"
)

func TestSSHDiagnosticCodeClassifiesWrappedFailures(t *testing.T) {
	tests := []struct {
		err  error
		code string
	}{
		{&HostKeyError{Kind: hostKeyMismatch}, "SSH_HOST_KEY_MISMATCH"},
		{&HostKeyError{Kind: hostKeyUnknown}, "SSH_HOST_KEY_UNKNOWN"},
		{&net.OpError{Op: "read", Err: os.ErrDeadlineExceeded}, "SSH_CONNECTION_TIMEOUT"},
		{&net.DNSError{Err: "no such host", Name: "example.invalid"}, "SSH_DNS_FAILED"},
		{syscall.ECONNREFUSED, "SSH_CONNECTION_REFUSED"},
		{errors.New("ssh: unable to authenticate, attempted methods [none password]"), "SSH_AUTHENTICATION_FAILED"},
		{errors.New("ssh: no common algorithm for host key"), "SSH_ALGORITHM_UNSUPPORTED"},
		{errors.New("unclassified failure with private details"), "SSH_HANDSHAKE_FAILED"},
	}
	for _, test := range tests {
		if got := SSHDiagnosticCode(fmt.Errorf("handshake failed: %w", test.err)); got != test.code {
			t.Errorf("got %s, want %s", got, test.code)
		}
	}
}
