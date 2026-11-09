//go:build darwin && cgo

package command_test

import (
	"bytes"
	"testing"
)

func TestDetectsRichInput(t *testing.T) {
	cases := []struct {
		name     string
		typeName string
		data     []byte
	}{
		{name: "PNG", typeName: "public.png", data: []byte("\x89PNG\r\n\x1a\n")},
		{name: "JPEG", typeName: "public.jpeg", data: []byte{0xff, 0xd8, 0xff}},
		{name: "little-endian TIFF", typeName: "public.tiff", data: []byte{'I', 'I', 42, 0}},
		{name: "big-endian TIFF", typeName: "public.tiff", data: []byte{'M', 'M', 0, 42}},
		{name: "PDF", typeName: "com.adobe.pdf", data: []byte("%PDF-1.7\n")},
		{name: "RTF", typeName: "public.rtf", data: []byte("{\\rtf1 hello}")},
		{name: "EPS", typeName: "com.adobe.encapsulated-postscript", data: []byte("%!PS-Adobe-3.0 EPSF-3.0\n")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := newInvocation(t, bytes.NewReader(tc.data))
			if code := run.command.Run(nil); code != 0 {
				t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
			}
			if got := representation(t, run.board, tc.typeName); !bytes.Equal(got, tc.data) {
				t.Fatalf("payload=%q, want %q", got, tc.data)
			}
		})
	}
}
