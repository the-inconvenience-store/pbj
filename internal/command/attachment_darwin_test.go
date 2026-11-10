//go:build darwin && cgo

package command_test

import (
	"bytes"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"testing/iotest"
)

func attachmentPath(t *testing.T, run *invocation) string {
	t.Helper()
	location, err := url.Parse(string(representation(t, run.board, typeFileURL)))
	if err != nil {
		t.Fatal(err)
	}
	if location.Scheme != "file" || location.Host != "" {
		t.Fatalf("not a local file reference: %s", location)
	}
	return location.Path
}

func TestAttachmentPreservesBytesAndFilename(t *testing.T) {
	for _, data := range [][]byte{{0xff, 0x00, 0x7f}, []byte("\x89PNG\r\n\x1a\n"), {}} {
		t.Run(string(data), func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			run := newInvocation(t, bytes.NewReader(data))
			if code := run.command.Run([]string{optionAttachment, "report ☃.data"}); code != 0 {
				t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
			}
			path := attachmentPath(t, run)
			if filepath.Base(path) != "report ☃.data" {
				t.Fatalf("filename=%q", filepath.Base(path))
			}
			got, err := os.ReadFile(filepath.Clean(path))
			if err != nil || !bytes.Equal(got, data) {
				t.Fatalf("backing bytes=%q err=%v; want %q", got, err, data)
			}
			assertAttachmentMode(t, path, 0o600)
			assertAttachmentMode(t, filepath.Dir(path), 0o700)
		})
	}
}

func assertAttachmentMode(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != mode {
		t.Fatalf("permissions=%o, want %o", info.Mode().Perm(), mode)
	}
}

func TestAttachmentRejectsInvalidInvocations(t *testing.T) {
	const name = "one"
	for _, args := range [][]string{
		{optionAttachment, ""},
		{optionAttachment, "."},
		{optionAttachment, ".."},
		{optionAttachment, "../escape"},
		{optionAttachment, "/absolute"},
		{optionAttachment, "folder/file"},
		{optionAttachment, "bad\\name"},
		{optionAttachment, "bad:name"},
		{optionAttachment, "bad\x00name"},
		{optionAttachment, "bad\nname"},
		{optionAttachment, "\xff"},
		{optionAttachment, name, optionAttachment, "two"},
		{optionAttachment, name, optionFile, "existing"},
		{optionAttachment, name, optionType, typeHTML},
		{optionAttachment, name, optionPlain, "fallback"},
		{optionAttachment, name, optionHelp},
		{optionAttachment, name, optionVersion},
		{optionAttachment, name, optionListTypes},
	} {
		t.Run(args[1], func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			run := newInvocation(t, iotest.ErrReader(errors.New("must validate before reading")))
			if code := run.command.Run(args); code != 2 {
				t.Fatalf("args=%q exit=%d, want usage failure; stderr=%s", args, code, &run.stderr)
			}
		})
	}
}

func TestAttachmentPreparationFailurePreservesClipboardAndFiles(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	run := newInvocation(t, bytes.NewBufferString("previous attachment"))
	if code := run.command.Run([]string{optionAttachment, "report.bin"}); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	previous := attachmentPath(t, run)
	run.command.Stdin = io.MultiReader(bytes.NewBufferString("partial"), iotest.ErrReader(errors.New("read failed")))
	if code := run.command.Run([]string{optionAttachment, "report.bin"}); code != 1 {
		t.Fatalf("exit=%d, want preparation failure; stderr=%s", code, &run.stderr)
	}
	if got := attachmentPath(t, run); got != previous {
		t.Fatalf("clipboard reference changed from %q to %q", previous, got)
	}
	got, err := os.ReadFile(filepath.Clean(previous))
	if err != nil || string(got) != "previous attachment" {
		t.Fatalf("previous attachment changed: bytes=%q err=%v", got, err)
	}
	entries, err := os.ReadDir(filepath.Dir(filepath.Dir(previous)))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(filepath.Dir(previous)) {
		t.Fatalf("failed preparation leaked backing files: %v", entries)
	}
}

func TestAttachmentsWithSameNameDoNotOverwrite(t *testing.T) {
	const name = "-help"
	t.Setenv("HOME", t.TempDir())
	run := newInvocation(t, bytes.NewBufferString("first"))
	if code := run.command.Run([]string{optionAttachment, name}); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	first := attachmentPath(t, run)
	run.command.Stdin = bytes.NewBufferString("second")
	if code := run.command.Run([]string{optionAttachment, name}); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	second := attachmentPath(t, run)
	if first == second || filepath.Base(second) != name {
		t.Fatalf("attachment references=%q, %q", first, second)
	}
	for path, expected := range map[string]string{first: "first", second: "second"} {
		got, err := os.ReadFile(filepath.Clean(path))
		if err != nil || string(got) != expected {
			t.Fatalf("backing bytes=%q err=%v, want %q", got, err, expected)
		}
	}
}
