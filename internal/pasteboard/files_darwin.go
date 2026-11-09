//go:build darwin && cgo

package pasteboard

/*
#include "pasteboard.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"
)

// WriteFiles publishes one eager file-URL item per absolute path, preserving
// order and symlink paths. Every target must exist; directories are accepted.
// Only filesystem metadata is inspected, never file contents. All preparation
// finishes before publication, with the same publication guarantees as Write.
func (b *Board) WriteFiles(paths []string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.handle == nil {
		return errors.New("pasteboard is closed")
	}
	if len(paths) == 0 {
		return errors.New("file references must not be empty")
	}
	var nativeError *C.char
	files := C.pbj_files_create(&nativeError)
	if files == nil {
		return nativeFailure(nativeError, "prepare file references")
	}
	defer C.pbj_release(files)
	for _, path := range paths {
		if err := validIdentifier(path); err != nil {
			return fmt.Errorf("file reference path %q: %w", path, err)
		}
		if !filepath.IsAbs(path) {
			return fmt.Errorf("file reference path %q must be absolute", path)
		}
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("prepare file reference: %w", err)
		}
		var directory C.int
		if info.IsDir() {
			directory = 1
		}
		ok := C.pbj_files_add(files, b.handle, unsafe.Pointer(unsafe.StringData(path)), C.size_t(len(path)),
			directory, &nativeError)
		runtime.KeepAlive(path)
		if ok == 0 {
			return nativeFailure(nativeError, "prepare file reference")
		}
	}
	if C.pbj_files_write(b.handle, files, &nativeError) == 0 {
		return nativeFailure(nativeError, "publish file references")
	}
	return nil
}
