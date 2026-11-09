//go:build darwin && cgo

package pasteboard

/*
#cgo LDFLAGS: -liconv
#include <stdlib.h>
#include "pasteboard.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"unicode/utf8"
	"unsafe"
)

// DecodeText decodes bytes in the supplied locale to UTF-8 without replacing
// malformed input. Unknown or empty locales use C (ASCII). It never changes the
// process locale. Valid UTF-8 and ASCII results may alias data.
func DecodeText(data []byte, locale string) ([]byte, error) {
	if locale == "" {
		locale = "C"
	}
	if strings.IndexByte(locale, 0) >= 0 {
		return nil, errors.New("text locale must not contain NUL")
	}
	nativeLocale := C.CString(locale)
	defer C.free(unsafe.Pointer(nativeLocale))
	var nativeError *C.char
	encoding := C.pbj_text_encoding(nativeLocale, &nativeError)
	if encoding == nil {
		return nil, nativeFailure(nativeError, "resolve text locale")
	}
	defer C.free(unsafe.Pointer(encoding))
	codeset := C.GoString(encoding)
	switch strings.ToUpper(codeset) {
	case "UTF-8", "UTF8":
		if !utf8.Valid(data) {
			return nil, fmt.Errorf("decode text in locale %q: malformed UTF-8", locale)
		}
		return data, nil
	case "US-ASCII", "ASCII", "ANSI_X3.4-1968", "ANSI_X3.4-1986":
		for _, value := range data {
			if value >= utf8.RuneSelf {
				return nil, fmt.Errorf("decode text in locale %q: byte 0x%02x is not ASCII", locale, value)
			}
		}
		return data, nil
	}
	if len(data) == 0 {
		return data, nil
	}
	var output unsafe.Pointer
	var outputLength C.size_t
	ok := C.pbj_text_convert(unsafe.Pointer(unsafe.SliceData(data)), C.size_t(len(data)), encoding,
		&output, &outputLength, &nativeError)
	runtime.KeepAlive(data)
	if ok == 0 {
		return nil, nativeFailure(nativeError, "decode text")
	}
	defer C.free(output)
	return append([]byte{}, unsafe.Slice((*byte)(output), int(outputLength))...), nil
}
