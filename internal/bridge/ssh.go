package bridge

import (
	"net"
	"time"

	"github.com/avast/retry-go"
	"github.com/limanmys/render-engine/pkg/helpers"
	"golang.org/x/crypto/ssh"
)

// InitShellWithPassword creates a SSH shell with password
func InitShellWithPassword(username, password, host, port string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 10,
	}

	ipAddress, err := helpers.ResolveIP(host)
	if err != nil {
		return nil, err
	}

	var conn *ssh.Client
	err = retry.Do(
		func() error {
			conn, err = ssh.Dial("tcp", net.JoinHostPort(ipAddress, port), config)
			if err != nil {
				return err
			}
			return nil
		},
		retry.Attempts(5),
		retry.Delay(1*time.Second),
	)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

// InitShellWithCert creates a SSH shell with certificate
func InitShellWithCert(username, certificate, host, port string) (*ssh.Client, error) {
	key, err := ssh.ParsePrivateKey([]byte(certificate))
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 10,
	}

	ipAddress, err := helpers.ResolveIP(host)
	if err != nil {
		return nil, err
	}

	var conn *ssh.Client
	err = retry.Do(
		func() error {
			conn, err = ssh.Dial("tcp", net.JoinHostPort(ipAddress, port), config)
			if err != nil {
				return err
			}
			return nil
		},
		retry.Attempts(5),
		retry.Delay(1*time.Second),
	)

	return conn, nil
}

// VerifySSH checks if remote end is active or not
func VerifySSH(username, password, host, port string) bool {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 10,
	}

	ipAddress, err := helpers.ResolveIP(host)
	if err != nil {
		return false
	}

	var conn *ssh.Client
	err = retry.Do(
		func() error {
			conn, err = ssh.Dial("tcp", net.JoinHostPort(ipAddress, port), config)
			if err != nil {
				return err
			}
			return nil
		},
		retry.Attempts(5),
		retry.Delay(1*time.Second),
	)

	defer conn.Close()
	return true
}

// VerifySSHCertificate checks if remote end is active or not
func VerifySSHCertificate(username, certificate, host, port string) bool {
	key, err := ssh.ParsePrivateKey([]byte(certificate))
	if err != nil {
		return false
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 10,
	}

	ipAddress, err := helpers.ResolveIP(host)
	if err != nil {
		return false
	}

	var conn *ssh.Client
	err = retry.Do(
		func() error {
			conn, err = ssh.Dial("tcp", net.JoinHostPort(ipAddress, port), config)
			if err != nil {
				return err
			}
			return nil
		},
		retry.Attempts(5),
		retry.Delay(1*time.Second),
	)
	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}
