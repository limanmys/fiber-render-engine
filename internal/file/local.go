package file

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/internal/constants"
)

// LocalFile confines file access to the authorized extension, retaining the opened descriptor.
// os.Root enforces the boundary during opening, including symlinks and directory rename races.
func LocalFile(extensionName, localPath, directory string, writable bool) (*os.File, error) {
	return openLocalFile(constants.EXTENSIONS_PATH, extensionName, localPath, directory, writable)
}

func openLocalFile(base, extensionName, localPath, directory string, writable bool) (*os.File, error) {
	denied := func() (*os.File, error) {
		return nil, fiber.NewError(fiber.StatusForbidden, "local file path is not allowed")
	}
	name := strings.ToLower(extensionName)
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return denied()
	}
	if localPath == "" || strings.ContainsRune(localPath, 0) {
		return denied()
	}
	scope := filepath.Join(base, name, directory)
	relative := localPath
	if filepath.IsAbs(localPath) {
		var err error
		relative, err = filepath.Rel(scope, localPath)
		if err != nil {
			return denied()
		}
	}
	if !filepath.IsLocal(relative) || relative == "." {
		return denied()
	}
	extensions, err := os.OpenRoot(base)
	if err != nil {
		return denied()
	}
	defer extensions.Close()
	extension, err := openDirectory(extensions, name)
	if err != nil {
		return denied()
	}
	defer extension.Close()
	root := extension
	if directory != "" {
		if writable && directory == "uploads" {
			if err := extension.Mkdir(directory, 0755); err != nil && !os.IsExist(err) {
				return denied()
			}
		}
		root, err = openDirectory(extension, directory)
		if err != nil {
			return denied()
		}
		defer root.Close()
	}
	flags := os.O_RDONLY | syscall.O_NONBLOCK
	if writable {
		flags = os.O_CREATE | os.O_APPEND | os.O_WRONLY | syscall.O_NONBLOCK
	}
	handle, err := root.OpenFile(relative, flags, 0644)
	if err != nil {
		return denied()
	}
	info, err := handle.Stat()
	if err != nil || !info.Mode().IsRegular() {
		handle.Close()
		return denied()
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); !ok || stat.Nlink != 1 {
		handle.Close()
		return denied()
	}
	return handle, nil
}

// Pin a non-symlink directory and verify that OpenRoot opened that same inode.
// This also rejects a directory exchanged for a symlink between the two opens.
func openDirectory(parent *os.Root, name string) (*os.Root, error) {
	initial, err := parent.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !initial.IsDir() || initial.Mode()&os.ModeSymlink != 0 {
		return nil, fiber.NewError(fiber.StatusForbidden, "local file path is not allowed")
	}
	directory, err := parent.OpenFile(name, os.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	defer directory.Close()
	expected, err := directory.Stat()
	if err != nil {
		return nil, err
	}
	if !os.SameFile(initial, expected) {
		return nil, fiber.NewError(fiber.StatusForbidden, "local file path is not allowed")
	}
	root, err := parent.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	actual, err := root.Stat(".")
	if err != nil || !os.SameFile(expected, actual) {
		root.Close()
		return nil, fiber.NewError(fiber.StatusForbidden, "local file path is not allowed")
	}
	return root, nil
}
