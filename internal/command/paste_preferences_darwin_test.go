//go:build darwin && cgo

package command_test

import (
	"bytes"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

func TestPasteDefaultsPreferencesAndLocale(t *testing.T) {
	t.Setenv("LC_ALL", "fr_FR.ISO8859-1")
	t.Setenv("LC_CTYPE", "C")
	t.Setenv("LANG", "C")
	for _, test := range []struct {
		name string
		args []string
		want []byte
	}{
		{"default", nil, []byte{'c', 'a', 'f', 0xe9}},
		{"rtf", []string{"-Prefer", "rtf"}, []byte(`{\rtf1 rich}`)},
		{"postscript", []string{"--prefer", "ps"}, []byte("%!PS-Adobe-3.0\n")},
	} {
		t.Run(test.name, func(t *testing.T) {
			run := newInvocation(t, nil)
			if err := run.board.Write([]pasteboard.Representation{
				{Type: typePlain, Data: []byte("café")},
				{Type: "public.rtf", Data: []byte(`{\rtf1 rich}`)},
				{Type: "com.adobe.encapsulated-postscript", Data: []byte("%!PS-Adobe-3.0\n")},
			}); err != nil {
				t.Fatal(err)
			}
			if code := run.command.Paste(test.args); code != 0 || !bytes.Equal(run.stdout.Bytes(), test.want) {
				t.Fatalf("exit=%d stdout=%v stderr=%s", code, run.stdout.Bytes(), &run.stderr)
			}
		})
	}
}
