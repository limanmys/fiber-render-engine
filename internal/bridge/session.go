package bridge

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/hirochachacha/go-smb2"
	"github.com/masterzen/winrm"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/text/encoding/unicode"
)

type Session struct {
	SSH            *ssh.Client
	SSHSession     *ssh.Client
	SFTP           *sftp.Client
	SMB            *smb2.Session
	WinRM          *winrm.Client
	Mutex          sync.Mutex
	LastConnection time.Time
	WindowsLetter  string
	WindowsPath    string
	Username       string
	IpAddr         string
	Port           string
	password       string
}

func (s *Session) CloseAllConnections() {
	if s.SSH != nil {
		s.SSH.Close()
	}

	if s.SFTP != nil {
		s.SFTP.Close()
	}

	if s.SMB != nil {
		s.SMB.Logoff()
	}
}

func (val *Session) checkOutput(in io.Writer, output *bytes.Buffer) bool {
	val.Mutex.Lock()
	defer val.Mutex.Unlock()
	if output != nil && output.Len() > 0 && strings.Contains(output.String(), "liman-pass-sudo") {
		in.Write([]byte(val.password + "\n"))
		return true
	}
	return false
}

func (val *Session) Run(command string) (string, error) {
	if val.SSH != nil {
		sess, err := val.SSH.NewSession()
		if err != nil {
			return "", err
		}
		defer sess.Close()
		modes := ssh.TerminalModes{
			ssh.ECHO:          0,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}
		err = sess.RequestPty("dumb", 1000, 1000, modes)
		if err != nil {
			return "", err
		}
		stdoutB := new(bytes.Buffer)
		sess.Stdout = stdoutB
		in, err := sess.StdinPipe()
		if err != nil {
			return "", err
		}
		if strings.Contains(command, "liman-pass-sudo") {
			endChan := make(chan struct{})
			defer close(endChan)
			go func(in io.Writer, output *bytes.Buffer, endChan chan struct{}) {
			For:
				for {
					select {
					case <-time.After(20 * time.Second):
					case <-endChan:
						break For
					default:
						if val.checkOutput(in, output) {
							break For
						}

						time.Sleep(500)
					}
				}
			}(in, stdoutB, endChan)
		}
		sess.Run("(" + command + ") 2> /dev/null")

		tmp := strings.Split(stdoutB.String(), "liman-pass-sudo")
		output := tmp[len(tmp)-1]

		return stripansi.Strip(strings.TrimSpace(output)), nil
	} else if val.WinRM != nil {
		command = "$ProgressPreference = 'SilentlyContinue';" + command
		encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
		encoded, _ := encoder.String(command)
		command = base64.StdEncoding.EncodeToString([]byte(encoded))
		stdout, stderr, _, err := val.WinRM.RunWithString("powershell.exe -encodedCommand "+command, "")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(stdout) + strings.TrimSpace(stderr), nil
	}
	return "", errors.New("cannot run command")
}
