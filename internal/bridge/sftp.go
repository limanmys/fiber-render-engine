package bridge

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/limanmys/render-engine/pkg/logger"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// OpenSFTPConnection creates a new SFTP connection
func OpenSFTPConnection(conn *ssh.Client) *sftp.Client {
	client, err := sftp.NewClient(conn)
	if err != nil {
		logger.Sugar().Errorw("cannot initialize sftp client", "details", err)
	}

	return client
}

// SftpPutFile uploads file on remote path
func (s *Session) SftpPutFile(localPath, remotePath string) error {
	if s.SSH != nil {
		s.SFTP = OpenSFTPConnection(s.SSH)
	} else {
		return errors.New("sftp connection is not alive")
	}

	conn := s.SFTP

	w := s.SFTP.Walk(filepath.Dir(remotePath))
	for w.Step() {
		if w.Err() != nil {
			continue
		}
	}

	f, err := conn.Create(remotePath)
	if err != nil {
		return err
	}
	defer f.Close()

	srcFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(f, srcFile)
	if err != nil {
		return err
	}

	return nil
}

// SftpGetFile downloads file from remote path
func (s *Session) SftpGetFile(localPath, remotePath string) error {
	if s.SSH != nil {
		s.SFTP = OpenSFTPConnection(s.SSH)
	} else {
		return errors.New("sftp connection is not alive")
	}

	conn := s.SFTP

	w := conn.Walk(filepath.Dir(remotePath))
	for w.Step() {
		if w.Err() != nil {
			continue
		}
	}

	f, err := conn.Open(remotePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = os.Stat(localPath)
	if os.IsNotExist(err) {
		_, err = os.Create(localPath)
		if err != nil {
			return err
		}
	}

	srcFile, err := os.OpenFile(localPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(srcFile, f)
	if err != nil {
		return err
	}

	return nil
}
