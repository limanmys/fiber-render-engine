package file

import (
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func RemotePath(serverOS, windowsPath, remotePath string) (string, error) {
	name := filepath.Base(strings.ReplaceAll(remotePath, "\\", "/"))
	if name == "" || name == "." || name == ".." || name == "/" || strings.ContainsRune(name, 0) {
		return "", fiber.NewError(fiber.StatusBadRequest, "invalid remote file name")
	}
	if serverOS == "linux" {
		return "/tmp/" + name, nil
	}
	if strings.ContainsAny(name, `:<>"|?*`) {
		return "", fiber.NewError(fiber.StatusBadRequest, "invalid remote file name")
	}
	return windowsPath + name, nil
}
