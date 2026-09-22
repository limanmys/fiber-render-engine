package linux

import (
	"strings"
	"testing"
)

func TestExecuteReturnsOutput(t *testing.T) {
	output, err := Execute("printf 'ok'")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output != "ok" {
		t.Fatalf("Execute() output = %q, want %q", output, "ok")
	}
}

func TestExecutePreservesStructuredOutputOnCommandFailure(t *testing.T) {
	wantOutput := `{"status":422,"message":"expected failure"}`
	output, err := Execute("printf '" + wantOutput + "'; exit 7")
	if err == nil {
		t.Fatal("Execute() error = nil, want non-zero exit error")
	}
	if output != wantOutput {
		t.Fatalf("Execute() output = %q, want %q", output, wantOutput)
	}
	if !strings.Contains(err.Error(), "exit status 7") {
		t.Fatalf("Execute() error = %q, want exit status", err)
	}
}
