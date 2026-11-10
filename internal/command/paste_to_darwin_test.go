//go:build darwin && cgo

package command_test

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

const optionTo = "--to"

func TestPasteToCopiesFileWithoutChangingClipboard(t *testing.T) {
	source := filepath.Join(t.TempDir(), "report ☃\nfile.bin")
	data := []byte{0, 255, 10, 127}
	if err := os.WriteFile(source, data, 0o600); err != nil {
		t.Fatal(err)
	}
	run := newInvocation(t, nil)
	if err := run.board.WriteFiles([]string{source}); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	if code := run.command.Paste([]string{optionTo, destination}); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	assertPastedBytes(t, filepath.Join(destination, filepath.Base(source)), data)
	if run.stdout.Len() != 0 || run.stderr.Len() != 0 {
		t.Fatalf("successful copy emitted stdout=%q stderr=%q", &run.stdout, &run.stderr)
	}
	if code := run.command.Paste([]string{optionFiles, optionNull}); code != 0 || run.stdout.String() != source+"\x00" {
		t.Fatalf("clipboard reference changed: exit=%d stdout=%q stderr=%s", code, &run.stdout, &run.stderr)
	}
}

func assertPastedBytes(t *testing.T, path string, expected []byte) {
	t.Helper()
	actual, err := os.ReadFile(filepath.Clean(path))
	if err != nil || !bytes.Equal(actual, expected) {
		t.Fatalf("file=%q bytes=%q err=%v; want %q", path, actual, err, expected)
	}
}

func TestPasteToCopiesDirectoryTreeAndSymlinks(t *testing.T) {
	const mode = 0o640
	const timestamp = 1700000000
	source := filepath.Join(t.TempDir(), "tree")
	if err := os.MkdirAll(filepath.Join(source, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(source, "nested", "payload")
	if err := os.WriteFile(payload, []byte("directory payload"), mode); err != nil {
		t.Fatal(err)
	}
	modified := time.Unix(timestamp, 0)
	if err := os.Chtimes(payload, modified, modified); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing-target", filepath.Join(source, "dangling")); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(filepath.Dir(source), "tree-link")
	if err := os.Symlink("tree", link); err != nil {
		t.Fatal(err)
	}
	run := newInvocation(t, nil)
	if err := run.board.WriteFiles([]string{source, link}); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	if code := run.command.Paste([]string{optionTo, destination}); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	copied := filepath.Join(destination, "tree", "nested", "payload")
	assertPastedBytes(t, copied, []byte("directory payload"))
	info, err := os.Stat(copied)
	if err != nil || info.Mode().Perm() != mode || !info.ModTime().Equal(modified) {
		t.Fatalf("metadata not preserved: info=%v err=%v", info, err)
	}
	assertPastedSymlink(t, filepath.Join(destination, "tree", "dangling"), "missing-target")
	assertPastedSymlink(t, filepath.Join(destination, "tree-link"), "tree")
}

func assertPastedSymlink(t *testing.T, path, expected string) {
	t.Helper()
	target, err := os.Readlink(path)
	if err != nil || target != expected {
		t.Fatalf("symlink=%q err=%v, want %q", target, err, expected)
	}
}

func TestPasteToRejectsDestinationInsideSourceDirectory(t *testing.T) {
	source := t.TempDir()
	destination := filepath.Join(source, "nested")
	if err := os.Mkdir(destination, 0o700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(destination, alias); err != nil {
		t.Fatal(err)
	}
	run := newInvocation(t, nil)
	if err := run.board.WriteFiles([]string{source}); err != nil {
		t.Fatal(err)
	}
	if code := run.command.Paste([]string{optionTo, alias}); code != 1 {
		t.Fatalf("recursive destination exit=%d stderr=%s", code, &run.stderr)
	}
	entries, err := os.ReadDir(destination)
	if err != nil || len(entries) != 0 {
		t.Fatalf("recursive paste modified destination: %v err=%v", entries, err)
	}
}

func TestPasteToRejectsAmbiguousInvocations(t *testing.T) {
	const destination = "destination"
	for _, args := range [][]string{
		{optionTo, ""},
		{optionTo, destination, optionTo, "another"},
		{optionTo, destination, optionFiles},
		{optionTo, destination, optionNull},
		{optionTo, destination, optionType, typeFileURL},
		{optionTo, destination, optionListTypes},
		{optionTo, destination, "--prefer=txt"},
		{optionTo, destination, optionHelp},
		{optionTo, destination, optionVersion},
	} {
		t.Run(args[len(args)-1], func(t *testing.T) {
			run := newInvocation(t, nil)
			if code := run.command.Paste(args); code != 2 {
				t.Fatalf("args=%q exit=%d stderr=%s", args, code, &run.stderr)
			}
		})
	}
}

func TestPasteToAcceptsFlagLikeDestination(t *testing.T) {
	const destination = "-help"
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, destination), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	run, _ := filePaste(t, []string{firstPasteFile})
	if code := run.command.Paste([]string{optionTo, destination}); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	assertPastedBytes(t, filepath.Join(root, destination, firstPasteFile), []byte("fixture"))
}

func TestPasteToRejectsSpecialFilesWithoutPublishing(t *testing.T) {
	source := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(source, "pipe"), 0o600); err != nil {
		t.Fatal(err)
	}
	run := newInvocation(t, nil)
	if err := run.board.WriteFiles([]string{source}); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	if code := run.command.Paste([]string{optionTo, destination}); code != 1 {
		t.Fatalf("special file exit=%d stderr=%s", code, &run.stderr)
	}
	entries, err := os.ReadDir(destination)
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed copy published data or leaked staging: %v err=%v", entries, err)
	}
}

