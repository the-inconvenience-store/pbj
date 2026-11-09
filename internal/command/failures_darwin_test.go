//go:build darwin && cgo

package command_test

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

func TestPreparationFailuresPreserveClipboard(t *testing.T) {
	t.Setenv("LC_ALL", "en_US.UTF-8")
	cases := []struct {
		name  string
		input io.Reader
		args  []string
	}{
		{name: "partial read fails", input: io.MultiReader(bytes.NewBufferString("partial"), iotest.ErrReader(errors.New("read failed")))},
		{name: "invalid text", input: bytes.NewReader([]byte{0xff})},
		{name: "explicit plain collision", input: bytes.NewBufferString("original"), args: []string{"-t", typePlain, optionPlain, "replacement"}},
		{name: "detected plain collision", input: bytes.NewBufferString("original"), args: []string{optionPlain, "replacement"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := newInvocation(t, tc.input)
			if err := run.board.Write([]pasteboard.Representation{{Type: typePlain, Data: []byte("keep me")}}); err != nil {
				t.Fatal(err)
			}
			if code := run.command.Run(tc.args); code == 0 {
				t.Fatal("invalid input succeeded")
			}
			if got := representation(t, run.board, typePlain); !bytes.Equal(got, []byte("keep me")) {
				t.Fatalf("previous clipboard changed to %q", got)
			}
		})
	}
}
