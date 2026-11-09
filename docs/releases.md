# Releases

pbj releases support only stable tags of the form `vMAJOR.MINOR.PATCH`.
Prereleases and build-metadata tags are rejected. Cog owns version selection,
`CHANGELOG.md`, version commits, and tags. Its repository template omits author
names, signatures, email addresses, and usernames from changelog entries.
GoReleaser publishes GitHub releases and archives; it does not select versions
or generate a second set of published release notes.

## Prepare publication

Before the first release:

1. Ensure GitHub Actions is enabled for
   `the-inconvenience-store/pbj`, and the release workflow is on the default
   branch. No pull-request or branch-push CI is configured.
2. Ensure the public
   [`the-inconvenience-store/homebrew-tap`](https://github.com/the-inconvenience-store/homebrew-tap)
   repository is available. Its default branch must permit the publication
   credential to create and update `Formula/pbj.rb` through the GitHub Contents
   API. Repository provisioning is separate from the release workflow.
3. Create a fine-grained GitHub token restricted to **only** that tap repository,
   with **Contents: read and write** (and the automatically required Metadata
   access). Complete organization approval or SSO authorization if required.
   Store it as the **`HOMEBREW_TAP_TOKEN`** Actions secret in the **pbj** repository.
   A repository-limited GitHub App installation token with the same Contents
   permission is also suitable, but must still be valid when the job runs.
   Do not grant workflow, administration, or access to unrelated repositories.
4. Use a repository-approved non-personal Git author and committer identity for
   version commits. Review commit subjects for personal information too: the
   template omits author metadata, not text deliberately placed in subjects.
5. On macOS with Apple Command Line Tools and Python 3 installed, prepare the pinned tools:

   ```sh
   mise trust
   mise install
   mise run bootstrap
   ```

The release job uses the repository's automatic `GITHUB_TOKEN` with Contents
write permission. The Homebrew job has only Contents read permission in pbj;
it downloads the release source with that token and exposes the separate tap
credential only to the formula-publication step. The default token cannot write
another repository. No Developer ID certificate, signing secret, notarization
credential, or GoReleaser Pro license is required.

A created tap is not proof that its token is installed or authorized. Missing
credentials, an unavailable tap, or rejected API writes fail publication rather
than silently skipping it. Do not push a release tag until these prerequisites
have been checked.

## Check and inspect a local snapshot

Run the complete repository checks and the release-specific gate:

```sh
mise run check
mise run release-check
mise run release-snapshot
```

`release-check` runs `goreleaser check`, `actionlint`, `go test ./internal/release`,
and the Python standard-library process-ownership regression test.
`release-snapshot` also runs that gate, then invokes GoReleaser with
`--snapshot --skip=publish --clean`. It builds both commands for `darwin/arm64`
and `darwin/amd64` on macOS, using cgo, the macOS 13 deployment target,
read-only dependencies, `-trimpath`, and an explicit version linker value.
It does not publish anything or create a version commit or tag.
Before the first Cog bump there is no `CHANGELOG.md`; snapshots omit it.
Once Cog creates it, binary archives include it through the standard changelog glob.

The ignored `dist/` directory contains:

- `pbj_0.0.0_source.tar.gz`, with the source under `pbj_0.0.0/`;
- `pbj_0.0.0_darwin_arm64.tar.gz` and
  `pbj_0.0.0_darwin_amd64.tar.gz`, each containing **both** `pbjcopy` and `pbjpaste`;
- checksums and GoReleaser's local metadata.

Snapshot version `0.0.0` is a deterministic **local-only verification fixture**,
not a release candidate. Never upload these archives or publish their formula.
The source archive comes from Git's committed source, so commit intended source
changes before checking the source-to-formula installation path. A snapshot is
not evidence that its uncommitted changes were archived or both architectures
were run.

To generate the fixture formula without publishing it:

```sh
mise exec -- go run ./scripts/release formula --tag v0.0.0 --source "$PWD/dist/pbj_0.0.0_source.tar.gz" --output "$TMPDIR/pbj.rb"
```

The generator streams SHA256 over the actual source archive. The fixture's
release URL is deliberately not a published download; local source installation
must point the formula at that archive. Building a formula locally does not
require Developer ID signing or notarization. See the
[verification records](agents/verification/README.md) for exercised installation
and runtime evidence rather than treating configuration as proof.

The renamed command source-install/name proof is **NOT RUN**. Its reserved report
path is `agents/verification/evidence/rename-2026-11-09/report.json`; obsolete
captures were retired rather than relabelled as a fresh run. Application-paste
outcomes remain **UNEXERCISED AFTER RENAME** and need separate clipboard consent.

For an opt-in source-built Homebrew proof, use the repository driver:

```sh
python3 scripts/verify-release.py --doctor
python3 scripts/verify-release.py --source dist/pbj_0.0.0_source.tar.gz --allow-install --evidence .scratch/release-proof-NEW
```

Choose a new evidence directory for each run. The driver uses a temporary local
tap and loopback source URL, refuses to replace an existing pbj installation,
and removes its installed formula and temporary tap afterward. This is local
verification, not publication; `--doctor` reports prerequisites before the
explicit installation opt-in. Running these commands is not a substitute for
reviewing their resulting evidence.

## Bump, review, and push

Start from the intended release branch with all intended changes committed and
a clean working tree. Fetch full history and tags before selecting the version:

```sh
git fetch origin --tags
```

For the first intentional version, run the checks **before** asking Cog to create
`0.1.0` and its `v0.1.0` tag:

```sh
mise run check
mise run release-check
mise exec -- cog bump --version 0.1.0
```

For subsequent versions, let Conventional Commits select the next version:

```sh
mise run release-bump
```

`release-bump` runs `check`, then `release-check`, then `cog bump --auto`.
Cog creates the changelog/version commit and a local `v`-prefixed tag. It does
**not** push; there are no auto-push bump hooks. Review the generated changelog,
version commit, and selected tag before proceeding. Bumping and pushing are
intentional release operations, not automatic side effects of ordinary commits.

For an intentional `v0.1.0` release, validate and inspect the exact notes:

```sh
mise exec -- go run ./scripts/release validate --tag v0.1.0
mise exec -- go run ./scripts/release notes --tag v0.1.0
```

Substitute the selected stable tag for later releases. Push the version commit
and that explicit tag with your normal authenticated Git client:

```sh
git push origin HEAD
git push origin v0.1.0
```

Do not create the tag in a workflow with `GITHUB_TOKEN`: those pushes do not
chain into this tag-triggered release workflow.

## What publication does

The tag-triggered macOS job checks out full history and tags, validates the tag,
runs `mise run check` and `mise run release-check`, and selects **that exact tag's**
section from committed `CHANGELOG.md`. It does not regenerate notes after Cog's
version-bump commit, which would add that commit to the body. Notes are written under
`RUNNER_TEMP`, outside the worktree and `dist/`, and passed to GoReleaser with
`--release-notes`. Keep the changelog pipe enabled: disabling it also discards
supplied notes. GitHub-generated notes are not used.

For `v0.1.0`, GoReleaser publishes `pbj_0.1.0_source.tar.gz`, prefixed with
`pbj_0.1.0/`, two optional-to-use macOS binary archives containing both commands,
and checksums. The archives are unsigned and not notarized. Directly downloaded
executables may be blocked by macOS security policy; this project does not
instruct users to bypass quarantine or Gatekeeper. Prefer the source-built
Homebrew formula instead. There is no cask, bottle, signing, or notarization
publication path.

Only after the GitHub release job succeeds does the separate Homebrew job
check out the release tag, download **that tag's source asset** into
`RUNNER_TEMP`, calculate its SHA256, generate the formula, and publish
`Formula/pbj.rb` to the tap's default branch. The formula downloads source from
`https://github.com/the-inconvenience-store/pbj/releases/download/v0.1.0/pbj_0.1.0_source.tar.gz`,
uses Go only as a build dependency, and builds both commands with explicit
release versions. It does not install GoReleaser's prebuilt binaries.
Content-identical publication is a no-op; changed content is updated against the
current file SHA with a bot commit identity.

After the formula is published, users can install it with:

```sh
brew install the-inconvenience-store/tap/pbj
pbjcopy --version
pbjpaste --version
```

## Resume only Homebrew publication

If the GitHub release succeeded but the Homebrew job failed, correct the tap or
token problem and use **Re-run failed jobs** on that workflow run. This retries
the separate formula job without rerunning GoReleaser or replacing uploaded
release assets. Do not recreate or move the published tag.

Alternatively, use the Release workflow's **Run workflow** control on the default
branch and supply the existing stable tag in its `tag` field. With GitHub CLI:

```sh
gh workflow run release.yml --repo the-inconvenience-store/pbj --ref main -f tag=v0.1.0
```

Dispatch is **tap-only**: it validates the stable tag before downloading anything,
checks out the full tagged source, requires an existing published non-prerelease
GitHub release and its exact source asset, then republishes the formula. It
cannot create or repair the GitHub release itself. Repeating publication with
identical source/formula content creates no extra commit. A retry of an older
release is rejected if the tap already offers a newer stable version.

For an explicitly authorized manual tap publication from the corresponding
checked-out release tag, download its source and invoke the same helper:

```sh
mise exec -- go run ./scripts/release validate --tag v0.1.0
gh release download v0.1.0 --repo the-inconvenience-store/pbj --pattern pbj_0.1.0_source.tar.gz --output "$TMPDIR/pbj_0.1.0_source.tar.gz"
mise exec -- go run ./scripts/release publish --tag v0.1.0 --source "$TMPDIR/pbj_0.1.0_source.tar.gz"
```

The last command requires `HOMEBREW_TAP_TOKEN` in the environment. Obtain it
through the approved secret mechanism; never paste it into this document,
command history, or a commit. A failed GitHub-release job is a different failure:
inspect its logs and release state before taking any publication action.

## Configuration references

- [GoReleaser source archives](https://goreleaser.com/customization/package/source/)
- [Go builds](https://goreleaser.com/customization/builds/builders/go/)
- [Binary archives](https://goreleaser.com/customization/package/archives/)
- [SCM releases](https://goreleaser.com/customization/publish/scm/)
- [Snapshots](https://goreleaser.com/customization/publish/snapshots/)
- [Cog version bumps](https://docs.cocogitto.io/guide/bump.html) and
  [changelogs](https://docs.cocogitto.io/guide/changelog.html)
- [GitHub Contents API](https://docs.github.com/en/rest/repos/contents#create-or-update-file-contents)
