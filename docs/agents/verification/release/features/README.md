# Release feature map

Local Cog/helper notes, snapshot assets, and source installation are **VERIFIED**.
Remote GitHub release/formula publication has not been authorized or exercised.
Use the [surface](../README.md) for consent, launch, Doctor, evidence, and cleanup.

| Capability | User outcome | Proof route | Outcome |
| --- | --- | --- | --- |
| [Author-free notes](notes.md) | Read exact-tag notes without contributor identities | Real Cog/helper CLI + published body inspection | Local VERIFIED; published body unrun |
| [Version ownership](versions.md) | Deliberate Cog tag and explicit command versions | Isolated bump + validation + source driver | Local VERIFIED; stable publication unrun |
| [Release assets](assets.md) | Source and two-command macOS archives | Actual GoReleaser archive/checksum/execution inspection | Snapshot VERIFIED; remote download unrun |
| [Source-built installation](source-install.md) | Build/install/test both native commands from source | Actual `scripts/verify-release.py` install | VERIFIED on native arm64 |
| [Publication/recovery](publication.md) | Publish/retry the formula separately | Authorized GitHub workflow/API evidence | Not run; first public release approval required |

The source driver is intentionally not an all-features pass: its report lists the
remaining notes, bump, remote publication, architecture, and OS coverage gaps.
The tap exists and secret metadata confirms `HOMEBREW_TAP_TOKEN` is configured.
Neither fact proves token authorization or remote publication. The
[current evidence](../../evidence/release-2026-11-09-verified/README.md) records
the local proofs, native arm64 and Rosetta execution, cleanup, and exact limits.
