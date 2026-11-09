# Source-built Homebrew installation

## User outcome

Homebrew checksums source and compiles both native commands with the explicit
release version, cgo/AppKit, macOS 13 deployment target, `-trimpath`, and read-only
Go dependencies. Users need no Developer ID signing, cask, bottle, or quarantine
workaround.

## How to reach it

Follow the [surface Launch and Doctor](../README.md). Obtain separate
`--allow-install` consent, supply the actual GoReleaser snapshot archive, and
ensure there is no installed `pbj` keg. Homebrew's prefix/Cellar must already be
writable. Shared Go and existing command links belong to the user, not the proof.

## How to drive it

```sh
python3 scripts/verify-release.py --doctor
python3 scripts/verify-release.py \
  --source dist/pbj_0.0.0_source.tar.gz \
  --allow-install \
  --evidence ".scratch/release-source-$(date +%Y%m%d-%H%M%S)"
```

Inspect `report.json`, both saved formula variants, `install-receipt.json`, and
captured Homebrew logs. The driver invokes the real Go formula generator with
`--tag v0.0.0`, loads/styles its result, then changes only the source URL to an
owned loopback server. The server supplies the unchanged actual snapshot archive.
Homebrew actually installs from source, runs the unchanged formula test, and
executes both installed binaries with help/version/invalid arguments. It uses
`--skip-link` when available and never overwrites user links.

## Proof

The generated canonical URL names the exact release tag/archive and its SHA256
matches the actual source bytes. Comparing saved formulas shows only the URL
changed; build/test/version/checksum logic remains identical. Homebrew loader/style
succeed, `brew install --build-from-source` and `brew test` succeed, both installed
versions are `0.0.0`, and the durable receipt says neither poured nor built as a
bottle. Help succeeds, invalid usage exits 2, and no clipboard payload is read or
written. Cleanup errors are empty, owned scratch/tap are gone, original command
links are preserved, and retained new dependencies are explicitly reported.

## Gotchas

Status is **pending** until the parent runs/inspects this driver. Local
`VERIFIED` is not GitHub release/tap publication proof. Never publish the snapshot
fixture. Existing `pbj` installation always aborts; without skip-link support,
conflicting prefix command links also abort. Never fix root-owned global caches
or uninstall shared Go: the driver uses owned temporary cache/log/temp locations.
Homebrew may retain newly installed dependencies and its style tooling, which are
reported instead of removed. Do not run other Homebrew mutation concurrently.
