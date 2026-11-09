//go:build darwin && cgo

package command_test

import (
	"bytes"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

func TestExplicitTypesPreserveArbitraryBytes(t *testing.T) {
	t.Setenv("LC_ALL", "C")
	payload := []byte{0xff, 0, 0x7f}
	run := newInvocation(t, bytes.NewReader(payload))
	args := []string{"-t", typeRaw, "-t", typeRaw, optionType, "com.example.alternate"}
	if code := run.command.Run(args); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	for _, typeName := range []string{typeRaw, "com.example.alternate"} {
		if got := representation(t, run.board, typeName); !bytes.Equal(got, payload) {
			t.Fatalf("%s=%v, want unchanged %v", typeName, got, payload)
		}
	}
}

func TestEmptyInputReplacesPreviousContent(t *testing.T) {
	run := newInvocation(t, bytes.NewReader(nil))
	if err := run.board.Write([]pasteboard.Representation{{Type: typePlain, Data: []byte("old content")}}); err != nil {
		t.Fatal(err)
	}
	if code := run.command.Run(nil); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	if got := representation(t, run.board, typePlain); len(got) != 0 {
		t.Fatalf("empty input left content %q", got)
	}
}
