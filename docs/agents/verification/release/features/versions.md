# Intentional version ownership

## User outcome

Cog creates one deliberate stable version, changelog commit, and `v`-prefixed tag.
Both installed commands report that same explicit version instead of `dev` or a
GoReleaser-selected version. A bump does not silently push anything.

## How to reach it

Use full Git history/tags and a clean disposable fixture for bump proof. Stable
release tags are exactly `vMAJOR.MINOR.PATCH`; prereleases and metadata suffixes
are outside the supported route. `cog.toml` owns bump/tag behavior;
`scripts/release` owns validation, committed-note selection, and formula entry
points. The snapshot uses the local-only fixed `0.0.0` version, never a release candidate.

## How to drive it

In the isolated fixture, the parent runs the actual Cog first-version command:

```sh
mise exec -- cog bump --version 0.1.0
```

Inspect the resulting version commit, `CHANGELOG.md`, and local `v0.1.0` tag using
sanitized Git reads (no raw signed-log evidence). Confirm no remote push occurred.
Use the public helper CLI to validate stable and unsupported examples:

```sh
mise exec -- go run ./scripts/release validate --tag v0.1.0
mise exec -- go run ./scripts/release validate --tag v0.1.0-rc.1
mise exec -- go run ./scripts/release validate --tag 0.1.0
```

The latter two must fail. The source-install driver exercises `v0.0.0` and checks
installed command output `pbjcopy 0.0.0` and `pbjpaste 0.0.0`. On an authorized installed
release, use bare public commands:

```sh
pbjcopy --version
pbjpaste --version
```

## Proof

A real Cog bump produces the intended stable tag/version exactly once without an
auto-push hook. Stable helper validation succeeds; unsupported tags fail before
network mutation. Both actual installed binaries report the explicit selected
version. Capture tag names, exit statuses, and version output without Git identity
or signature fields. Matching local fixture versions are not published-release
version proof.

## Gotchas

**Installed snapshot versions VERIFIED:** [source install](../../evidence/rename-2026-11-09/report.json)
and [archive execution](../../evidence/rename-2026-11-09/archives.json) report
`pbjcopy 0.0.0` and `pbjpaste 0.0.0`. Cog bump and stable publication were not
rerun for this name-only change. A Cog bump in the live
checkout is a release action, not a harmless check. Do not move or recreate tags,
create a snapshot release, or create tags from a GitHub-token workflow to chain
another workflow. The intentional real bump/push remains release-operator owned.
