package file

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func localFixture(t *testing.T) (string, string) {
	t.Helper()
	base := t.TempDir()
	for _, dir := range []string{"alpha/uploads", "alpha/scripts", "alpha/views", "beta/uploads"} {
		if err := os.MkdirAll(filepath.Join(base, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"alpha/uploads/input.txt", "alpha/scripts/test.sh", "alpha/views/functions.php", "beta/uploads/input.txt", "private.txt"} {
		if err := os.WriteFile(filepath.Join(base, name), []byte("synthetic fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return base, filepath.Join(base, "alpha")
}

func TestLocalFileAllowsExtensionFilesAndScopedDownloads(t *testing.T) {
	base, extension := localFixture(t)
	cases := []struct {
		path, directory string
		write           bool
	}{
		{filepath.Join(extension, "uploads/input.txt"), "", false},
		{"uploads/input.txt", "", false},
		{filepath.Join(extension, "scripts/test.sh"), "scripts", false},
		{filepath.Join(extension, "uploads/new.txt"), "uploads", true},
	}
	for _, item := range cases {
		handle, err := openLocalFile(base, "Alpha", item.path, item.directory, item.write)
		if err != nil {
			t.Fatalf("allowed file rejected: %v", err)
		}
		handle.Close()
	}
}

func TestLocalFileRejectsEscapesAndCodeWrites(t *testing.T) {
	base, extension := localFixture(t)
	if err := os.Symlink(filepath.Join(base, "private.txt"), filepath.Join(extension, "uploads/absolute-link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../../private.txt", filepath.Join(extension, "uploads/relative-link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../views/functions.php", filepath.Join(extension, "uploads/code-link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(filepath.Join(extension, "views/functions.php"), filepath.Join(extension, "uploads/code-hardlink")); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(extension, "uploads/pipe"), 0600); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, path, directory string
		write                 bool
	}{
		{"absolute host path", filepath.Join(base, "private.txt"), "", false},
		{"other extension", filepath.Join(base, "beta/uploads/input.txt"), "", false},
		{"prefix collision", extension + "-other/input.txt", "", false},
		{"relative traversal", "../private.txt", "", false},
		{"absolute traversal", extension + "/../private.txt", "", false},
		{"absolute symlink", "uploads/absolute-link", "", false},
		{"relative symlink", "uploads/relative-link", "", false},
		{"script outside scripts", filepath.Join(extension, "uploads/input.txt"), "scripts", false},
		{"download to code", filepath.Join(extension, "views/functions.php"), "uploads", true},
		{"download through code symlink", "code-link", "uploads", true},
		{"download through code hardlink", "code-hardlink", "uploads", true},
		{"download traversal", "../views/functions.php", "uploads", true},
		{"directory", "uploads", "", false},
		{"fifo", "uploads/pipe", "", false},
		{"nul", "input\x00.txt", "", false},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			handle, err := openLocalFile(base, "alpha", item.path, item.directory, item.write)
			if handle != nil {
				handle.Close()
				t.Fatal("forbidden file was opened")
			}
			var denial *fiber.Error
			if !errors.As(err, &denial) || denial.Code != fiber.StatusForbidden {
				t.Fatalf("expected 403, got %v", err)
			}
		})
	}
	contents, err := os.ReadFile(filepath.Join(extension, "views/functions.php"))
	if err != nil || string(contents) != "synthetic fixture" {
		t.Fatal("code was modified")
	}
}

func TestDownloadCreatesUploadsAndPreservesAppendBehavior(t *testing.T) {
	base := t.TempDir()
	if err := os.Mkdir(filepath.Join(base, "alpha"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, data := range []string{"first", "second"} {
		handle, err := openLocalFile(base, "alpha", "download.txt", "uploads", true)
		if err != nil {
			t.Fatal(err)
		}
		_, err = handle.WriteString(data)
		handle.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(base, "alpha/uploads/download.txt"))
	if err != nil || string(data) != "firstsecond" {
		t.Fatalf("download append changed: %q, %v", data, err)
	}
}

func TestLocalFileRejectsSymlinkScopeDirectories(t *testing.T) {
	base, extension := localFixture(t)
	if err := os.Symlink("alpha", filepath.Join(base, "linked")); err != nil {
		t.Fatal(err)
	}
	if handle, err := openLocalFile(base, "linked", "uploads/input.txt", "", false); err == nil {
		handle.Close()
		t.Fatal("symlink extension accepted")
	}
	if err := os.Rename(filepath.Join(extension, "uploads"), filepath.Join(extension, "old-uploads")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("views", filepath.Join(extension, "uploads")); err != nil {
		t.Fatal(err)
	}
	if handle, err := openLocalFile(base, "alpha", "functions.php", "uploads", true); err == nil {
		handle.Close()
		t.Fatal("symlink upload directory accepted")
	}
}

func TestOpenedLocalFileIsNotReopenedAfterPathReplacement(t *testing.T) {
	base, extension := localFixture(t)
	path := filepath.Join(extension, "uploads/input.txt")
	handle, err := openLocalFile(base, "alpha", path, "", false)
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Close()
	if err := os.Rename(path, path+".old"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(base, "private.txt"), path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "private.txt"), []byte("outside fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	contents, err := io.ReadAll(handle)
	if err != nil || string(contents) != "synthetic fixture" {
		t.Fatalf("opened file changed: %q, %v", contents, err)
	}
}
