//go:build darwin && cgo

package command_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/command"
	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

type invocation struct {
	command command.Command
	board   *pasteboard.Board
	stdout  bytes.Buffer
	stderr  bytes.Buffer
}

func newInvocation(t *testing.T, input io.Reader) *invocation {
	t.Helper()
	board, err := pasteboard.Unique()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := board.Close(); err != nil {
			t.Error(err)
		}
	})
	run := &invocation{board: board}
	run.command = command.Command{
		Stdin:  input,
		Stdout: &run.stdout,
		Stderr: &run.stderr,
		OpenPasteboard: func(_ string) (*pasteboard.Board, error) {
			return pasteboard.Open(board.Name())
		},
	}
	return run
}

func representation(t *testing.T, board *pasteboard.Board, typeName string) []byte {
	t.Helper()
	items, err := board.Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d clipboard items, want one", len(items))
	}
	for _, rep := range items[0].Representations {
		if rep.Type == typeName {
			return rep.Data
		}
	}
	t.Fatalf("missing clipboard type %q", typeName)
	return nil
}

func TestHTMLWithPlainTextFallback(t *testing.T) {
	run := newInvocation(t, bytes.NewBufferString("<b>hello</b>"))
	if code := run.command.Run([]string{"-t", typeHTML, optionPlain, "hello"}); code != 0 {
		t.Fatalf("exit=%d; stderr=%s", code, &run.stderr)
	}
	if got := representation(t, run.board, typeHTML); !bytes.Equal(got, []byte("<b>hello</b>")) {
		t.Fatalf("HTML=%q", got)
	}
	if got := representation(t, run.board, typePlain); !bytes.Equal(got, []byte("hello")) {
		t.Fatalf("plain text=%q", got)
	}
	if run.stdout.Len() != 0 || run.stderr.Len() != 0 {
		t.Fatalf("successful copy emitted output: stdout=%q stderr=%q", &run.stdout, &run.stderr)
	}
}
