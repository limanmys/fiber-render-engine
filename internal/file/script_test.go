package file

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"al.essio.dev/pkg/shellescape"
)

func TestScriptCommandDoesNotExecuteArgumentOrFilenameSyntax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "task; echo unexpected.sh")
	marker := filepath.Join(dir, "injected")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf '%s\\0' \"$@\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	parameters := "one 'two words' '' $(touch " + shellescape.Quote(marker) + ") ; touch " + shellescape.Quote(marker)
	command, err := ScriptCommand(path, parameters, false)
	if err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("sh", "-c", command).Output()
	if err != nil {
		t.Fatal(err)
	}
	arguments := strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00")
	expected := []string{"one", "two words", "", "$(touch", marker + ")", ";", "touch", marker}
	if !reflect.DeepEqual(arguments, expected) {
		t.Fatalf("arguments changed: %#v", arguments)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("shell argument syntax was executed")
	}
}

func TestScriptCommandRejectsMalformedParameters(t *testing.T) {
	for _, parameters := range []string{"'unclosed", "\"unclosed", "trailing\\", "nul\x00byte"} {
		if _, err := ScriptCommand("/tmp/script.sh", parameters, false); err == nil {
			t.Fatalf("accepted %q", parameters)
		}
	}
}

func TestWindowsScriptCommandKeepsPathsAndExpressionsLiteral(t *testing.T) {
	command, err := ScriptCommand(`C:\Temp\task'name.ps1`, `C:\Temp\input.txt 'two words' 'O''Brien' $(whoami)`, true)
	if err != nil {
		t.Fatal(err)
	}
	expected := `& 'C:\Temp\task''name.ps1' 'C:\Temp\input.txt' 'two words' 'O''Brien' '$(whoami)'`
	if command != expected {
		t.Fatalf("unexpected PowerShell command: %s", command)
	}
}

func TestRemotePathKeepsFilesInTemporaryDirectory(t *testing.T) {
	for _, item := range []struct{ os, input, expected string }{
		{"linux", "../../etc/fixture", "/tmp/fixture"},
		{"windows", `..\..\fixture.txt`, `Temp\fixture.txt`},
	} {
		actual, err := RemotePath(item.os, `Temp\`, item.input)
		if err != nil || actual != item.expected {
			t.Fatalf("unexpected remote path: %q, %v", actual, err)
		}
	}
	for _, name := range []string{"..", ".", "/", "file\x00", "file:stream", `bad"name`} {
		if _, err := RemotePath("windows", `Temp\`, name); err == nil {
			t.Fatalf("invalid Windows file name accepted: %q", name)
		}
	}
}
