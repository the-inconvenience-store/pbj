//go:build darwin && cgo

package pasteboard_test

import (
	"bytes"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

func TestLocaleTextDecoding(t *testing.T) {
	cases := []struct {
		name      string
		locale    string
		input     []byte
		want      []byte
		wantError bool
	}{
		{name: "UTF-8", locale: "en_US.UTF-8", input: []byte("snowman ☃"), want: []byte("snowman ☃")},
		{name: "Latin-1 conversion", locale: "fr_FR.ISO8859-1", input: []byte{'c', 'a', 'f', 0xe9}, want: []byte("café")},
		{name: "C rejects non-ASCII", locale: "C", input: []byte{0xff}, wantError: true},
		{name: "malformed UTF-8", locale: "en_US.UTF-8", input: []byte{0xc3}, wantError: true},
		{name: "unknown locale falls back to C", locale: "pbj_nonexistent_locale", input: []byte{0xff}, wantError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := pasteboard.DecodeText(tc.input, tc.locale)
			if (err != nil) != tc.wantError {
				t.Fatalf("error=%v; wantError=%v", err, tc.wantError)
			}
			if !tc.wantError && !bytes.Equal(got, tc.want) {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
