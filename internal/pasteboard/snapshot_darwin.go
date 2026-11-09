//go:build darwin && cgo

package pasteboard

/*
#include "pasteboard.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"unsafe"
)

// Snapshot retains item metadata without fetching representation payloads.
// Data fails if another writer changes the board. Close releases the capture,
// never the clipboard contents. Do not copy a Snapshot.
type Snapshot struct {
	mu     sync.Mutex
	handle *C.pbj_snapshot
	types  [][]string
}

// Capture inspects the current items and types without reading their payloads.
// The caller must Close the returned snapshot, even after closing the Board.
func (b *Board) Capture() (*Snapshot, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.handle == nil {
		return nil, errors.New("pasteboard is closed")
	}
	var nativeError *C.char
	handle := C.pbj_board_capture(b.handle, &nativeError)
	if handle == nil {
		return nil, nativeFailure(nativeError, "inspect pasteboard")
	}
	snapshot := &Snapshot{handle: handle}
	items := unsafe.Slice(handle.items, int(handle.count))
	snapshot.types = make([][]string, len(items))
	for i, item := range items {
		representations := unsafe.Slice(item.representations, int(item.count))
		snapshot.types[i] = make([]string, len(representations))
		for j, representation := range representations {
			snapshot.types[i][j] = string(unsafe.Slice((*byte)(representation._type), int(representation.type_length)))
		}
	}
	return snapshot, nil
}

// Types returns the captured types per item in native order. Treat the returned
// metadata as read-only. It remains available after Close.
func (s *Snapshot) Types() [][]string { return s.types }

// Data copies only the exact advertised representation into Go ownership.
// Item indexes are zero-based. No fallback or format conversion is performed.
func (s *Snapshot) Data(item int, typeName string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.handle == nil {
		return nil, errors.New("pasteboard snapshot is closed")
	}
	if item < 0 || item >= len(s.types) {
		return nil, fmt.Errorf("pasteboard item %d does not exist", item)
	}
	index := slices.Index(s.types[item], typeName)
	if index < 0 {
		return nil, fmt.Errorf("pasteboard item %d has no representation %q", item, typeName)
	}
	var data unsafe.Pointer
	var length C.size_t
	var nativeError *C.char
	owner := C.pbj_snapshot_data(s.handle, C.size_t(item), C.size_t(index), &data, &length, &nativeError)
	if owner == nil {
		return nil, nativeFailure(nativeError, "extract pasteboard representation")
	}
	defer C.pbj_release(owner)
	return append([]byte{}, unsafe.Slice((*byte)(data), int(length))...), nil
}

// Close releases captured native references exactly once; it never clears data.
func (s *Snapshot) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.handle != nil {
		C.pbj_snapshot_free(s.handle)
		s.handle = nil
	}
}
