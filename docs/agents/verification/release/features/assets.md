# Versioned source and binary assets

## User outcome

Consumers receive an exact-version source archive for Homebrew builds and optional
macOS architecture archives that each contain both commands. Formula consumers do
not download the optional compiled binaries.

## How to reach it

The parent commits intended changes before invoking the repository's pinned
GoReleaser on macOS. `.goreleaser.yaml` owns archive names, source prefix, targets,
cgo/deployment settings, and explicit linker versions. A snapshot is local-only;
it must never be uploaded as a release.

## How to drive it

```sh
mise run release-snapshot
tar -tzf dist/pbj_0.0.0_source.tar.gz
tar -tzf dist/pbj_0.0.0_darwin_arm64.tar.gz
tar -tzf dist/pbj_0.0.0_darwin_amd64.tar.gz
```

The source-install driver inspects the actual source archive's prefix and required
module/command sources and records its SHA256. The parent additionally inspects
the actual optional binary archive members and file architectures, then invokes
extracted commands only with `--version`/`--help` if native execution is available.
Use scratch extraction directories outside evidence and preserve sanitized
observations before cleanup. Never replace an absent archive with a hand-built
fixture or run clipboard copy/paste just to smoke these assets.

For a separately authorized published release, download its actual assets using
[release operations](../../../../releases.md), independently recalculate checksums,
and inspect those downloaded archives rather than reusing snapshot results.

## Proof

The exact source basename is `pbj_0.0.0_source.tar.gz`, with `pbj_0.0.0/` source
prefix, `go.mod`, `go.sum`, and both command entry points. Both binary archives
contain `pbjcopy` and `pbjpaste`; they target their stated Darwin architecture and embed
`0.0.0`. Real release names/prefixes/versions instead use the intentional release
version, for example `pbj_0.1.0_source.tar.gz` under `pbj_0.1.0/`. Observe the actual
published release asset list/checksums before claiming remote delivery.

## Gotchas

**Published v0.1.0 VERIFIED:** [actual downloaded archives, checksums, and executable runs](../../evidence/v0.1.0-publication/report.json).
Each archive contains exactly `pbjcopy` and `pbjpaste` as executable members.
All four downloaded binaries passed version/help/error checks: arm64 natively,
x86_64 through Rosetta. This is not native Intel-hardware or macOS 13 runtime
proof. Never bypass quarantine/Gatekeeper to execute a downloaded binary.
