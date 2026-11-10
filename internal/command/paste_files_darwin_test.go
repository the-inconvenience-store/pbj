//go:build darwin && cgo

package command_test

import (
	"bytes"
	"testing"
)

const optionFiles = "--files"

func TestPasteFilesRejectsAmbiguousLinesAndSupportsNUL(t *testing.T) {
	run, paths := filePaste(t, []string{firstPasteFile, "second\nline.txt"})
	if code := run.command.Paste([]string{optionFiles}); code != 1 || run.stdout.Len() != 0 {
		t.Fatalf("unsafe newline output exit=%d stdout=%q", code, &run.stdout)
	}
	run.stderr.Reset()
	if code := run.command.Paste([]string{optionFiles, optionNull}); code != 0 {
		t.Fatalf("NUL extraction exit=%d stderr=%s", code, &run.stderr)
	}
	expected := []byte(paths[0] + "\x00" + paths[1] + "\x00")
	if !bytes.Equal(run.stdout.Bytes(), expected) {
		t.Fatalf("paths=%q want=%q", run.stdout.Bytes(), expected)
	}
	run.stdout.Reset()
	if code := run.command.Paste([]string{optionFiles, optionItem, "0"}); code != 0 || run.stdout.String() != paths[0]+"\n" {
		t.Fatalf("selected path exit=%d stdout=%q", code, &run.stdout)
	}
}
