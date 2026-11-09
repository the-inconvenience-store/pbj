package release_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/release"
)

func TestValidateTagRejectsAmbiguousAndUnsafeVersions(t *testing.T) {
	for _, tag := range []string{"1.2.3", "v1.2", "v01.2.3", "v1.02.3", "v1.2.03", "v1.2.3-rc1", "v1.2.3/extra", "v1.2.3\n", "v1.2.3+build", "v"} {
		if err := release.ValidateTag(tag); err == nil {
			t.Errorf("invalid stable release tag accepted: %q", tag)
		}
	}
	if err := release.ValidateTag("v1.2.3"); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareRejectsSourceFromDifferentVersion(t *testing.T) {
	source := filepath.Join(t.TempDir(), "pbj_1.2.4_source.tar.gz")
	if err := os.WriteFile(source, []byte("archive fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := release.Prepare("v1.2.3", source); err == nil {
		t.Fatal("a source archive for another version was accepted")
	}
}
