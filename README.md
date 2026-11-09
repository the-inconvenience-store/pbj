# pbj

Peanut butter and jelly for the macOS clipboard: enriched `pbcopy` and
`pbpaste` commands written in Go, inspired by
[pbrich](https://github.com/waynehoover/pbrich).

`pbjc` copies text, rich bytes, or file references to a macOS pasteboard.
`pbjp` reads representations or file-reference paths from a pasteboard.

## Build and install

Source builds require macOS and Apple Command Line Tools (`xcode-select --install`).
Install [mise](https://mise.jdx.dev/), which manages the pinned Go toolchain, then:

```sh
mise trust
mise install
mise run bootstrap
mise run build
printf 'Hello, clipboard!\n' | bin/pbjc
bin/pbjp
```

The build produces `bin/pbjc` and `bin/pbjp`. Copy those executables to a
directory on `PATH` to install them. Runtime requires macOS 13 or newer, with
no Go installation, compiler, helper executable, or daemon. Supported targets are `darwin/arm64`
and `darwin/amd64`; this is not a claim that both have been exercised.

```sh
bin/pbjc < image.png
printf '<strong>Hello</strong>' | bin/pbjc --type public.html --plain 'Hello'
bin/pbjc --file ./report.pdf --file ./images
bin/pbjc --help
bin/pbjp --list-types
bin/pbjp --type public.html > clipboard.html
bin/pbjp --files --null
bin/pbjp --help
```

Copying is silent on success; pasting writes the requested data to stdout.
See the [copy command reference](docs/pbjc.md) and
[paste command reference](docs/pbjp.md) for all flags, locale handling, file
references, and failure semantics. The copy examples replace the general
pasteboard's contents; paste operations leave it unchanged.

## Development

Bootstrap downloads module dependencies and installs Lefthook's local Git
hooks. Commit hooks check Go formatting and lint without changing or staging
files. Commit messages must follow Conventional Commits, for example
`feat: add HTML clipboard copying`.

```sh
mise run check       # Complete local checks, including module verification
mise run fmt         # Apply formatting and import cleanup
mise run fmt-check   # Read-only formatting check
mise run lint        # Validate lint config and analyze Go packages
mise run build       # Build bin/pbjc and bin/pbjp
mise run test        # Race-enabled tests
mise run audit       # Vulnerability scan; requires network access
```

Dependencies are pinned by `go.mod` and, when needed, `go.sum`; checks use
read-only module resolution. Update them deliberately with Go module
commands and commit both files when changed. Dependabot checks Go module
updates weekly. Tool versions are pinned in `mise.toml` and updated manually.

There is no CI workflow yet. Run the complete check locally before pushing.
Agent configuration lives in [AGENTS.md](AGENTS.md).

Verification records live in
[the verification index](docs/agents/verification/README.md); documentation
and supported-target declarations are not evidence of exercised behavior.
