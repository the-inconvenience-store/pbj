# Release feature map

Renamed command source installation, snapshot archives, and authenticated
**v0.1.0 publication are VERIFIED**. Published downloads and all four binary
executions are recorded separately from the local snapshot proof.
Use the [surface](../README.md) for consent, launch, Doctor, evidence, and cleanup.

| Capability | User outcome | Proof route | Outcome |
| --- | --- | --- | --- |
| [Author-free notes](notes.md) | Read exact-tag notes without contributor identities | Real Cog/helper CLI + published body inspection | Published v0.1.0 body VERIFIED |
| [Version ownership](versions.md) | Deliberate Cog tag and explicit command versions | Cog bump + downloaded executable versions | v0.1.0 and all four published command versions VERIFIED |
| [Release assets](assets.md) | Source and two-command macOS archives | Published download/checksum/execution inspection | Published v0.1.0 VERIFIED on arm64 and Rosetta |
| [Source-built installation](source-install.md) | Build/install/test both native commands from source | Actual `scripts/verify-release.py` install | Local snapshot VERIFIED; published v0.1.0 not locally installed |
| [Publication/recovery](publication.md) | Publish/retry the formula separately | Authorized GitHub workflow/API evidence | First publication VERIFIED; retry/no-op not exercised |

The source driver is intentionally not an all-features pass: its report lists the
remaining notes, bump, remote publication, architecture, and OS coverage gaps.
[Published v0.1.0 evidence](../../evidence/v0.1.0-publication/report.json) proves
the hosted jobs, committed-note equality, downloads, binary versions, and bot tap
commit. The local source proof does not imply clipboard/application-paste coverage.
Recovery dispatch and repeated identical publication remain unexercised.
