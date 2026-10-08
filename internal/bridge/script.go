package bridge

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	"github.com/pkg/sftp"
)

// StageLinuxScript uploads into a private directory without replacing existing files.
// Cleanup removes only the uploaded file and its empty directory over SFTP.
func (s *Session) StageLinuxScript(source io.Reader, name string) (string, func(), error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return "", nil, errors.New("invalid script file name")
	}
	if s.SSH == nil {
		return "", nil, errors.New("ssh connection is not alive")
	}
	client, err := sftp.NewClient(s.SSH)
	if err != nil {
		return "", nil, err
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		client.Close()
		return "", nil, err
	}
	directory := "/tmp/.liman-script-" + hex.EncodeToString(random[:])
	if err := client.Mkdir(directory); err != nil {
		client.Close()
		return "", nil, err
	}
	remotePath := path.Join(directory, name)
	created := false
	cleanup := func() {
		if created {
			client.Remove(remotePath)
		}
		client.RemoveDirectory(directory)
		client.Close()
	}
	if err := client.Chmod(directory, 0700); err != nil {
		cleanup()
		return "", nil, err
	}
	file, err := client.OpenFile(remotePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	created = true
	_, copyErr := io.Copy(file, source)
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil {
		cleanup()
		return "", nil, errors.Join(copyErr, closeErr)
	}
	if err := client.Chmod(remotePath, 0700); err != nil {
		cleanup()
		return "", nil, err
	}
	return remotePath, cleanup, nil
}
