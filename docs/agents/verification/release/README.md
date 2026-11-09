# Release and Homebrew verification

## Surface

Consumers touch Cog-generated notes/version tags, GitHub release assets, and the
source-built `the-inconvenience-store/tap/pbj` formula. `.goreleaser.yaml`,
`cog.toml`, `scripts/changelog.tera`, `scripts/release`, `internal/release`, and
`.github/workflows/release.yml` own this surface. This is not a cask or a
Developer ID/signing/notarization route. Existing clipboard runtime behavior is
covered by the separate CLI surfaces.

**Rename outcome: VERIFIED.** [Source installation](../evidence/rename-2026-11-09/report.json)
and [actual archive execution](../evidence/rename-2026-11-09/archives.json) confirm
the renamed commands. Obsolete captures were retired, not relabelled as fresh proof.
GitHub publication has not run. `HOMEBREW_TAP_TOKEN` is present in pbj secrets;
its write authorization remains unexercised. The local driver never publishes.

## Launch

Use macOS 13+, Apple Command Line Tools, Python 3, an already installed Homebrew,
and the repository's pinned mise tools. Keep all other Homebrew operations idle.
The orchestrator first commits the intended source (source archives use committed
Git content), then runs:

```sh
mise run check
mise run release-check
mise run release-snapshot
```

Readiness is `dist/pbj_0.0.0_source.tar.gz`, a real GoReleaser snapshot with source
under `pbj_0.0.0/`. Do not synthesize an archive to satisfy the driver. `0.0.0` is
only a deterministic local fixture; never publish it. The driver owns an ephemeral
loopback HTTP server and one `/tmp/pbj-release-verify-*` scratch directory. It
stops the server and removes its own state in `finally`.

## Doctor

```sh
python3 scripts/verify-release.py --doctor
```

Expected: JSON with `"doctor": "ready"`, tool versions, helper readiness, and
successful scratch cleanup. Doctor invokes read-only version/help operations for
Homebrew, mise, Cog, GoReleaser, and Go, then the actual Go helper's `--help` and
`validate --tag v0.0.0`. It checks prefix/Cellar writability and detects Homebrew's
`--skip-link` support. It does not install a formula, create a tap, inspect the
clipboard, publish remotely, or run `mise install`. Go compilation and tool
scratch/cache are permitted inside its owned temporary directory.

Homebrew's usual cache may be root-owned. Every driver subprocess instead receives
owned `HOMEBREW_CACHE`, `HOMEBREW_LOGS`, and `HOMEBREW_TEMP` directories. Never
repair shared cache permissions or run Homebrew as root. After a surprising result, rerun Doctor
before another drive. Tool/provisioning failures are not installation proof.

The driver also uses scratch-owned `XDG_CONFIG_HOME`. On Homebrew versions with
explicit tap trust, it trusts only the generated local tap there; the user's
trust configuration is unchanged and the temporary trust is removed with scratch.

## Drive

Obtain consent for an actual source install and possible retained dependency
additions. From the repository root, choose a **new** evidence directory:

```sh
python3 scripts/verify-release.py \
  --source dist/pbj_0.0.0_source.tar.gz \
  --allow-install \
  --evidence ".scratch/release-proof-$(date +%Y%m%d-%H%M%S)"
```

The driver uses `mise exec -- go run ./scripts/release formula --tag v0.0.0`
with the actual archive, not a reimplemented formula. It creates a unique
`pbj-local/release-<random>` tap directory (a local `tap-new` equivalent without
Git initialization or personal identity). It checks the generated formula with
Homebrew's real loader and style command. It replaces **only the generated URL**
with its owned loopback archive URL, retaining exact SHA256, version, build, and
test logic; it loads/styles that formula too. It actually runs
`brew install --build-from-source`, `brew test --force`, both installed commands' version
and help, and invalid-argument smoke checks. No current clipboard contents are
read or written.

`--skip-link` is preferred when available. A pre-existing `pbj` Cellar path always
aborts, because taps share that path. Without skip-link support, existing
Homebrew-prefix `pbjcopy`/`pbjpaste` files or links abort before installation. The driver
does not replace unrelated executables or force installation/linking. The test-only
`--force` permits testing the unlinked keg; it does not overwrite commands.
The private installed paths in the helper are implementation details, not user
command instructions. Installed users invoke `pbjcopy` and `pbjpaste` normally.

## Evidence

`report.json` records commands, deadlines, exits, sanitized output, archive SHA256,
source layout, loader/style results, actual brew actions, binary version/smoke,
and cleanup. `generated-pbj.rb` preserves the canonical generated formula;
`local-pbj.rb` preserves its URL-only adaptation. `install-receipt.json` must show
`poured_from_bottle: false`, `built_as_bottle: false`, and the owned tap. Captured
Homebrew build logs are retained under `homebrew-logs/` when available.

Evidence must be outside the driver's cleanup-owned scratch and must not already
exist. Home, repository, evidence, and scratch prefixes, Git identity/signature
lines, and emails are redacted; credentials are neither printed nor needed, and
secret/token/password environment variables are removed from subprocesses.

A `VERIFIED` report proves **only this native local source installation**, not
GitHub upload/publication, author-free notes, a non-native binary run, or macOS 13
runtime compatibility. All mapped capabilities require their own evidence. Do not
rename a local snapshot proof into a published-release proof.

## Cleanup

The driver stops only its loopback server, uninstalls only a `pbj` keg whose
receipt identifies its unique tap, removes that exact tap directory and its own
scratch/cache/log/temp, and preserves evidence. It checks original prefix command
links remain unchanged. `scratch_removed` and `tap_removed` must be true and
`cleanup.errors` empty for success. It will report rather than delete a keg whose
ownership cannot be established after an interrupted/failed installation.

Dependency additions and updates are reported in `added_dependencies_retained`
and never uninstalled. Homebrew may update a build dependency even with automatic
updates/install upgrades disabled; this is not a promise to freeze existing dependencies.
Homebrew style can also populate its own Ruby/bundler tooling. Never run a global
upgrade, cleanup, or autoremove as verification. Read-only directories inside the
owned scratch are made writable only as needed to delete that scratch; shared
cache permissions remain untouched.
If cleanup reports an ownership ambiguity, inspect the receipt and the exact
recorded tap before taking action; do not uninstall an unrelated user's `pbj`.

## Isolation

Run one release driver at a time and no other Homebrew install/link/uninstall
operations concurrently. Random taps and ephemeral ports isolate fixtures, but
Homebrew's Cellar and prefix remain shared. Clipboard and GUI drivers are not
needed. Evidence is separately durable. All child commands have deadlines; timed
out subprocess groups are terminated rather than left running.

## Feature map

[Every release capability belongs in the feature map](features/README.md).
It separates notes, version ownership, assets, source installation, and publication
boundaries. The parent orchestrator must run the checks/actual install, inspect
reports, and record outcomes; this assignment adds the reusable route without
running it. Authenticated publication requires explicit separate authorization and
provisioned credentials, not `--allow-install`.
