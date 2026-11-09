//go:build darwin && cgo

package command_test

import (
	"bytes"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"testing/iotest"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

func TestFileReferencesPreserveSelectionAndOrder(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "sample.txt"), []byte("not copied as content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("sample.txt", filepath.Join(directory, "chosen-link")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(directory)
	t.Setenv("HOME", directory)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	run := newInvocation(t, iotest.ErrReader(errors.New("file mode must not read stdin")))
	args := []string{"-f", "chosen-link", optionFile, "./sample.txt", "-f", cwd + "/unused/../sample.txt", "-f", "~/"}
	if code := run.command.Run(args); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	items, err := run.board.Read()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{filepath.Join(cwd, "chosen-link"), filepath.Join(cwd, "sample.txt"), directory}
	if len(items) != len(expected) {
		t.Fatalf("got %d references, want %d", len(items), len(expected))
	}
	for index, path := range expected {
		var actual string
		for _, rep := range items[index].Representations {
			if rep.Type == "public.file-url" {
				location, err := url.Parse(string(rep.Data))
				if err != nil {
					t.Fatal(err)
				}
				actual = filepath.Clean(location.Path)
			}
		}
		if actual != path {
			t.Fatalf("reference %d=%q, want selected path %q", index, actual, path)
		}
	}
}

func TestBrokenFileReferencePreservesClipboard(t *testing.T) {
	directory := t.TempDir()
	valid := filepath.Join(directory, "valid.txt")
	broken := filepath.Join(directory, "broken-link")
	if err := os.WriteFile(valid, []byte("valid"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing-target", broken); err != nil {
		t.Fatal(err)
	}
	run := newInvocation(t, bytes.NewReader(nil))
	if err := run.board.Write([]pasteboard.Representation{{Type: typePlain, Data: []byte("previous selection")}}); err != nil {
		t.Fatal(err)
	}
	if code := run.command.Run([]string{"-f", valid, "-f", broken}); code != 1 {
		t.Fatalf("invalid file batch exit=%d; stderr=%s", code, &run.stderr)
	}
	if got := representation(t, run.board, typePlain); !bytes.Equal(got, []byte("previous selection")) {
		t.Fatalf("invalid file batch replaced clipboard with %q", got)
	}
}
