package release

import (
	"errors"
	"fmt"
	"strings"
)

func (r *Release) rejectDowngrade(formula []byte) error {
	current := ""
	for line := range strings.SplitSeq(string(formula), "\n") {
		value, matched := strings.CutPrefix(strings.TrimSpace(line), "version \"")
		if !matched {
			continue
		}
		value, quoted := strings.CutSuffix(value, "\"")
		if !quoted || current != "" {
			return errors.New("tap formula has ambiguous version metadata")
		}
		current = value
	}
	if err := ValidateTag("v" + current); err != nil {
		return fmt.Errorf("cannot establish current tap version: %w", err)
	}
	if versionLess(r.version, current) {
		return fmt.Errorf("refusing to replace newer tap version %s with %s", current, r.version)
	}
	return nil
}

// Both versions have already passed the canonical stable-version boundary.
// Compare digit lengths before lexical values so components cannot overflow.
func versionLess(left, right string) bool {
	leftParts, rightParts := strings.Split(left, "."), strings.Split(right, ".")
	for i, a := range leftParts {
		b := rightParts[i]
		if len(a) != len(b) {
			return len(a) < len(b)
		}
		if a != b {
			return a < b
		}
	}
	return false
}
