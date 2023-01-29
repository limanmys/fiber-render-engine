package bridge

import (
	"context"
	"strconv"
	"strings"

	"github.com/masterzen/winrm"
)

// InitWinRm creates a new WinRM client and returns it
func InitWinRm(username, password, host, port string, secure bool) (*winrm.Client, error) {
	winrmPort, _ := strconv.Atoi(port)
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
	winrmPort, _ := strconv.Atoi(port)
	endpoint := winrm.NewEndpoint(host, winrmPort, secure, true, nil, nil, nil, 0)

	params := winrm.DefaultParameters
	params.TransportDecorator = func() winrm.Transporter {
		return &winrm.ClientNTLM{}
	}

	client, err := winrm.NewClientWithParameters(endpoint, username, password, params)
	if err != nil {
		return false
	}

	stdout, _, _, _ := client.RunWithContextWithString(context.TODO(), "hostname", "")
	return strings.TrimSpace(stdout) != ""
}
