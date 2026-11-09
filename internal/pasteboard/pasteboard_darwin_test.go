//go:build darwin && cgo

package pasteboard_test

import (
	"bytes"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

func TestPublishRepresentations(t *testing.T) {
	board, err := pasteboard.Unique()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := board.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := board.Write([]pasteboard.Representation{
		{Type: "public.html", Data: []byte("<b>hello</b>")},
		{Type: "public.utf8-plain-text", Data: []byte("hello")},
	}); err != nil {
		t.Fatal(err)
	}
	items, err := board.Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d pasteboard items, want one", len(items))
	}
	want := map[string][]byte{
		"public.html":            []byte("<b>hello</b>"),
		"public.utf8-plain-text": []byte("hello"),
	}
	for _, rep := range items[0].Representations {
		if data, ok := want[rep.Type]; ok {
			if !bytes.Equal(rep.Data, data) {
				t.Errorf("%s: got %q, want %q", rep.Type, rep.Data, data)
			}
			delete(want, rep.Type)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing representations: %v", want)
	}
}
