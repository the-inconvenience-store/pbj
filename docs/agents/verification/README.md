# Verification

## Surfaces

| Surface | Choose it for | Route |
| --- | --- | --- |
| pbjc CLI | Copying stdin, rich representations, file references, and CLI errors | [Launch, Doctor, Drive, Evidence, Cleanup, Isolation](pbjc/README.md) |
| pbjp CLI | Extracting exact/default representations, item metadata, file paths, and read-only CLI failures | [Launch, Doctor, Drive, Evidence, Cleanup, Isolation](pbjp/README.md) |
| Release / Homebrew | Author-free release notes, versioned source/assets, actual source-built installation, and publication boundaries | [Launch, Doctor, Drive, Evidence, Cleanup, Isolation](release/README.md) |


## Consent and evidence

The general clipboard is shared mutable state. Obtain explicit permission before
running any fixture-writing proof. Consent here covers general only, not the find,
ruler, or font boards. Run drivers serially; do not copy or interact with test
windows concurrently. Neither driver saves old clipboard data, and each leaves
its own harmless completion message afterward.
The pbjc app driver closes only generated documents/windows, not applications or
pre-existing user documents. The pbjp driver needs no application windows.

Payload reads follow generated fixture writes. Doctor and standalone information
checks read no clipboard payloads. Unexpected clipboard bytes are never included
in failure reports. pbjc window captures target only generated TextEdit and Preview
windows, never the desktop. Keep evidence outside cleanup-owned scratch.

Release verification uses separate `--allow-install` consent, never clipboard
consent. It installs an owned local-only formula from a real GoReleaser snapshot
and removes only its own installation/tap. It may leave newly installed build
dependencies; it never removes shared Go or changes cache permissions. Keep all
other Homebrew operations idle during that driver.

The [pbjc feature map](pbjc/features/README.md) and
[pbjp feature map](pbjp/features/README.md) own coverage. A passing build or test
suite is supporting evidence, not a substitute for driving the compiled command.

The [release feature map](release/features/README.md) separates local installation
proof from release-note/version/asset inspection and authenticated publication.
Local release-note selection, snapshot archives, and native source-built Homebrew
installation are **VERIFIED**. Authenticated GitHub release/tap publication has
not been run; a configured secret is not proof of its write authorization.

## Evidence

- [Current pbjc live run](evidence/pbjc-2026-11-09/report.json).
- [Implementation and verification notes](evidence/pbjc-2026-11-09/README.md).
- [Historical bootstrap evidence](evidence/bootstrap.md), before commands existed.
- [Current pbjp live run](evidence/pbjp-2026-11-09/report.json), **VERIFIED** on the general clipboard.
- [pbjp implementation and verification notes](evidence/pbjp-2026-11-09/README.md).
- [Current source-built Homebrew run](evidence/release-2026-11-09-verified/report.json), **VERIFIED**.
- [Release proof, cleanup, and publication limits](evidence/release-2026-11-09-verified/README.md).

Invoke verification with **Maintain** when flags, launch commands, native
publication, supported targets, application drivers, or feature outcomes change.
