//go:build darwin && cgo

// Package pasteboard owns synchronous, eager access to macOS pasteboards.
// A Board serializes its operations; separate handles and other processes may
// still replace its contents. Close is required, especially for unique boards.
package pasteboard

/*
#cgo CFLAGS: -x objective-c -fobjc-arc -mmacosx-version-min=13.0
#cgo LDFLAGS: -framework Foundation -framework AppKit -framework ApplicationServices -mmacosx-version-min=13.0
#include <stdlib.h>
#include "pasteboard.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unicode/utf8"
	"unsafe"
)

// Representation is one eager byte representation of a pasteboard item.
// Type is a nonempty UTF-8 pasteboard identifier without NUL bytes.
type Representation struct {
	Type string
	Data []byte
}

// Item groups alternative representations of the same pasteboard item.
type Item struct {
	Representations []Representation
}

// Board owns a retained native pasteboard handle. Do not copy a Board.
// Data passed to Write is borrowed synchronously and copied by AppKit.
type Board struct {
	mu     sync.Mutex
	handle unsafe.Pointer
	name   string
	mode   boardMode
}

type boardMode uint8

const (
	namedBoard boardMode = iota
	uniqueBoard
)

// Open opens a standard board (general, find, ruler, font) or an actual named
// board. Closing this handle never releases the named board globally.
func Open(name string) (*Board, error) {
	if err := validIdentifier(name); err != nil {
		return nil, fmt.Errorf("pasteboard name: %w", err)
	}
	return acquireBoard(name, namedBoard)
}

// Unique creates a uniquely named board. Close releases it globally.
func Unique() (*Board, error) {
	return acquireBoard("", uniqueBoard)
}

func acquireBoard(name string, mode boardMode) (*Board, error) {
	var actualName, nativeError *C.char
	var flag C.int
	if mode == uniqueBoard {
		flag = 1
	}
	handle := C.pbj_board_open(unsafe.Pointer(unsafe.StringData(name)), C.size_t(len(name)), flag, &actualName, &nativeError)
	runtime.KeepAlive(name)
	if handle == nil {
		return nil, nativeFailure(nativeError, "open pasteboard")
	}
	defer C.free(unsafe.Pointer(actualName))
	return &Board{handle: handle, name: C.GoString(actualName), mode: mode}, nil
}

// Name returns the actual native name, including after Close.
func (b *Board) Name() string {
	return b.name
}

// Write prepares one complete item before clearing and publishing to the board.
// Preparation errors leave existing contents untouched. Once publication starts,
// AppKit provides no transaction or rollback guarantee.
func (b *Board) Write(representations []Representation) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.handle == nil {
		return errors.New("pasteboard is closed")
	}
	if len(representations) == 0 {
		return errors.New("pasteboard item has no representations")
	}
	for _, representation := range representations {
		if err := validIdentifier(representation.Type); err != nil {
			return fmt.Errorf("pasteboard representation type: %w", err)
		}
	}
	var nativeError *C.char
	item := C.pbj_item_create(&nativeError)
	if item == nil {
		return nativeFailure(nativeError, "prepare pasteboard item")
	}
	defer C.pbj_release(item)
	for _, representation := range representations {
		ok := C.pbj_item_set(item,
			unsafe.Pointer(unsafe.StringData(representation.Type)), C.size_t(len(representation.Type)),
			unsafe.Pointer(unsafe.SliceData(representation.Data)), C.size_t(len(representation.Data)), &nativeError)
		runtime.KeepAlive(representation)
		if ok == 0 {
			return nativeFailure(nativeError, "prepare pasteboard representation")
		}
	}
	if C.pbj_board_write(b.handle, item, &nativeError) == 0 {
		return nativeFailure(nativeError, "publish pasteboard item")
	}
	return nil
}

// Read eagerly copies every available item and representation into Go ownership.
func (b *Board) Read() ([]Item, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.handle == nil {
		return nil, errors.New("pasteboard is closed")
	}
	var nativeError *C.char
	snapshot := C.pbj_board_read(b.handle, &nativeError)
	if snapshot == nil {
		return nil, nativeFailure(nativeError, "read pasteboard")
	}
	defer C.pbj_snapshot_free(snapshot)
	nativeItems := unsafe.Slice(snapshot.items, int(snapshot.count))
	items := make([]Item, len(nativeItems))
	for i, item := range nativeItems {
		representations := unsafe.Slice(item.representations, int(item.count))
		items[i].Representations = make([]Representation, len(representations))
		for j, representation := range representations {
			items[i].Representations[j] = Representation{
				Type: string(unsafe.Slice((*byte)(representation._type), int(representation.type_length))),
				Data: append([]byte{}, unsafe.Slice((*byte)(representation.data), int(representation.data_length))...),
			}
		}
	}
	return items, nil
}

// Close releases the native handle exactly once. Only boards created by Unique
// are released globally; ordinary named and standard boards retain their data.
func (b *Board) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.handle == nil {
		return nil
	}
	handle := b.handle
	b.handle = nil
	var flag C.int
	if b.mode == uniqueBoard {
		flag = 1
	}
	var nativeError *C.char
	if C.pbj_board_close(handle, flag, &nativeError) == 0 {
		return nativeFailure(nativeError, "close pasteboard")
	}
	return nil
}

func validIdentifier(identifier string) error {
	if identifier == "" {
		return errors.New("must not be empty")
	}
	if !utf8.ValidString(identifier) {
		return errors.New("must be valid UTF-8")
	}
	if strings.IndexByte(identifier, 0) >= 0 {
		return errors.New("must not contain NUL")
	}
	return nil
}

func nativeFailure(message *C.char, operation string) error {
	if message == nil {
		return fmt.Errorf("%s failed", operation)
	}
	defer C.free(unsafe.Pointer(message))
	return fmt.Errorf("%s: %s", operation, C.GoString(message))
}
