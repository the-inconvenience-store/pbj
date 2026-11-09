//go:build darwin && cgo

package pasteboard_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

func TestFileReferencesSurviveWriterExit(t *testing.T) {
	if name := os.Getenv("PBJC_TEST_WRITE_BOARD"); name != "" {
		writeFileReferencesInChild(t, name)
		return
	}
	directory := t.TempDir()
	file := filepath.Join(directory, "sample.txt")
	if err := os.WriteFile(file, []byte("generated fixture"), 0o600); err != nil {
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
	paths := []string{file, directory}
	encoded, err := json.Marshal(paths)
	if err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	// #nosec G204 -- The kernel supplies this test binary's path via os.Executable, not argv or user input.
	writer := exec.Command(executable, "-test.run=^TestFileReferencesSurviveWriterExit$")
	writer.Env = append(os.Environ(), "PBJC_TEST_WRITE_BOARD="+board.Name(), "PBJC_TEST_WRITE_FILES="+string(encoded))
	if output, err := writer.CombinedOutput(); err != nil {
		t.Fatalf("writer failed: %v\n%s", err, output)
	}
	items, err := board.Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != len(paths) {
		t.Fatalf("after writer exited, got %d references, want %d", len(items), len(paths))
	}
	for index, path := range paths {
		location := fileReference(t, items[index])
		if filepath.Clean(location.Path) != path {
			t.Fatalf("reference %d path=%q, want %q", index, location.Path, path)
		}
	}
}

func writeFileReferencesInChild(t *testing.T, name string) {
	t.Helper()
	var paths []string
	if err := json.Unmarshal([]byte(os.Getenv("PBJC_TEST_WRITE_FILES")), &paths); err != nil {
		t.Fatal(err)
	}
	board, err := pasteboard.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := board.WriteFiles(paths); err != nil {
		t.Fatal(err)
	}
	if err := board.Close(); err != nil {
		t.Fatal(err)
	}
}
