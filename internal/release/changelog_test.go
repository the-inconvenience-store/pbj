package release_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCogBumpDoesNotPublishAuthorIdentity(t *testing.T) {
	fixture := cogFixture(t)
	runFixture(t, fixture, "git", "init", "--initial-branch=main")
	runFixture(t, fixture, "git", "config", "user.name", "Private Release Author")
	runFixture(t, fixture, "git", "config", "user.email", "private-author@example.invalid")
	runFixture(t, fixture, "git", "config", "commit.gpgsign", "false")
	runFixture(t, fixture, "git", "add", ".")
	runFixture(t, fixture, "git", "commit", "-m", "feat(clipboard)!: preserve archive hashes", "-m", "BREAKING CHANGE: use explicit formats\n\nCo-authored-by: Private Coauthor <private-coauthor@example.invalid>")
	runFixture(t, fixture, "cog", "bump", "--version", "0.1.0")
	notes := runFixture(t, fixture, "cog", "changelog")
	for _, identity := range []string{"Private Release Author", "private-author@example.invalid", "Private Coauthor", "private-coauthor@example.invalid"} {
		if strings.Contains(notes, identity) {
			t.Fatalf("changelog exposed a fixture author identity")
		}
	}
	if !strings.Contains(notes, "preserve archive hashes") {
		t.Fatalf("release notes lost the user-facing change: %s", notes)
	}
}

func cogFixture(t *testing.T) string {
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
	return fixture
}

func runFixture(t *testing.T, directory, program string, arguments ...string) string {
	t.Helper()
	// #nosec G204 -- Programs and arguments are fixed test-owned Cog/Git fixture commands.
	command := exec.Command(program, arguments...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", program, err, output)
	}
	return string(output)
}
