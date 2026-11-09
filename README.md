# pbj

Peanut butter and jelly for the macOS clipboard: enriched `pbcopy` and
`pbpaste` commands written in Go, inspired by
[pbrich](https://github.com/waynehoover/pbrich).

Planned commands: `pbjc` for copy, then `pbjp` for paste.
This repository currently contains development setup only; neither command
is implemented.

## Development

Install [mise](https://mise.jdx.dev/), then:

```sh
mise trust
mise install
mise run bootstrap
mise run check
```

Bootstrap downloads module dependencies and installs Lefthook's local Git
hooks. Commit hooks check Go formatting and lint without changing or staging
files. Commit messages must follow Conventional Commits, for example
`feat: add HTML clipboard copying`.

```sh
mise run fmt         # Apply formatting and import cleanup
mise run fmt-check   # Read-only formatting check
mise run lint        # Validate lint config and analyze Go packages
mise run build       # Compile packages
mise run test        # Race-enabled tests
mise run audit       # Vulnerability scan; requires network access
```

Until Go packages exist, package checks explicitly report a skip. This is
not evidence of working clipboard behavior or behavioral test coverage.
Dependencies are pinned by `go.mod` and, when needed, `go.sum`; checks use
read-only module resolution. Update them deliberately with Go module
commands and commit both files when changed. Dependabot checks Go module
updates weekly. Tool versions are pinned in `mise.toml` and updated manually.

There is no CI workflow yet. Run the complete check locally before pushing.
Agent configuration lives in [AGENTS.md](AGENTS.md).

Verification status and the missing runtime prerequisite are recorded in
[the verification index](docs/agents/verification/README.md).