func TestPasteToCollisionKeepsEarlierCopiesAndStops(t *testing.T) {
	run, paths := filePaste(t, []string{"copied-before-collision", "collision", "later"})
	destination := t.TempDir()
	if err := os.WriteFile(filepath.Join(destination, "collision"), []byte("keep existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code := run.command.Paste([]string{optionTo, destination}); code != 1 || run.stderr.Len() == 0 {
		t.Fatalf("collision exit=%d stderr=%s", code, &run.stderr)
	}
	assertPastedBytes(t, filepath.Join(destination, "copied-before-collision"), []byte("fixture"))
	assertPastedBytes(t, filepath.Join(destination, "collision"), []byte("keep existing"))
	entries, err := os.ReadDir(destination)
	if err != nil || len(entries) != 2 {
		t.Fatalf("copy continued or leaked staging: %v err=%v", entries, err)
	}
	if code := run.command.Paste([]string{optionFiles, optionNull}); code != 0 ||
		run.stdout.String() != paths[0]+"\x00"+paths[1]+"\x00"+paths[2]+"\x00" {
		t.Fatalf("failed paste changed clipboard: exit=%d stdout=%q", code, &run.stdout)
	}
}

func TestPasteToUnreadableSubdirectoryDoesNotPublishTree(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read mode-000 directories")
	}
	source := t.TempDir()
	blocked := filepath.Join(source, "unreadable")
	if err := os.Mkdir(blocked, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(blocked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// #nosec G302 -- Restore traversal permission on this owned directory for cleanup.
		if err := os.Chmod(blocked, 0o700); err != nil {
			t.Error(err)
		}
	})
	run := newInvocation(t, nil)
	if err := run.board.WriteFiles([]string{source}); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	if code := run.command.Paste([]string{optionTo, destination}); code != 1 {
		t.Fatalf("unreadable subtree exit=%d stderr=%s", code, &run.stderr)
	}
	entries, err := os.ReadDir(destination)
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed tree was published or staging leaked: %v err=%v", entries, err)
	}
}

func symlinkParentFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	owned, other := filepath.Join(root, "owned"), filepath.Join(root, "other")
	if err := os.Mkdir(owned, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(other, "child"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(other, "child"), filepath.Join(owned, "link")); err != nil {
		t.Fatal(err)
	}
	return owned, other
}

func TestPasteToResolvesDestinationSymlinksBeforeParentTraversal(t *testing.T) {
	owned, other := symlinkParentFixture(t)
	run, _ := filePaste(t, []string{firstPasteFile})
	if code := run.command.Paste([]string{optionTo, owned + "/link/.."}); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	assertPastedBytes(t, filepath.Join(other, firstPasteFile), []byte("fixture"))
	if _, err := os.Lstat(filepath.Join(owned, firstPasteFile)); !os.IsNotExist(err) {
		t.Fatalf("paste wrote into the lexical parent: %v", err)
	}
}

func TestPasteToResolvesSourceParentsWithoutFollowingFinalSymlink(t *testing.T) {
	owned, other := symlinkParentFixture(t)
	if err := os.WriteFile(filepath.Join(owned, "payload"), []byte("wrong source"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "payload"), []byte("referenced source"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("payload", filepath.Join(other, "final-link")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"payload", "final-link"} {
		t.Run(name, func(t *testing.T) {
			run := newInvocation(t, nil)
			location := &url.URL{Scheme: "file", Path: owned + "/link/../" + name}
			if err := run.board.Write([]pasteboard.Representation{{Type: typeFileURL, Data: []byte(location.String())}}); err != nil {
				t.Fatal(err)
			}
			destination := t.TempDir()
			if code := run.command.Paste([]string{optionTo, destination}); code != 0 {
				t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
			}
			if name == "final-link" {
				assertPastedSymlink(t, filepath.Join(destination, name), "payload")
			} else {
				assertPastedBytes(t, filepath.Join(destination, name), []byte("referenced source"))
			}
		})
	}
}
