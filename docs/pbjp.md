# pbjp command reference

`pbjp` reads representations or file-reference paths from a macOS pasteboard
and writes them to stdout. It is the paste companion to [`pbjc`](pbjc.md).
Reading leaves the selected pasteboard unchanged, on success and failure.

## Examples

Paste plain text, or the next supported default representation:

```sh
bin/pbjp
```

Inspect the advertised types before choosing an exact representation:

```sh
bin/pbjp --list-types
bin/pbjp --type public.html > clipboard.html
bin/pbjp --type public.png > clipboard.png
```

Choose one item from a pasteboard containing several items:

```sh
bin/pbjp --item 0
bin/pbjp --item 1 --type public.utf8-plain-text
```

Read file-reference paths, not file contents:

```sh
bin/pbjp --files
bin/pbjp --files --null > paths.nul
```

Prefer RTF or read the find pasteboard using native-style options:

```sh
bin/pbjp -Prefer rtf > clipboard.rtf
bin/pbjp -pboard find
```

A preference can fall back to another supported representation; use `--type`
when the output must have one exact format. Shell redirection can create or
truncate the destination before `pbjp` runs, including when the command fails.

## Options and modes

| Option | Meaning |
| --- | --- |
| `-t TYPE`, `--type TYPE` | Read one exact advertised representation as raw bytes. |
| `--list-types` | List items and their advertised type identifiers as JSON. |
| `--item N` | Select a zero-based pasteboard item. |
| `--files` | Read ordered file-reference paths from all items, or the selected item. |
| `--null` | Use NUL-delimited paths instead of newline-delimited paths; requires `--files`. |
| `--pboard BOARD`, `-pboard BOARD` | Select `general`, `find`, `ruler`, or `font`. Default: `general`. |
| `--prefer FORMAT`, `-Prefer FORMAT` | Prefer `txt`, `rtf`, or `ps`; `ascii` is an alias for `txt`. |
| `-h`, `--help`, `-help` | Print generated command help without reading a pasteboard. |
| `-v`, `--version` | Print the version without reading a pasteboard. |

Every option is nonrepeatable, including repetitions through different aliases.
There are no positional operands. Unknown options, missing values, invalid
board names or preferences, invalid item indices, and incompatible modes are
usage errors rather than requests to guess. Help and version are standalone:
they cannot be combined with other options or each other.

The extraction modes are mutually exclusive:

- **Default:** choose a supported representation using the preference order
  below. `--prefer` only applies to this mode.
- **Exact type:** `--type` reads only the identifier requested. It cannot be
  combined with `--prefer`, `--files`, or `--list-types`.
- **File references:** `--files` reads file URL paths, optionally with `--null`.
  It cannot be combined with `--type`, `--prefer`, or `--list-types`.
- **Type listing:** `--list-types` can be combined with `--pboard` and `--item`,
  but not `--type`, `--files`, `--prefer`, or `--null`.

`--pboard` and `--item` also apply to default, exact-type, and file extraction.
`--null` is invalid without `--files`.

## Items and type listing

A pasteboard can contain several items, each offering several representations.
The distinction matters: representations are alternative formats of one item,
not additional items to concatenate.

`--list-types` emits JSON in pasteboard item order, with zero-based item indices
and the advertised identifiers for each item, for example:

```json
[{"item":0,"types":["public.utf8-plain-text","public.rtf"]},{"item":1,"types":["public.file-url"]}]
```

An empty pasteboard produces `[]`. With `--item`, only the selected item's
entry is listed, retaining its original index. Listing enumerates metadata;
it does not fetch representation payloads. It is an inspection of the selected
pasteboard, not a static registry or allowlist of recognized types.

Default and exact-type extraction require `--item` when there is more than one
item, even if only one item has the desired format. This is an intentional
safety difference from native `pbpaste`: `pbjp` does not implicitly combine
items or choose one from a multi-item pasteboard. Select an item explicitly,
or use `--files` to request the deliberate multi-item file-reference behavior.
An index outside the available items is an error.

## Default selection and locale

Without `--type`, `--files`, or `--list-types`, the default order is:

