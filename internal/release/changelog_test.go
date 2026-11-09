package release_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/the-inconvenience-store/pbj/internal/release"
)

func TestCogBumpDoesNotPublishAuthorIdentity(t *testing.T) {
	fixture, root := cogFixture(t)
	runFixture(t, fixture, "git", "init", "--initial-branch=main")
	runFixture(t, fixture, "git", "config", "user.name", "Private Release Author")
	runFixture(t, fixture, "git", "config", "user.email", "private-author@example.invalid")
	runFixture(t, fixture, "git", "config", "commit.gpgsign", "false")
	runFixture(t, fixture, "git", "add", ".")
	runFixture(t, fixture, "git", "commit", "-m", "feat(clipboard)!: preserve archive hashes", "-m", "BREAKING CHANGE: use explicit formats\n\nCo-authored-by: Private Coauthor <private-coauthor@example.invalid>")
	runFixture(t, fixture, "cog", "bump", "--version", "0.1.0")
	runFixture(t, fixture, "git", "commit", "--allow-empty", "-m", "fix: keep next release separate")
	runFixture(t, fixture, "cog", "bump", "--version", "0.1.1")
	changelog, err := root.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatal(err)
	}
	selected, err := release.Notes("v0.1.0", changelog)
	if err != nil {
		t.Fatal(err)
	}
	notes := string(selected)
	if !strings.HasPrefix(notes, "## v0.1.0 - ") || strings.Contains(notes, "keep next release separate") || strings.Contains(notes, "**version:**") {
		t.Fatalf("notes did not select the committed exact-tag section: %s", notes)
	}
	if _, err := release.Notes("v9.9.9", changelog); err == nil {
		t.Fatal("missing release notes were accepted")
	}
	for _, identity := range []string{"Private Release Author", "private-author@example.invalid", "Private Coauthor", "private-coauthor@example.invalid"} {
		if strings.Contains(notes, identity) {
			t.Fatalf("changelog exposed a fixture author identity")
		}
	}
	if !strings.Contains(notes, "preserve archive hashes") {
		t.Fatalf("release notes lost the user-facing change: %s", notes)
	}
}

func cogFixture(t *testing.T) (string, *os.Root) {
	t.Helper()
	t.Chdir("../..")
	fixture := t.TempDir()
	root, err := os.OpenRoot(fixture)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	})
	config, err := os.ReadFile("cog.toml")
	if err != nil {
		t.Fatal(err)
	}
	if err := root.WriteFile("cog.toml", config, 0o600); err != nil {
		t.Fatal(err)
	}
	template, err := os.ReadFile("scripts/changelog.tera")
	if err != nil {
		t.Fatal(err)
	}
	if err := root.Mkdir("scripts", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := root.WriteFile("scripts/changelog.tera", template, 0o600); err != nil {
		t.Fatal(err)
	}
	return fixture, root
}

func runFixture(t *testing.T, directory, program string, arguments ...string) {
	t.Helper()
	// #nosec G204 -- Programs and arguments are fixed test-owned Cog/Git fixture commands.
	command := exec.Command(program, arguments...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", program, err, output)
	}
}
