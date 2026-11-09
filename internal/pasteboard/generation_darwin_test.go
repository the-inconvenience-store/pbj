//go:build darwin && cgo

package pasteboard_test

import (
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

func TestCaptureRejectsReplacementInsteadOfReadingNewGeneration(t *testing.T) {
	board, err := pasteboard.Unique()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := board.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := board.Write([]pasteboard.Representation{{Type: nativePlainType, Data: []byte("before")}}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := board.Capture()
	if err != nil {
		t.Fatal(err)
	}
	defer snapshot.Close()
	if err := board.Write([]pasteboard.Representation{{Type: nativePlainType, Data: []byte("after")}}); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshot.Data(0, nativePlainType); err == nil {
		t.Fatal("replacement was not detected")
	}
	snapshot.Close()
	if _, err := snapshot.Data(0, nativePlainType); err == nil {
		t.Fatal("closed snapshot remained readable")
	}
}
