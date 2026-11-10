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

func TestDetectsJSONContainers(t *testing.T) {
	for _, input := range []string{`{"answer":42}`, " \n[true, null, \"☃\"]\t"} {
		t.Run(input, func(t *testing.T) {
			t.Setenv("LC_ALL", "C")
			run := newInvocation(t, bytes.NewBufferString(input))
			if code := run.command.Run(nil); code != 0 {
				t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
			}
			if got := representation(t, run.board, "public.json"); string(got) != input {
				t.Fatalf("JSON=%q, want %q", got, input)
			}
		})
	}
}

func TestDetectsDeclaredXML(t *testing.T) {
	input := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<message>☃</message>\n"
	t.Setenv("LC_ALL", "C")
	run := newInvocation(t, bytes.NewBufferString(input))
	if code := run.command.Run([]string{"--plain", "snowman"}); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
	}
	if got := representation(t, run.board, "public.xml"); string(got) != input {
		t.Fatalf("XML=%q, want %q", got, input)
	}
	if got := representation(t, run.board, typePlain); string(got) != "snowman" {
		t.Fatalf("fallback=%q", got)
	}
}

func TestDetectsHTMLDocumentsBeforeXML(t *testing.T) {
	for _, input := range []string{
		"<!DOCTYPE html><html><body><br>Hello</body></html>",
		"<HTML lang=\"en\"><body>Hello</body></HTML>",
		"<?xml version=\"1.0\"?><html xmlns=\"http://www.w3.org/1999/xhtml\"><body>Hello</body></html>",
	} {
		t.Run(input, func(t *testing.T) {
			run := newInvocation(t, bytes.NewBufferString(input))
			if code := run.command.Run(nil); code != 0 {
				t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
			}
			if got := representation(t, run.board, typeHTML); string(got) != input {
				t.Fatalf("HTML=%q, want %q", got, input)
			}
		})
	}
}

func TestAmbiguousStructuredInputRemainsText(t *testing.T) {
	for _, input := range []string{
		"true", "42", `"hello"`, `{"broken":}`, "{} trailing", "[1,]",
		"<b>hello</b>", "<root/>", "<htmlish>hello</htmlish>", "<html>unfinished",
		"<?xml broken?><root/>", "<?xml version=\"1.0\"?><a/><b/>",
		"<?xml version=\"1.0\"?><a></b>", "<?xml version=\"1.0\"?>outside<a/>",
		"<?xml version=\"1.0\"?><root/><!garbage>",
		"<?xml version=\"1.0\"?><root><!garbage></root>",
		"<?xml version=\"1.0\"?><root><?xml version=\"1.0\"?></root>",
		"<?xml version=\"1.0\"?><root a=\"1\" a=\"2\"/>",
	} {
		t.Run(input, func(t *testing.T) {
			run := newInvocation(t, bytes.NewBufferString(input))
			if code := run.command.Run(nil); code != 0 {
				t.Fatalf("exit=%d stderr=%s", code, &run.stderr)
			}
			if got := representation(t, run.board, typePlain); string(got) != input {
				t.Fatalf("text=%q, want %q", got, input)
			}
		})
	}
}
