package bridge

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
)

func TestLinuxScriptStagingUsesPrivateUniqueDirectories(t *testing.T) {
	session := syntheticFileSession(t)
	existing, err := os.CreateTemp("/tmp", "liman-stage-;echo-*.sh")
	if err != nil {
		t.Fatal(err)
	}
	existingPath := existing.Name()
	t.Cleanup(func() { os.Remove(existingPath) })
	if _, err := existing.WriteString("existing remote file"); err != nil {
		existing.Close()
		t.Fatal(err)
	}
	existing.Close()
	name := filepath.Base(existingPath)
	first, firstCleanup, err := session.StageLinuxScript(strings.NewReader("first script"), name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(firstCleanup)
	if contents, err := os.ReadFile(existingPath); err != nil || string(contents) != "existing remote file" {
		t.Fatal("existing remote file was changed")
	}
	protected := filepath.Join(t.TempDir(), "protected")
	if err := os.WriteFile(protected, []byte("protected contents"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(existingPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(protected, existingPath); err != nil {
		t.Fatal(err)
	}
	second, secondCleanup, err := session.StageLinuxScript(strings.NewReader("second script"), name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(secondCleanup)
	if contents, err := os.ReadFile(protected); err != nil || string(contents) != "protected contents" {
		t.Fatal("upload followed an existing remote symlink")
	}
	if first == second || filepath.Dir(first) == filepath.Dir(second) {
		t.Fatal("script paths were reused")
	}
	for _, path := range []string{first, second} {
		if filepath.Base(path) != name || !strings.HasPrefix(filepath.Dir(path), "/tmp/.liman-script-") {
			t.Fatalf("unexpected staging path: %q", path)
		}
		for _, item := range []string{path, filepath.Dir(path)} {
			info, err := os.Stat(item)
			if err != nil || info.Mode().Perm() != 0700 {
				t.Fatalf("staged file/directory is not private: %v", err)
			}
		}
	}
	contents, err := os.ReadFile(first)
	if err != nil || string(contents) != "first script" {
		t.Fatalf("first script was overwritten: %q, %v", contents, err)
	}
	// Non-recursive cleanup must leave an unexpected file in place.
	extra := filepath.Join(filepath.Dir(first), "preserve-me")
	if err := os.WriteFile(extra, []byte("unrelated file"), 0600); err != nil {
		t.Fatal(err)
	}
	firstCleanup()
	if _, err := os.Stat(first); !os.IsNotExist(err) {
		t.Fatal("script was not cleaned up")
	}
	if contents, err := os.ReadFile(extra); err != nil || string(contents) != "unrelated file" {
		t.Fatal("cleanup affected an unrelated file")
	}
	t.Cleanup(func() { os.Remove(extra); os.Remove(filepath.Dir(first)) })
	secondCleanup()
	if _, err := os.Stat(filepath.Dir(second)); !os.IsNotExist(err) {
		t.Fatal("empty staging directory was not cleaned up")
	}
}

func TestLinuxScriptStagingRejectsPathsAndCleansFailedUpload(t *testing.T) {
	session := syntheticFileSession(t)
	for _, name := range []string{"", ".", "..", "../fixture", "/etc/fixture", `..\fixture`, "nul\x00byte"} {
		if _, cleanup, err := session.StageLinuxScript(strings.NewReader("script"), name); err == nil || cleanup != nil {
			t.Fatalf("invalid name accepted: %q", name)
		}
	}
	stagingDirectories := func() []string {
		t.Helper()
		entries, err := os.ReadDir("/tmp")
		if err != nil {
			t.Fatal(err)
		}
		names := []string{}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".liman-script-") {
				names = append(names, entry.Name())
			}
		}
		return names
	}
	before := stagingDirectories()
	uploadError := errors.New("synthetic upload failure")
	if _, cleanup, err := session.StageLinuxScript(iotest.ErrReader(uploadError), "task.sh"); !errors.Is(err, uploadError) || cleanup != nil {
		t.Fatalf("unexpected failed upload result: %v", err)
	}
	if after := stagingDirectories(); !reflect.DeepEqual(before, after) {
		t.Fatal("failed upload left a staging directory")
	}
}
