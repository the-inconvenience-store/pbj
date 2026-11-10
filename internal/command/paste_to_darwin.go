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
	directory, err := resolvePhysicalPath(destination)
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
	for i := first; i < last; i++ {
		data, err := snapshot.Data(i, fileURLType)
		if err != nil {
			return err
		}
		source, err := localFilePath(data)
		if err != nil {
			return fmt.Errorf("file reference in item %d: %w", i, err)
		}
		if err := copyReferenceTo(source, directory); err != nil {
			return fmt.Errorf("copy item %d (%q): %w", i, source, err)
		}
	}
	return nil
}

func copyReferenceTo(source, directory string) (err error) {
	source, err = resolveCopySource(source)
	if err != nil {
		return err
	}
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
	relative, err := filepath.Rel(source, directory)
	if err != nil {
		return err
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return errors.New("destination must not be inside the source directory")
	}
	return nil
}

// Resolve traversal before cleaning: link/.. refers to the symlink target's parent.
func resolvePhysicalPath(path string) (string, error) {
	if strings.ContainsRune(path, 0) {
		return "", fmt.Errorf("path %q: %w", path, os.ErrInvalid)
	}
	nativePath := C.CString(path)
	defer C.free(unsafe.Pointer(nativePath))
	resolved, err := C.realpath(nativePath, nil)
	if resolved == nil {
		return "", fmt.Errorf("resolve %q: %w", path, err)
	}
	defer C.free(unsafe.Pointer(resolved))
	return C.GoString(resolved), nil
}

// Resolve only source parents so the final symlink is copied, not dereferenced.
func resolveCopySource(source string) (string, error) {
	source = strings.TrimRight(source, string(os.PathSeparator))
	if source == "" {
		return string(os.PathSeparator), nil
	}
	split := strings.LastIndexByte(source, os.PathSeparator)
	name := source[split+1:]
	if name == "." || name == ".." {
		return resolvePhysicalPath(source)
	}
	parent, err := resolvePhysicalPath(source[:split+1])
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, name), nil
}
