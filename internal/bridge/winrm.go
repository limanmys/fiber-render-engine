package bridge

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/masterzen/winrm"
)

// InitWinRm creates a new WinRM client and returns it
func InitWinRm(username, password, host, port string, secure bool) (*winrm.Client, error) {
	winrmPort, err := parseWinRMPort(port)
	if err != nil {
		return nil, err
	}
	endpoint := winrm.NewEndpoint(host, winrmPort, secure, true, nil, nil, nil, 0)

	params := winrm.DefaultParameters
	params.TransportDecorator = func() winrm.Transporter {
		return &winrm.ClientNTLM{}
	}

	client, err := winrm.NewClientWithParameters(endpoint, username, password, params)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// VerifyWinRm checks if WinRM authentication is valid to remote end
func VerifyWinRm(username, password, host, port string, secure bool) bool {
	winrmPort, err := parseWinRMPort(port)
	if err != nil {
		return false
	}
	endpoint := winrm.NewEndpoint(host, winrmPort, secure, true, nil, nil, nil, 0)

	params := winrm.DefaultParameters
	params.TransportDecorator = func() winrm.Transporter {
		return &winrm.ClientNTLM{}
	}

	client, err := winrm.NewClientWithParameters(endpoint, username, password, params)
	if err != nil {
		return false
	}

	stdout, _, exitCode, err := client.RunWithContextWithString(context.TODO(), "hostname", "")
	if err != nil || exitCode != 0 {
		return false
	}
	return strings.TrimSpace(stdout) != ""
}

func parseWinRMPort(port string) (int, error) {
	winrmPort, err := strconv.Atoi(port)
	if err != nil {
		return 0, fmt.Errorf("invalid WinRM port: %w", err)
	}
	if winrmPort < 1 || winrmPort > 65535 {
		return 0, fmt.Errorf("invalid WinRM port: %d is outside the TCP port range", winrmPort)
	}

	return winrmPort, nil
}
