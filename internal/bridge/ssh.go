package bridge

import (
	"net"
	"time"

	"github.com/limanmys/render-engine/pkg/helpers"
	"golang.org/x/crypto/ssh"
)

func InitShellWithPassword(username string, password string, host string, port string) (*ssh.Client, error) {
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

	conn, err := ssh.Dial("tcp", net.JoinHostPort(ipAddress, port), config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func InitShellWithCert(username string, certificate string, host string, port string) (*ssh.Client, error) {
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

	conn, err := ssh.Dial("tcp", net.JoinHostPort(ipAddress, port), config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func VerifySSH(username string, password string, host string, port string) bool {
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

	conn, err := ssh.Dial("tcp", net.JoinHostPort(ipAddress, port), config)
	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}

func VerifySSHCertificate(username string, certificate string, host string, port string) bool {
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

	conn, err := ssh.Dial("tcp", net.JoinHostPort(ipAddress, port), config)
	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}
