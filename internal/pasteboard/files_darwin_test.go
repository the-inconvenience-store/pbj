//go:build darwin && cgo

package pasteboard_test

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

func TestPublishFileReferences(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "a file.txt")
	if err := os.WriteFile(file, []byte("file contents are not the clipboard payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	board, err := pasteboard.Unique()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := board.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := board.WriteFiles([]string{file, directory}); err != nil {
		t.Fatal(err)
	}
	items, err := board.Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d file-reference items, want two", len(items))
	}
	for index, expected := range []string{file, directory} {
		location := fileReference(t, items[index])
		if location == nil || location.Scheme != "file" || filepath.Clean(location.Path) != expected {
			t.Fatalf("item %d file URL=%v, want %q", index, location, expected)
		}
	}
}

func fileReference(t *testing.T, item pasteboard.Item) *url.URL {
	t.Helper()
	for _, rep := range item.Representations {
		if rep.Type == "public.file-url" {
			location, err := url.Parse(string(rep.Data))
			if err != nil {
				t.Fatal(err)
			}
			return location
		}
	}
	t.Fatal("missing public.file-url representation")
	return nil
}
