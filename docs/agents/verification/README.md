# Verification

## Surfaces

| Surface | Choose it for | Route |
| --- | --- | --- |
| pbjcopy CLI | Copying stdin, rich representations, file references, and CLI errors | [Launch, Doctor, Drive, Evidence, Cleanup, Isolation](pbjcopy/README.md) |
| pbjpaste CLI | Extracting representations/metadata/paths, copying references to disk, and failure safety | [Launch, Doctor, Drive, Evidence, Cleanup, Isolation](pbjpaste/README.md) |
| Release / Homebrew | Author-free release notes, versioned source/assets, actual source-built installation, and publication boundaries | [Launch, Doctor, Drive, Evidence, Cleanup, Isolation](release/README.md) |


## Consent and evidence

The general clipboard is shared mutable state. Obtain explicit permission before
running any fixture-writing proof. Consent here covers general only, not the find,
ruler, or font boards. Run drivers serially; do not copy or interact with test
windows concurrently. Neither driver saves old clipboard data, and each leaves
its own harmless completion message afterward.
The pbjcopy app driver closes only generated documents/windows, not applications or
pre-existing user documents. The pbjpaste driver needs no application windows.
Mail compose-window proof requires separate `--allow-mail` consent. It creates
an unsent message, pastes only generated attachments, and discards that message.

Payload reads follow generated fixture writes. Doctor and standalone information
checks read no clipboard payloads. Unexpected clipboard bytes are never included
in failure reports. pbjcopy captures target generated TextEdit, Preview, and Mail
windows by WindowServer ID, never a desktop region. Mail captures are cropped to
the generated message body. Keep evidence outside cleanup-owned scratch.

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
The renamed commands' source installation, executable identities, and information
paths are **VERIFIED**. Authenticated `v0.1.0` GitHub release/tap publication and
downloaded asset/version checks are also **VERIFIED**, including the tap write.

## Evidence

- [Published v0.2.0 release and Homebrew formula](evidence/v0.2.0-publication/report.json),
  **VERIFIED**: both hosted jobs, exact-tag notes, downloaded checksums/source
  layout, native arm64 and Rosetta x86_64 command smoke, and the canonical tap formula.
- [Filesystem paste and full paste regression proof](evidence/issue-1-filesystem-paste/report.json),
  **VERIFIED**: 160 real CLI actions covering no-clobber copying, recursive contents,
  metadata/symlinks, physical parent traversal, partial success, and unchanged
  clipboard generations/payloads. General clipboard reset and scratch cleanup succeeded.
- [Stdin file attachments and Mail paste](evidence/issue-2-attachments/report.json),
  **VERIFIED**: exact post-exit backing bytes, names, private permissions,
  retention, failure preservation, and a named attachment in an unsent message.
  [Body-only visual proof](evidence/issue-2-attachments/mail-attachment.png).
- [Representation regressions](evidence/issue-2-representations/report.json) and
  [safe failures](evidence/issue-2-safe-failures/report.json), **VERIFIED**.
  All issue #2 runs reset the clipboard and removed owned scratch/cache files.
- [Structured-input detection and exact CLI extraction](evidence/issue-3-structured-input/report.json),
  **VERIFIED**: JSON, HTML, XML, precedence, explicit overrides, fallback, and ambiguous input.
- [Copy preparation failure preservation](evidence/issue-3-safe-failures/report.json),
  **VERIFIED**. Both issue #3 runs reset the clipboard and removed owned scratch.
- [Fresh rename/source-install proof](evidence/rename-2026-11-09/report.json), **VERIFIED**.
- [Archive identities and both architecture runs](evidence/rename-2026-11-09/archives.json), **VERIFIED**.
- [Copy information](evidence/rename-2026-11-09/pbjcopy-information/report.json) and
  [paste information](evidence/rename-2026-11-09/pbjpaste-information/report.json), **VERIFIED**.
- [Rename evidence and limits](evidence/rename-2026-11-09/README.md).
- [Published v0.1.0 workflow, notes, assets, and tap proof](evidence/v0.1.0-publication/report.json), **VERIFIED**.
- Obsolete command and release captures were retired rather than rewritten to
  imply that the renamed commands ran.
- Actual TextEdit, Preview, and Finder paste proof is **UNEXERCISED AFTER RENAME**.
  Earlier GUI observations are not fresh evidence for the renamed executables.
- [Historical bootstrap evidence](evidence/bootstrap.md) predates command
  implementation and is not runtime proof.

Invoke verification with **Maintain** when flags, launch commands, native
publication, supported targets, application drivers, or feature outcomes change.
