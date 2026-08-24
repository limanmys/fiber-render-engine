package bridge

import (
	"strings"
	"time"

	"github.com/avast/retry-go"
	"golang.org/x/crypto/ssh"
)

// InitShellWithPassword creates a SSH shell with password
func InitShellWithPassword(username, password, host, port string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout: time.Second * 5,
	}

	var conn *ssh.Client
	var err error
	err = retry.Do(
		func() error {
			conn, err = dialSsh(host, port, config)
			if err != nil {
				if strings.Contains(err.Error(), "unable to authenticate") || IsHostKeyError(err) {
					return retry.Unrecoverable(err)
				}
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
		Timeout: time.Second * 5,
	}

	var conn *ssh.Client
	err = retry.Do(
		func() error {
			conn, err = dialSsh(host, port, config)
			if err != nil {
				if strings.Contains(err.Error(), "unable to authenticate") || IsHostKeyError(err) {
					return retry.Unrecoverable(err)
				}
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

// VerifySSH checks if remote end is active or not
func VerifySSH(username, password, host, port string) bool {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout: time.Second * 5,
	}

	var conn *ssh.Client
	var err error
	err = retry.Do(
		func() error {
			conn, err = dialSsh(host, port, config)
			if err != nil {
				if strings.Contains(err.Error(), "unable to authenticate") || IsHostKeyError(err) {
					return retry.Unrecoverable(err)
				}
				return err
			}
			return nil
		},
		retry.Attempts(5),
		retry.Delay(1*time.Second),
	)

	if err != nil || conn == nil {
		return false
	}

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
		Timeout: time.Second * 5,
	}

	var conn *ssh.Client
	err = retry.Do(
		func() error {
			conn, err = dialSsh(host, port, config)
			if err != nil {
				if strings.Contains(err.Error(), "unable to authenticate") || IsHostKeyError(err) {
					return retry.Unrecoverable(err)
				}
				return err
			}
			return nil
		},
		retry.Attempts(5),
		retry.Delay(1*time.Second),
	)
	if err != nil || conn == nil {
		return false
	}

	defer conn.Close()
	return true
}
