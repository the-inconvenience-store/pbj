package command

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/the-inconvenience-store/pbj/internal/pasteboard"
)

const fileURLType = "public.file-url"

func extractFilePaths(snapshot *pasteboard.Snapshot, first, last int, separator byte) ([]byte, error) {
	if first == last {
		return nil, errors.New("clipboard has no file references")
	}
	var output []byte
	for i := first; i < last; i++ {
		data, err := snapshot.Data(i, fileURLType)
		if err != nil {
			return nil, err
		}
		path, err := localFilePath(data)
		if err != nil {
			return nil, fmt.Errorf("file reference in item %d: %w", i, err)
		}
		if separator == '\n' && strings.ContainsRune(path, '\n') {
			return nil, errors.New("file path contains a newline; use --files --null")
		}
		output = append(output, path...)
		output = append(output, separator)
	}
	return output, nil
}

func localFilePath(data []byte) (string, error) {
	location, err := url.Parse(string(data))
	if err != nil {
		return "", fmt.Errorf("invalid file URL: %w", err)
	}
	if location.Scheme != "file" || location.User != nil || (location.Host != "" && !strings.EqualFold(location.Host, "localhost")) {
		return "", errors.New("reference must be a local file URL")
	}
	if location.RawQuery != "" || location.ForceQuery || location.Fragment != "" {
		return "", errors.New("file URL must not contain a query or fragment")
	}
	path := location.Path
	if !filepath.IsAbs(path) || !utf8.ValidString(path) || strings.ContainsRune(path, 0) {
		return "", errors.New("file URL must contain an absolute UTF-8 path without NUL")
	}
	return path, nil
}
