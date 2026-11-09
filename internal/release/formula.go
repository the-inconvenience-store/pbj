// Package release prepares source-built Homebrew updates from tagged release archives.
package release

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const repositoryURL = "https://github.com/the-inconvenience-store/pbj"

// Release is a validated stable version and the checksum of its actual source archive.
// Its fields are immutable after Prepare; no source bytes are retained.
type Release struct {
	tag      string
	version  string
	archive  string
	checksum [sha256.Size]byte
}

// ValidateTag accepts only canonical stable vMAJOR.MINOR.PATCH tags.
func ValidateTag(tag string) error {
	const versionComponents = 3
	if !strings.HasPrefix(tag, "v") {
		return invalidTag(tag)
	}
	parts := strings.Split(tag[1:], ".")
	if len(parts) != versionComponents {
		return invalidTag(tag)
	}
	for _, part := range parts {
		if !numericIdentifier(part) {
			return invalidTag(tag)
		}
	}
	return nil
}

func numericIdentifier(part string) bool {
	if part == "" || (len(part) > 1 && part[0] == '0') {
		return false
	}
	for _, digit := range part {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func invalidTag(tag string) error {
	return fmt.Errorf("invalid stable release tag %q; use vMAJOR.MINOR.PATCH without leading zeroes", tag)
}

// Prepare verifies archive/version agreement and streams the archive checksum.
func Prepare(tag, source string) (*Release, error) {
	if err := ValidateTag(tag); err != nil {
		return nil, err
	}
	version := tag[1:]
	expected := "pbj_" + version + "_source.tar.gz"
	if filepath.Base(source) != expected {
		return nil, fmt.Errorf("source archive must be named %s for tag %s", expected, tag)
	}
	// #nosec G304 -- The operator selects this local archive; its basename is validated against the release tag above.
	archive, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("open source archive: %w", err)
	}
	hash := sha256.New()
	_, readErr := io.Copy(hash, archive)
	if err := errors.Join(readErr, archive.Close()); err != nil {
		return nil, fmt.Errorf("hash source archive: %w", err)
	}
	result := &Release{tag: tag, version: version, archive: expected}
	hash.Sum(result.checksum[:0])
	return result, nil
}

// Formula renders a source-only Homebrew formula, including both versioned commands.
func (r *Release) Formula() []byte {
	return fmt.Appendf(nil, formulaTemplate, repositoryURL+"/releases/download/"+r.tag+"/"+r.archive, r.version, r.checksum[:])
}

//nolint:dupword // Repeated Ruby block endings are syntax, not duplicate prose.
const formulaTemplate = `class Pbj < Formula
  desc "Enriched macOS clipboard copy and paste commands"
  homepage "https://github.com/the-inconvenience-store/pbj"
  url "%s"
  version "%s"
  sha256 "%x"

  depends_on "go" => :build
  depends_on macos: :ventura

  def install
    ENV["CGO_ENABLED"] = "1"
    ENV["GOTOOLCHAIN"] = "local"
    ENV["GOFLAGS"] = "-mod=readonly"
    ENV["MACOSX_DEPLOYMENT_TARGET"] = "13.0"
    ENV.append "CGO_CFLAGS", "-mmacosx-version-min=13.0"

    %%w[pbjcopy pbjpaste].each do |command|
      system "go", "build", *std_go_args(output: bin/command, ldflags: "-X main.version=#{version}"), "./cmd/#{command}"
    end
  end

  test do
    %%w[pbjcopy pbjpaste].each do |command|
      assert_equal "#{command} #{version}\n", shell_output("#{bin}/#{command} --version")
      assert_empty shell_output("#{bin}/#{command} --invalid-release-test-option", 2)
    end
  end
end
`
