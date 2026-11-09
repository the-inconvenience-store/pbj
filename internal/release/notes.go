package release

import (
	"bytes"
	"fmt"
)

// Notes selects Cog's committed section, not a regeneration that includes the
// later version-bump commit. The result shares the changelog's backing storage.
func Notes(tag string, changelog []byte) ([]byte, error) {
	if err := ValidateTag(tag); err != nil {
		return nil, err
	}
	start, offset := -1, 0
	for line := range bytes.SplitSeq(changelog, []byte("\n")) {
		heading := bytes.HasPrefix(line, []byte("## "))
		if start >= 0 && (heading || bytes.Equal(line, []byte("- - -"))) {
			return bytes.TrimSpace(changelog[start:offset]), nil
		}
		if heading {
			version, _, found := bytes.Cut(line[len("## "):], []byte(" - "))
			if found && string(version) == tag {
				start = offset
			}
		}
		offset += len(line) + 1
	}
	if start >= 0 {
		return bytes.TrimSpace(changelog[start:]), nil
	}
	return nil, fmt.Errorf("committed changelog has no notes for %s", tag)
}
