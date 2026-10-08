package bridge

import (
	"errors"
	"io"
)

// Put sends file to a remote path
func (s *Session) Put(localFile io.Reader, remotePath string) error {
	if s.SMB != nil {
		err := s.SmbPutFile(localFile, remotePath, s.WindowsLetter)
		if err != nil {
			return err
		}

		return nil
	}

	if s.SFTP != nil {
		err := s.SftpPutFile(localFile, remotePath)
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("cannot found active connection")
}

// Get downloads file to local path
func (s *Session) Get(localFile io.Writer, remotePath string) error {
	if s.SMB != nil {
		err := s.SmbGetFile(localFile, remotePath, s.WindowsLetter)
		if err != nil {
			return err
		}

		return nil
	}

	if s.SFTP != nil {
		err := s.SftpGetFile(localFile, remotePath)
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("cannot found active connection")
}
