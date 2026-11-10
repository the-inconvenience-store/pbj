# pbjpaste command reference

`pbjpaste` reads representations or file-reference paths from a macOS pasteboard
and writes them to stdout, or copies referenced files with `--to DIR`. It is the
paste companion to [`pbjcopy`](pbjcopy.md). The selected pasteboard is unchanged
on success and failure.

## Examples

Paste plain text, or the next supported default representation:

```sh
pbjpaste
```

Inspect the advertised types before choosing an exact representation:

```sh
pbjpaste --list-types
pbjpaste --type public.html > clipboard.html
pbjpaste --type public.png > clipboard.png
```

Choose one item from a pasteboard containing several items:

```sh
pbjpaste --item 0
pbjpaste --item 1 --type public.utf8-plain-text
```

Read file-reference paths, not file contents:

```sh
pbjpaste --files
pbjpaste --files --null > paths.nul
```

Copy referenced files and directories into an existing destination directory:

```sh
pbjcopy --file ./report.pdf --file ./images
mkdir ./pasted-files
pbjpaste --to ./pasted-files
```

Filesystem paste is explicit, silent on success, and never overwrites an existing
name. `--files` still prints paths; redirection still saves representation bytes.
Use `--to .` to select the current directory, or add `--item N` to copy only one
selected clipboard item.

Prefer RTF or read the find pasteboard using native-style options:

```sh
pbjpaste -Prefer rtf > clipboard.rtf
pbjpaste -pboard find
```

A preference can fall back to another supported representation; use `--type`
when the output must have one exact format. Shell redirection can create or
truncate the destination before `pbjpaste` runs, including when the command fails.

## Options and modes

| Option | Meaning |
| --- | --- |
| `-t TYPE`, `--type TYPE` | Read one exact advertised representation as raw bytes. |
| `--list-types` | List items and their advertised type identifiers as JSON. |
| `--item N` | Select a zero-based pasteboard item. |
| `--files` | Read ordered file-reference paths from all items, or the selected item. |
| `--to DIR` | Copy referenced files/directories into an existing directory without overwriting. |
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

The paste modes are mutually exclusive:

- **Default:** choose a supported representation using the preference order
  below. `--prefer` only applies to this mode.
- **Exact type:** `--type` reads only the identifier requested. It cannot be
  combined with `--prefer`, `--files`, or `--list-types`.
- **File references:** `--files` reads file URL paths, optionally with `--null`.
  It cannot be combined with `--type`, `--prefer`, or `--list-types`.
- **Type listing:** `--list-types` can be combined with `--pboard` and `--item`,
  but not `--type`, `--files`, `--prefer`, or `--null`.
- **Filesystem paste:** `--to DIR` copies referenced objects instead of writing
  stdout. It cannot be combined with `--type`, `--files`, `--prefer`,
  `--list-types`, or `--null`.

`--pboard` and `--item` also apply to default, exact-type, file extraction, and filesystem paste.
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
safety difference from native `pbpaste`: `pbjpaste` does not implicitly combine
items or choose one from a multi-item pasteboard. Select an item explicitly,
or use `--files` to request the deliberate multi-item file-reference behavior.
An index outside the available items is an error.

## Default selection and locale

Without `--type`, `--files`, `--list-types`, or `--to`, the default order is:

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
errors fail without falling back. `pbjpaste` itself performs no format conversion.

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
`pbjpaste` does not convert HTML, images, PDFs, or other representations to plain
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
is the decoded absolute path. Paths remain in pasteboard order; `pbjpaste` does
not check whether they exist, inspect filesystem metadata, or resolve symlinks.

Paths are output as UTF-8, independent of locale. By default each path has a
trailing newline. If any path contains a newline, the command fails before
writing paths to stdout and recommends `--null`, avoiding ambiguous records.
With `--null`, each path has a trailing NUL byte, including the last path.
This mode preserves path boundaries even when paths contain newlines.

## Filesystem paste

