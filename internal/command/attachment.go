package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	privateDirectoryMode = 0o700
	privateFileMode      = 0o600
)

func (cmd Command) prepareAttachment(name string) (path string, err error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("locate attachment cache: %w", err)
	}
	cache, err = filepath.Abs(cache)
	if err != nil {
		return "", fmt.Errorf("resolve attachment cache: %w", err)
	}
	if err := os.MkdirAll(cache, privateDirectoryMode); err != nil {
		return "", fmt.Errorf("prepare attachment cache: %w", err)
	}
	directory, err := os.MkdirTemp(cache, "pbj-attachment-")
	if err != nil {
		return "", fmt.Errorf("create attachment directory: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.RemoveAll(directory))
		}
	}()
	path = filepath.Join(directory, name)
	root, err := os.OpenRoot(directory)
	if err != nil {
		return "", fmt.Errorf("open attachment directory: %w", err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, privateFileMode)
	if err != nil {
		return "", fmt.Errorf("create attachment: %w", err)
	}
	_, copyErr := io.Copy(file, cmd.Stdin)
	if err := errors.Join(copyErr, file.Close()); err != nil {
		return "", fmt.Errorf("prepare attachment bytes: %w", err)
	}
	return path, nil
}
