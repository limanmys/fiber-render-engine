package bridge

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/hirochachacha/go-smb2"
)

// OpenSmbConnection creates a new smb connection session
func OpenSmbConnection(host, username, password string) (*smb2.Session, error) {
	dialer := net.Dialer{Timeout: time.Second * 5}
	conn, err := dialer.Dial("tcp", host+":445")
	if err != nil {
		return nil, err
	}

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     username,
			Password: password,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// SmbPutFile uploads file from local path to remote path
func (s *Session) SmbPutFile(localFile io.Reader, remotePath, remoteDisk string) error {
	if s.SMB == nil {
		return errors.New("smb connection does not exist")
	}

	fs, err := s.SMB.Mount(remoteDisk + "$")
	if err != nil {
		return err
	}
	defer fs.Umount()

	f, err := fs.Create(remotePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, localFile)
	if err != nil {
		return err
	}

	return nil
}

// SmbGetFile downloads file from remote path to local path
func (s *Session) SmbGetFile(localFile io.Writer, remotePath, remoteDisk string) error {
	if s.SMB == nil {
		return errors.New("smb connection does not exist")
	}

	fs, err := s.SMB.Mount(remoteDisk + "$")
	if err != nil {
		return err
	}
	defer fs.Umount()

	f, err := fs.Open(remotePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(localFile, f)
	if err != nil {
		return err
	}

	return nil
}
