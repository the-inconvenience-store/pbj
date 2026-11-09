package command_test

import (
	"bytes"
	"errors"
	"testing"
	"testing/iotest"

	"github.com/the-inconvenience-store/pbj/internal/command"
)

func TestRejectsOperandBeforeReadingInput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := command.Command{
		Stdin:  iotest.ErrReader(errors.New("stdin was read")),
		Stdout: &stdout,
		Stderr: &stderr,
	}
	if code := cmd.Run([]string{"image.png"}); code != 2 {
		t.Fatalf("operand invocation exited %d, want usage failure 2; stderr=%s", code, &stderr)
	}
	if stdout.Len() != 0 {
		t.Fatalf("usage error wrote to stdout: %q", &stdout)
	}
}