`--to DIR` copies local file references from Finder, `pbjcopy --file`, or
`pbjcopy --attachment` into an **existing directory**. `DIR` is required; use
`--to .` explicitly for the current directory. Relative destinations are relative
to the current working directory. Shell `~` expansion works as usual; quoted `~`
is a literal path, not expanded by `pbjpaste`. A destination symlink to a directory
is resolved before copying. Missing directories are not created.
Symlink/`..` components follow filesystem traversal semantics, not lexical path
cleaning. Source parent components are resolved while preserving the final
source symlink as a link.

All clipboard items are processed in order unless `--item N` selects one.
Every selected item must supply a valid local `public.file-url`, under the same
URL rules as `--files`. Empty clipboards, non-file items, invalid URLs, missing or
unreadable sources, invalid destinations, and copy failures return status 1 with
stderr diagnostics. Newlines, spaces, and Unicode in filenames need no separator
option: names are filesystem objects, not parsed stdout records.

Each object is copied under its source basename. There is **no overwrite,
directory merging, automatic renaming, or overwrite option**. Existing files,
directories, and symlinks—including dangling destination symlinks—are collisions.
Equal basenames from different source directories also collide. Repeating a paste
fails at the first existing name rather than silently replacing earlier results.
Copying a directory into itself or a descendant is rejected, including when a
destination symlink would otherwise disguise that relationship.

Regular files, directory trees, and symlinks are supported. Symlinks are copied
as links, not followed; relative links keep their original target text and may
therefore resolve differently after copying. Links may continue to point outside
the copied tree. Special objects such as FIFOs, sockets, and devices are not
supported. Native macOS `copyfile` requests data, POSIX metadata, ACLs, and extended
attributes, including resource forks. Supported metadata is subject to filesystem
and user permissions; ownership, set-id bits, hard-link relationships, and file
identities are not guaranteed. Source files are not frozen while copying:
concurrent source edits can affect the result.

### Failure and partial success

The operation stops at the first failure. **Earlier successfully copied objects
remain**; later clipboard items are not attempted. There is no multi-object
transaction, rollback, resume, retry, or restoration of destination contents.
The clipboard itself is never modified.

Each top-level object is first prepared in a private `.pbjpaste-*` staging
directory inside the destination, then published using native exclusive rename.
This prevents replacement even if a destination name appears during copying.
An individual tree is not published under its final name before copying
finishes. Owned staging is removed after normal completion/failure; cleanup
errors are reported and can leave a staging directory behind. A crash or forced
termination can also leave staging. Remove only staging known to belong to that
interrupted invocation, never unrelated files.

Publication does not guarantee crash durability or provide a filesystem
snapshot. Concurrent changes to source/destination trees are not isolated.
Filesystem paste does not convert images/rich text or extract embedded
attachments; it only copies referenced filesystem objects.

## Read consistency and failures

`pbjpaste` snapshots pasteboard metadata and fetches only the payloads needed for
the requested operation. It detects pasteboard changes during a read and fails
rather than combining metadata or payloads from different generations. It does
not retry, restore old contents, or write to the pasteboard.

Errors are reported on stderr. Output write failures can leave partial stdout;
stdout is not a transactional destination. Reading does not create history,
archives, background watchers, or a daemon.

| Exit status | Meaning |
| --- | --- |
| `0` | Successful extraction, filesystem paste, type listing, help, or version; also empty default output. |
| `1` | Operational failure, including missing data/references, unavailable items, locale encoding, native reads, pasteboard changes, output errors, or filesystem copy/preparation/publication/cleanup failures. |
| `2` | Invalid command-line usage. |

## Build and runtime

See the [repository README](../README.md) for setup and build commands.
`mise run build` produces both `pbjcopy` and `pbjpaste`. Source builds require
macOS, Apple Command Line Tools, and the Go toolchain managed by mise; the
executables use cgo and Objective-C/AppKit.

Runtime requires macOS 13 or newer, with no Go installation, compiler, helper
executable, or daemon. Supported targets are `darwin/arm64` and `darwin/amd64`;
this declaration is not evidence that both have been exercised.
