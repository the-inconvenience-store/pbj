//go:build darwin && cgo

package pasteboard_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

const nativePlainType = "public.utf8-plain-text"

func TestCaptureSelectsExactRepresentation(t *testing.T) {
	board, err := pasteboard.Unique()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := board.Close(); err != nil {
			t.Error(err)
		}
	})
	payload := []byte{0, 255, 1, 128}
	if err := board.Write([]pasteboard.Representation{{Type: "com.example.raw", Data: payload}, {Type: nativePlainType, Data: []byte("fallback")}}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := board.Capture()
	if err != nil {
		t.Fatal(err)
	}
	defer snapshot.Close()
	types := snapshot.Types()
	if len(types) != 1 || !slices.Contains(types[0], "com.example.raw") || !slices.Contains(types[0], nativePlainType) {
		t.Fatalf("types=%v", types)
	}
	got, err := snapshot.Data(0, "com.example.raw")
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("data=%v err=%v", got, err)
	}
	if _, err := snapshot.Data(0, "public.png"); err == nil {
		t.Fatal("missing type succeeded")
	}
	if _, err := snapshot.Data(1, "com.example.raw"); err == nil {
		t.Fatal("missing item succeeded")
	}
}
