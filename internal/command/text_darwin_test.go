//go:build darwin && cgo

package command_test

import (
	"bytes"
	"testing"
)

func TestLocalePrecedenceAndDecoding(t *testing.T) {
	t.Setenv("LC_ALL", "fr_FR.ISO8859-1")
	t.Setenv("LC_CTYPE", "C")
	t.Setenv("LANG", "C")
	run := newInvocation(t, bytes.NewReader([]byte{'c', 'a', 'f', 0xe9}))
	if code := run.command.Run(nil); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	if got := representation(t, run.board, typePlain); !bytes.Equal(got, []byte("café")) {
		t.Fatalf("decoded text=%q, want UTF-8 café", got)
	}
}
