package bridge

import (
	"errors"
	"io"
	"net"
	"os"
	"time"

	"github.com/hirochachacha/go-smb2"
)

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

func (s *Session) SmbPutFile(localPath, remotePath, remoteDisk string) error {
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

func (s *Session) SmbGetFile(localPath, remotePath, remoteDisk string) error {
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

	_, err = os.Stat(localPath)
	if os.IsNotExist(err) {
		os.Create(localPath)
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
