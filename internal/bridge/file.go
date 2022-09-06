package bridge

import "errors"

func (s *Session) Put(localPath, remotePath string) error {
	if s.SFTP != nil {
		err := s.SftpPutFile(localPath, remotePath)
		if err != nil {
			return err
		}

		return nil
	} else if s.SMB != nil {
		err := s.SmbPutFile(localPath, remotePath, s.WindowsLetter)
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("cannot found active connection")
}

func (s *Session) Get(localPath, remotePath string) error {
	if s.SFTP != nil {
		err := s.SftpGetFile(localPath, remotePath)
		if err != nil {
			return err
		}

		return nil
	} else if s.SMB != nil {
		err := s.SmbGetFile(localPath, remotePath, s.WindowsLetter)
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("cannot found active connection")
}
