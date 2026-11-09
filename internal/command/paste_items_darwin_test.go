//go:build darwin && cgo

package command_test

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func filePaste(t *testing.T, names []string) (*invocation, []string) {
	t.Helper()
	run := newInvocation(t, nil)
	directory := t.TempDir()
	paths := make([]string, len(names))
	for i, name := range names {
		paths[i] = filepath.Join(directory, name)
		if err := os.WriteFile(paths[i], []byte("fixture"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := run.board.WriteFiles(paths); err != nil {
		t.Fatal(err)
	}
	return run, paths
}

const firstPasteFile = "first file.txt"

func TestPasteInspectsTypesPerItem(t *testing.T) {
	run, _ := filePaste(t, []string{firstPasteFile, "second.txt"})
	if code := run.command.Paste([]string{optionListTypes}); code != 0 {
		t.Fatalf("inspect exit=%d stderr=%s", code, &run.stderr)
	}
	var items []struct {
		Item  int      `json:"item"`
		Types []string `json:"types"`
	}
	if err := json.Unmarshal(run.stdout.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Item != 0 || items[1].Item != 1 || !slices.Contains(items[0].Types, typeFileURL) || !slices.Contains(items[1].Types, typeFileURL) {
		t.Fatalf("inspection=%v", items)
	}
}

func TestPasteRequiresExplicitSelectionForMultipleItems(t *testing.T) {
	run, paths := filePaste(t, []string{firstPasteFile, "second.txt"})
	if code := run.command.Paste([]string{"-t", typeFileURL}); code != 1 || run.stdout.Len() != 0 {
		t.Fatalf("ambiguous extraction exit=%d stdout=%q", code, &run.stdout)
	}
	run.stderr.Reset()
	if code := run.command.Paste([]string{"-t", typeFileURL, optionItem, "1"}); code != 0 {
		t.Fatalf("select exit=%d stderr=%s", code, &run.stderr)
	}
	ref, err := url.Parse(run.stdout.String())
	if err != nil || ref.Path != paths[1] {
		t.Fatalf("selected URL=%q err=%v", run.stdout.String(), err)
	}
}
