# pbj

Peanut butter and jelly for the macOS clipboard: enriched `pbcopy` and
`pbpaste` commands written in Go.

`pbjcopy` copies text, rich bytes, or file references to a macOS pasteboard.
`pbjpaste` reads representations or file-reference paths from a pasteboard.

## Install with Homebrew

Install the latest stable release:

```sh
brew install the-inconvenience-store/tap/pbj
pbjcopy --version
pbjpaste --version
```

If Homebrew refuses to load an untrusted tap, review its `Formula/pbj.rb`, run
`brew trust --tap the-inconvenience-store/tap`, and retry the install.

The formula builds both commands from the release's source archive on macOS
13 or newer. It requires Apple Command Line Tools (`xcode-select --install`);
Homebrew supplies Go as a build dependency. It is not a cask or a prebuilt bottle
and does not require Developer ID signing or notarization.

## Build and install

Source builds require macOS and Apple Command Line Tools (`xcode-select --install`).
Install [mise](https://mise.jdx.dev/), which manages the pinned Go toolchain, then:

```sh
mise trust
mise install
mise run bootstrap
mise run build
printf 'Hello, clipboard!\n' | pbjcopy
pbjpaste
```

The build produces `pbjcopy` and `pbjpaste`. Runtime requires macOS 13 or newer,
with no Go installation, compiler, helper executable, or daemon. Supported targets
are `darwin/arm64` and `darwin/amd64`; this is not a claim that both have been
exercised.

```sh
pbjcopy < image.png
printf '<strong>Hello</strong>' | pbjcopy --type public.html --plain 'Hello'
pbjcopy --file ./report.pdf --file ./images
pbjcopy --help
pbjpaste --list-types
pbjpaste --type public.html > clipboard.html
pbjpaste --files --null
pbjpaste --help
```

Copying is silent on success; pasting writes the requested data to stdout.
See the [copy command reference](docs/pbjcopy.md) and
[paste command reference](docs/pbjpaste.md) for all flags, locale handling, file
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
mise run build       # Build pbjcopy and pbjpaste
mise run test        # Race-enabled tests
mise run audit       # Vulnerability scan; requires network access
```

Dependencies are pinned by `go.mod` and, when needed, `go.sum`; checks use
read-only module resolution. Update them deliberately with Go module
commands and commit both files when changed. Dependabot checks Go module
updates weekly. Tool versions are pinned in `mise.toml` and updated manually.

Release-only GitHub Actions automation runs the complete checks for pushed stable
version tags, publishes GitHub release assets, then updates the Homebrew formula.
Branch pushes and pull requests do not run CI; run the complete check locally
before pushing. See [the release guide](docs/releases.md) for preparation, version
bumps, local snapshots, and tap-publication retries.
Agent configuration lives in [AGENTS.md](AGENTS.md).

Verification records live in
[the verification index](docs/agents/verification/README.md); documentation
and supported-target declarations are not evidence of exercised behavior.
