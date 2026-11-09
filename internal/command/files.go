package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func normalizeFiles(paths []string) ([]string, error) {
	normalized := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if path == "" {
			return nil, errors.New("file-reference path must not be empty")
		}
		if strings.HasPrefix(path, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("expand home directory: %w", err)
			}
			path = filepath.Join(home, path[2:])
		}
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve file reference %q: %w", path, err)
		}
		if _, exists := seen[absolute]; !exists {
			normalized = append(normalized, absolute)
			seen[absolute] = struct{}{}
		}
	}
	return normalized, nil
}
