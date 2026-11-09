# Verification

## Surfaces

| Surface | Choose it for | Route |
| --- | --- | --- |
| pbjcopy CLI | Copying stdin, rich representations, file references, and CLI errors | [Launch, Doctor, Drive, Evidence, Cleanup, Isolation](pbjcopy/README.md) |
| pbjpaste CLI | Extracting exact/default representations, item metadata, file paths, and read-only CLI failures | [Launch, Doctor, Drive, Evidence, Cleanup, Isolation](pbjpaste/README.md) |
| Release / Homebrew | Author-free release notes, versioned source/assets, actual source-built installation, and publication boundaries | [Launch, Doctor, Drive, Evidence, Cleanup, Isolation](release/README.md) |


## Consent and evidence

The general clipboard is shared mutable state. Obtain explicit permission before
running any fixture-writing proof. Consent here covers general only, not the find,
ruler, or font boards. Run drivers serially; do not copy or interact with test
windows concurrently. Neither driver saves old clipboard data, and each leaves
its own harmless completion message afterward.
The pbjcopy app driver closes only generated documents/windows, not applications or
pre-existing user documents. The pbjpaste driver needs no application windows.

Payload reads follow generated fixture writes. Doctor and standalone information
checks read no clipboard payloads. Unexpected clipboard bytes are never included
in failure reports. pbjcopy window captures target only generated TextEdit and Preview
windows, never the desktop. Keep evidence outside cleanup-owned scratch.

Release verification uses separate `--allow-install` consent, never clipboard
consent. It installs an owned local-only formula from a real GoReleaser snapshot
and removes only its own installation/tap. It may leave newly installed build
dependencies; it never removes shared Go or changes cache permissions. Keep all
other Homebrew operations idle during that driver.

The [pbjcopy feature map](pbjcopy/features/README.md) and
[pbjpaste feature map](pbjpaste/features/README.md) own coverage. A passing build or test
suite is supporting evidence, not a substitute for driving the compiled command.

The [release feature map](release/features/README.md) separates local installation
proof from release-note/version/asset inspection and authenticated publication.
The renamed commands' source installation and name checks are **NOT RUN** pending
fresh proof. Authenticated GitHub release/tap publication has not been run; a
configured secret is not proof of its write authorization.

## Evidence

- Fresh rename/source-install proof is **NOT RUN**. Its reserved report path is
  `evidence/rename-2026-11-09/report.json`; it is not an observed result yet.
- Obsolete command and release captures were retired rather than rewritten to
  imply that the renamed commands ran.
- Actual TextEdit, Preview, and Finder paste proof is **UNEXERCISED AFTER RENAME**.
  Earlier GUI observations are not fresh evidence for the renamed executables.
- [Historical bootstrap evidence](evidence/bootstrap.md) predates command
  implementation and is not runtime proof.

Invoke verification with **Maintain** when flags, launch commands, native
publication, supported targets, application drivers, or feature outcomes change.
