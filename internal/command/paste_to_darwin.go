//go:build darwin && cgo

package command

/*
#include <copyfile.h>
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

func pasteFilesTo(snapshot *pasteboard.Snapshot, first, last int, destination string) error {
	if first == last {
		return errors.New("clipboard has no file references")
	}
	directory, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("resolve destination: %w", err)
	}
	info, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf("destination: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("destination %q is not a directory", directory)
	}
	directory, err = filepath.EvalSymlinks(directory)
	if err != nil {
		return fmt.Errorf("resolve destination links: %w", err)
	}
	for i := first; i < last; i++ {
		data, err := snapshot.Data(i, fileURLType)
		if err != nil {
			return err
		}
		source, err := localFilePath(data)
		if err != nil {
			return fmt.Errorf("file reference in item %d: %w", i, err)
		}
		if err := copyReferenceTo(filepath.Clean(source), directory); err != nil {
			return fmt.Errorf("copy item %d (%q): %w", i, source, err)
		}
	}
	return nil
}

func copyReferenceTo(source, directory string) (err error) {
	if err := validateCopySource(source, directory); err != nil {
		return err
	}
	target := filepath.Join(directory, filepath.Base(source))
	if _, err := os.Lstat(target); err == nil {
		return fmt.Errorf("destination %q: %w", target, os.ErrExist)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	stage, err := os.MkdirTemp(directory, ".pbjpaste-")
	if err != nil {
		return fmt.Errorf("prepare destination: %w", err)
	}
	defer func() { err = errors.Join(err, os.RemoveAll(stage)) }()
	return copyAndPublish(source, filepath.Join(stage, "item"), target)
}

func copyAndPublish(source, staging, target string) error {
	nativeSource := C.CString(source)
	defer C.free(unsafe.Pointer(nativeSource))
	nativeStaging := C.CString(staging)
	defer C.free(unsafe.Pointer(nativeStaging))
	if result, err := C.copyfile(nativeSource, nativeStaging, nil, C.COPYFILE_ALL|C.COPYFILE_EXCL|C.COPYFILE_NOFOLLOW|C.COPYFILE_RECURSIVE); result != 0 {
		return fmt.Errorf("prepare copy: %w", err)
	}
	nativeTarget := C.CString(target)
	defer C.free(unsafe.Pointer(nativeTarget))
	if result, err := C.renamex_np(nativeStaging, nativeTarget, C.RENAME_EXCL); result != 0 {
		return fmt.Errorf("publish %q without overwriting: %w", target, err)
	}
	return nil
}

func validateCopySource(source, directory string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return nil
	}
	realSource, err := filepath.EvalSymlinks(source)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(realSource, directory)
	if err != nil {
		return err
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return errors.New("destination must not be inside the source directory")
	}
	return nil
}
