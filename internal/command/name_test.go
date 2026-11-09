package command_test

import (
	"bytes"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/command"
)

func TestCopyVersionIdentifiesExecutable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := command.Command{Stdout: &stdout, Stderr: &stderr, Version: "1.2.3"}
	if code := cmd.Run([]string{"--version"}); code != 0 {
		t.Fatalf("version exit=%d; stderr=%s", code, &stderr)
	}
	if got := stdout.String(); got != "pbjcopy 1.2.3\n" {
		t.Fatalf("copy version identifies the wrong executable: %q", got)
	}
}

func TestPasteVersionIdentifiesExecutable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := command.Command{Stdout: &stdout, Stderr: &stderr, Version: "1.2.3"}
	if code := cmd.Paste([]string{"--version"}); code != 0 {
		t.Fatalf("version exit=%d; stderr=%s", code, &stderr)
	}
	if got := stdout.String(); got != "pbjpaste 1.2.3\n" {
		t.Fatalf("paste version identifies the wrong executable: %q", got)
	}
}
