//go:build darwin && cgo

package command_test

import (
	"bytes"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

func TestPasteExtractsExactBinaryWithoutLocaleConversion(t *testing.T) {
	t.Setenv("LC_ALL", "C")
	run := newInvocation(t, nil)
	payload := []byte{0, 255, 10, 128}
	if err := run.board.Write([]pasteboard.Representation{{Type: "com.example.binary", Data: payload}, {Type: typePlain, Data: []byte("fallback")}}); err != nil {
		t.Fatal(err)
	}
	if code := run.command.Paste([]string{"-t", "com.example.binary"}); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	if !bytes.Equal(run.stdout.Bytes(), payload) {
		t.Fatalf("output=%v", run.stdout.Bytes())
	}
	if got := representation(t, run.board, "com.example.binary"); !bytes.Equal(got, payload) {
		t.Fatalf("clipboard modified: %v", got)
	}
	run.stdout.Reset()
	if code := run.command.Paste([]string{"-t", "public.png"}); code != 1 || run.stdout.Len() != 0 {
		t.Fatalf("missing type: exit=%d stdout=%q", code, &run.stdout)
	}
}
