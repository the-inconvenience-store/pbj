# Verification

## Surfaces

| Surface | Choose it for | Route |
| --- | --- | --- |
| pbjc CLI | Copying stdin, rich representations, file references, and CLI errors | [Launch, Doctor, Drive, Evidence, Cleanup, Isolation](pbjc/README.md) |

`pbjp` remains deferred; it has no executable or feature map.

## Consent and evidence

The general clipboard is shared mutable state. Obtain explicit permission before
running a copy proof. Run serially; do not change the clipboard or interact with
the test windows concurrently. The driver does not save existing clipboard data
and leaves `pbjc verification complete` afterward. It closes only its generated
documents/windows, not the applications or any pre-existing user documents.

All reads follow generated fixture writes. Unexpected clipboard bytes are never
included in failure reports. Window captures target only generated TextEdit and
Preview windows, never the desktop. Keep evidence outside cleanup-owned scratch.

The [feature map](pbjc/features/README.md) owns coverage. A passing build or test
suite is supporting evidence, not a substitute for driving the compiled command.

## Evidence

- [Current pbjc live run](evidence/pbjc-2026-11-09/report.json).
- [Implementation and verification notes](evidence/pbjc-2026-11-09/README.md).
- [Historical bootstrap evidence](evidence/bootstrap.md), before commands existed.

Invoke verification with **Maintain** when flags, launch commands, native
publication, supported targets, application drivers, or feature outcomes change.