1. Plain text (`public.utf8-plain-text`).
2. EPS/PostScript (`com.adobe.encapsulated-postscript`).
3. RTF (`public.rtf`).

`--prefer txt` (or `ascii`) uses that order. `--prefer ps` tries EPS first,
then plain text, then RTF. `--prefer rtf` tries RTF first, then plain text,
then EPS. A preference changes selection order; it neither converts content
nor requires that format to exist.

macOS or an application may advertise generated representations whose payloads
are unavailable when read. Default and preference selection try the next
supported representation when a payload is unavailable. If supported formats
are advertised but all their payloads are unavailable, the command fails.
Native read errors, pasteboard changes, malformed text, and locale encoding
errors fail without falling back. `pbjp` itself performs no format conversion.

Selected plain UTF-8 text is encoded for the first nonempty locale setting in
this order:

1. `LC_ALL`
2. `LC_CTYPE`
3. `LANG`
4. `C` when none is set

An unknown locale name, or no nonempty locale setting, falls back to `C`.
Use a UTF-8 locale when UTF-8 output is required. Malformed text or text that
cannot be represented in the selected encoding is an error rather than
silently substituted output. Locale processing applies only to default
plain-text output; EPS and RTF retain their bytes.

An empty pasteboard, or an item with none of these supported default formats,
produces empty stdout and succeeds. No newline is added to extracted data.
`pbjp` does not convert HTML, images, PDFs, or other representations to plain
text, and does not perform OCR.

## Exact representations

`--type` accepts a nonempty pasteboard type identifier, including a custom
identifier. The selected item must advertise that exact identifier and supply
its payload. A missing type or unavailable payload is an error: there is no
implicit conversion, alternate-type lookup, or fallback to another representation.

The representation's bytes are written unchanged, including for familiar text
identifiers. Exact extraction bypasses locale encoding and does not append a
newline. The caller is responsible for interpreting the resulting bytes.

## File references

`--files` extracts filesystem paths from file URL representations
(`public.file-url`), not the contents of those files or directories. It
processes all pasteboard items in order unless `--item` selects a subset.
It does not treat plain text containing a path or an ordinary web URL as a
file reference. Missing file references, malformed file URL data, and invalid
file URLs are errors, not empty-success results or guessed paths.

Only local file URLs are accepted: the host must be empty or `localhost`
(case-insensitive), with no user information, query, or fragment. The output
is the decoded absolute path. Paths remain in pasteboard order; `pbjp` does
not check whether they exist, inspect filesystem metadata, or resolve symlinks.

Paths are output as UTF-8, independent of locale. By default each path has a
trailing newline. If any path contains a newline, the command fails before
writing paths to stdout and recommends `--null`, avoiding ambiguous records.
With `--null`, each path has a trailing NUL byte, including the last path.
This mode preserves path boundaries even when paths contain newlines.

## Read consistency and failures

`pbjp` snapshots pasteboard metadata and fetches only the payloads needed for
the requested operation. It detects pasteboard changes during a read and fails
rather than combining metadata or payloads from different generations. It does
not retry, restore old contents, or write to the pasteboard.

Errors are reported on stderr. Output write failures can leave partial stdout;
stdout is not a transactional destination. Reading does not create history,
archives, background watchers, or a daemon.

| Exit status | Meaning |
| --- | --- |
| `0` | Successful extraction, type listing, help, or version; also empty default output. |
| `1` | Operational failure, including missing exact data or file references, an unavailable item, locale encoding, native read, pasteboard change, or output failure. |
| `2` | Invalid command-line usage. |

## Build and runtime

See the [repository README](../README.md) for setup and build commands.
`mise run build` produces both `bin/pbjc` and `bin/pbjp`. Source builds require
macOS, Apple Command Line Tools, and the Go toolchain managed by mise; the
executables use cgo and Objective-C/AppKit. Copy `bin/pbjp` to a directory on
`PATH` to install it.

Runtime requires macOS 13 or newer, with no Go installation, compiler, helper
executable, or daemon. Supported targets are `darwin/arm64` and `darwin/amd64`;
this declaration is not evidence that both have been exercised.
