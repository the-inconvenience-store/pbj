//go:build darwin && cgo

package pasteboard_test

import (
	"bytes"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

const nativeUTF8Locale = "en_US.UTF-8"

func TestEncodeTextPreservesLocaleAndRejectsLoss(t *testing.T) {
	got, err := pasteboard.EncodeText([]byte("café"), "fr_FR.ISO8859-1")
	if err != nil || !bytes.Equal(got, []byte{'c', 'a', 'f', 0xe9}) {
		t.Fatalf("encoded=%v err=%v", got, err)
	}
	for _, test := range []struct {
		data   []byte
		locale string
	}{
		{[]byte("café"), "C"},
		{[]byte("☃"), "fr_FR.ISO8859-1"},
		{[]byte{255}, nativeUTF8Locale},
	} {
		if _, err := pasteboard.EncodeText(test.data, test.locale); err == nil {
			t.Fatalf("lossy/invalid conversion accepted: %v in %s", test.data, test.locale)
		}
	}
}
