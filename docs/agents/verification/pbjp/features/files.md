# Extract ordered file paths

## User outcome

`--files` emits every generated file URL item's UTF-8 filesystem path in native
order, normally with a trailing newline per path. `--null` instead supplies a NUL
after every path, including the last, so spaces and embedded newline characters
remain unambiguous. `--files --item N` limits extraction to one item. A newline in
any selected path makes ordinary newline mode fail before emitting any stdout,
with guidance to use `--null`.

## How to reach it

Build, pass Doctor, and authorize clipboard replacement. The driver creates only
its own files under one unique `/tmp/pbjp-verify-*` directory and publishes them
using `bin/pbjc --file FIRST --file SECOND`.

Source entry points: `cmd/pbjp`; `internal/command.Command.Paste` (file mode,
URL decoding, item selection, and output buffering);
`internal/pasteboard/snapshot_darwin.go` (exact selected `public.file-url` reads);
and the public pbjc fixture publication path in
`internal/pasteboard/files_darwin.go` and `pasteboard_darwin.m`.

## How to drive it

```sh
python3 scripts/verify-pbjp.py --allow-clipboard --feature files \
  --evidence ".scratch/pbjp-files-$(date +%Y%m%d-%H%M%S)"
```

## Proof

Two ordered file references survive pbjc exit and yield exact paths, including a
space-containing name, in newline and trailing-NUL modes. Selecting either item
returns only its path; metadata filtering retains the original item index and
exact selected `public.file-url` extraction preserves URL bytes without adding
a newline. Out-of-range selection is exit 1. Without selection, default/exact
multi-item extraction is rejected while deliberate `--files` enumeration works.

After deleting only its own generated newline-named target, the driver extracts
the same reference successfully: extraction does not require target existence.

A safe first path followed by a newline-containing second path produces exit 1,
zero stdout, and a `--null` diagnostic in newline mode. NUL mode produces both
unaltered paths with a final NUL. Selecting just the safe item succeeds even
though an unselected item has a newline. Every read/failure preserves generated
native file URL payloads, type metadata, and change count. Expected byte counts
and hashes are evidence; no scratch absolute paths or unexpected bytes are saved.

## Gotchas

`--null` is only valid with `--files`. Missing file URL data fails rather than
silently skipping items. The failure group proves the missing-format case.
No Finder window, actual file copy, or file-content read by pbjp is part of this
extraction capability. Scratch fixture files are removed after proof; the report
lives elsewhere. Do not replace these generated fixtures with user files.
