# Inspect metadata and select items or boards

## User outcome

`--list-types` returns JSON `[{"item":0,"types":[...]}]` with native per-item type
order, or `[]` for an empty board. `--item N` selects a zero-based native item and
may filter metadata without renumbering it. `--pboard general|find|ruler|font`
(including native `-pboard`) chooses the system board. Default/exact extraction
rejects multiple items without a selection instead of concatenating them.

## How to reach it

Pass Doctor and authorize the [representation](representations.md) and
[file](files.md) fixture groups. Source entry points: `cmd/pbjp`,
`internal/command.Command.Paste`, `internal/pasteboard/snapshot_darwin.go`
(`Capture`, `Types`, `Data`), and `internal/pasteboard/pasteboard_darwin.m`
(native board selection and generation checks).

## How to drive it

```sh
python3 scripts/verify-pbjp.py --allow-clipboard --feature all \
  --evidence ".scratch/pbjp-items-$(date +%Y%m%d-%H%M%S)"
```

For focused runs, use `--feature representations` for zero/one-item metadata and
both general-board flag spellings, or `--feature files` for multiple-item metadata
and selection. Every run needs its own new evidence directory.

## Proof

Listing equals an independent native metadata enumeration for raw, HTML/plain,
empty text, genuinely empty board, and ordered file items. Filtering item 1
returns an object whose `item` remains 1. The general board yields its generated
text with native `-pboard general`, and both extraction and listing combine
`--pboard general` with `--item 0`. Missing indexes are exit 1 with no stdout. Negative
and nonnumeric indexes are usage errors, exit 2. Both default and exact
multi-item reads without `--item` fail instead of merging items; selected exact
file URL bytes match the corresponding generated item.

The native snapshot entry point captures metadata and accesses selected payloads.
For an RTF-only publication on the recorded host, an independent observer found
advertised UTF-8 data unavailable after writer exit. `pbjp --list-types` still
succeeded; exact UTF-8 failed; default reached readable RTF. This closes the
unavailable-payload listing gap without relying on an eager all-data fixture.
Native tests separately cover rejection when a writer replaces a capture.

## Gotchas

`--list-types` is clipboard metadata, not standalone help. It may combine only
with pboard/item, not type/files/prefer/null. The failure group covers those
invalid combinations. Read-only metadata does not grant permission to read or
save previous user payloads. Only the general board is read or overwritten and
then reset to generated harmless text. The supported find/ruler/font selection
paths remain **unexercised** because consent covers general only; do not inspect
or save their pre-existing metadata or payloads.
