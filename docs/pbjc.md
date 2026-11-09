# pbjc command reference

`pbjc` publishes stdin bytes or file references to a macOS pasteboard. It is
inspired by [pbrich](https://github.com/waynehoover/pbrich). `pbjp`, the companion
paste command, is planned; it is not included.

## Examples

Copy text from stdin:

```sh
printf 'Hello, clipboard!\n' | bin/pbjc
```

Copy an image using signature detection (the filename is not consulted):

```sh
bin/pbjc < image.png
```

Offer HTML with a separate plain-text fallback:

```sh
printf '<strong>Hello</strong>' | bin/pbjc --type public.html --plain 'Hello'
```

Publish arbitrary bytes under explicit identifiers:

```sh
bin/pbjc --type com.example.payload --type com.example.alternate < payload.bin
```

Both identifiers receive exactly the same input bytes. These options do not
convert one format into another.

Copy file references rather than their contents:

```sh
bin/pbjc --file ./report.pdf --file ./images --file ~/Documents/notes.txt
```

Use a native `pbcopy` pasteboard selector:

```sh
printf 'Search text' | bin/pbjc -pboard find
```

These copy examples replace the selected pasteboard's contents. The default is
the general pasteboard.

## Options and modes

| Option | Meaning |
| --- | --- |
| `-t TYPE`, `--type TYPE` | Explicit pasteboard type identifier; repeatable. |
| `-p TEXT`, `--plain TEXT` | Separate plain-text fallback; at most once. |
| `-f PATH`, `--file PATH` | File reference; repeatable. Selects file mode. |
| `-pboard BOARD` | Select `general`, `ruler`, `find`, or `font`; at most once. Default: `general`. |
| `-h`, `--help`, `-help` | Print help without accessing a pasteboard. |
| `-v`, `--version` | Print the version without accessing a pasteboard. |
| `-l`, `--list-types` | Print common type identifiers without accessing a pasteboard. |

The common-types list is a convenience, not an exhaustive registry or an
allowlist. Explicit types may be any nonempty identifier, including custom
identifiers.

Value options accept a separate value or an equals sign, for example
`-t public.html`, `--type public.html`, `-t=public.html`, or
`--type=public.html`. Short options cannot be clustered (`-tf` is not shorthand
for two options). `--` ends option parsing, but any following positional
operands are still rejected.

There are two copy modes:

- **Stdin mode:** read stdin through EOF, then publish its representations.
  There is no arbitrary input-size limit. Memory and operating-system limits
  still apply. Empty stdin publishes empty plain text.
- **File mode:** one or more `--file` options publish file references. Stdin is
  not read. File mode cannot be combined with `--type` or `--plain`.

Help, version, and common-types listing are standalone operations: they neither
read stdin nor read or mutate a pasteboard. Combining an informational option
with copy options, or combining informational operations, is a usage error.
Unknown options, positional operands, missing option values, repeated singleton
options (including through different aliases), invalid board names, and mixed
copy modes are rejected rather than guessed. Use stdin redirection for content
and `--file` for file references; a bare filename is not an input option.

## Types, detection, and text

Without `--type`, `pbjc` recognizes PNG, JPEG, TIFF, PDF, RTF, and EPS from their
byte signatures. EPS recognition follows the native `pbcopy`-style PostScript
header convention. Detection is signature recognition, **not complete format
validation**: a recognized header does not prove that an image or document is
well formed or that another application can open it. There is no conversion,
rendering, or repair of rich content.

If no rich signature matches, stdin is decoded as plain text using the first
nonempty locale setting in this order:

1. `LC_ALL`
2. `LC_CTYPE`
3. `LANG`
4. `C` when none is set

Malformed text for the selected locale is an error, not silently replaced with
substitution characters. Select a UTF-8 locale when supplying UTF-8 text. The
locale governs plain-text decoding; it is not a request to transcode rich data.

HTML, JSON, XML, and URLs require explicit types; they are not inferred from
markup, syntax, or a URL-looking string. For example, use `--type public.html`
for HTML or `--type public.url` for a URL representation.

Supplying any `--type` disables automatic detection and locale decoding for the
stdin payload. Each distinct identifier receives the original bytes unchanged;
repeated occurrences of the same identifier are deduplicated. Explicit raw types
do not trigger content validation, even for a familiar text or image identifier.
The caller is responsible for matching the bytes to the advertised type.

`--plain TEXT` adds a separate plain-text representation; it never extracts text
from HTML, RTF, a PDF, or any other input. It is rejected if the input already
publishes the same plain-text type (`public.utf8-plain-text`), whether selected
explicitly or by automatic plain-text detection. This avoids overwriting one
representation with another. An empty fallback string is allowed.

## File references

`--file` accepts existing files, directories, and symlink paths whose targets
exist. Broken symlinks and nonexistent paths are errors. `pbjc` does not read
file contents to publish a reference.

A leading `~/` is expanded to the current user's home directory. Relative paths
are made absolute and lexically cleaned (for example, redundant separators and
`.` segments are removed). Duplicate normalized paths are removed while
preserving first-occurrence order. Symlinks are not resolved to their targets:
the selected path is preserved. Two different symlink paths to the same target
therefore remain distinct references. Lexical normalization, not filesystem
identity, determines deduplication.

## Publication, lifetime, and failures

Input reading, option and content preparation, locale decoding, and file-path
preparation finish **before publication begins**. A read, parse, or preparation
error leaves the existing pasteboard untouched.

Publication itself is not transactional. Once it begins, a native failure can
leave the pasteboard cleared or partially populated. Native return values are
checked, but there is **no rollback, restoration, retry, or all-or-nothing
guarantee**. Concurrent writers follow normal pasteboard behavior: the last
writer wins. `pbjc` does not abort merely because the pasteboard changed since
the command started.

Representations are supplied eagerly and remain available after `pbjc` exits;
there is no background daemon or process that must stay alive to serve them.
Which representation a receiving application chooses is up to that application.

Successful copying is silent: no success message or copied data is printed to
stdout. Informational operations print their requested output. Errors are
reported on stderr.

| Exit status | Meaning |
| --- | --- |
| `0` | Successful copy or informational operation. |
| `1` | Runtime failure, including input, decoding, path preparation, or native publication failure. |
| `2` | Invalid command-line usage. |

## Build and runtime

See the [repository README](../README.md) for setup and build commands. Source
builds require macOS, Apple Command Line Tools, and the Go toolchain managed by
mise. The executable uses cgo and Objective-C/AppKit. The installed executable
requires macOS 13 or newer, but no Go installation, compiler, helper executable,
or daemon at runtime. The supported architecture targets are `darwin/arm64` and
`darwin/amd64`; this statement is not a claim that both targets have been
exercised. Copying `bin/pbjc` to a directory on `PATH` is sufficient to install
the built executable.
