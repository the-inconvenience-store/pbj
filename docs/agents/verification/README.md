# Verification

## Status

**INCONCLUSIVE — no product surface exists yet.** The approved scope is
repository setup only. Neither `pbjc` nor `pbjp` is implemented; there are
no Go packages, CLI flags, clipboard writes, or behavioral tests to drive.
Do not treat passing development checks as proof of clipboard behavior.

## Surface inventory

| Intended surface | Current state | Verification route |
| --- | --- | --- |
| `pbjc` copy CLI | Not implemented | No launch command or feature map exists yet |
| `pbjp` paste CLI | Explicitly deferred | No launch command or feature map exists yet |

No surface directories or executable feature recipes are defined because
there is no application interface to describe. Do not invent a command,
flag, build path, or supported pasteboard type from the reference project.

## Read-only repository doctor

From the repository root after `mise trust` and `mise install`:

```sh
mise exec -- go list ./...
mise run check
```

At initial setup, `go list` emits `matched no packages`. The check validates
the analyzer configuration and module checksums, while package operations
explicitly report `skipped: no Go packages`. That is the expected setup-only
state, not a verified application feature.

## Establish the first real route

After a real command is implemented, invoke verification with **Maintain**.
Inspect its actual entry point, supported behavior, macOS requirements,
and launch command. Add one surface directory with Launch, Doctor, Drive,
Evidence, Cleanup, Isolation, and a feature map; follow it end to end before
marking any feature VERIFIED or adding the verified Agent skills pointer.

The macOS pasteboard is shared mutable state. Future clipboard proofs must
run serially, obtain permission before replacing existing clipboard
content, avoid recording private clipboard data, and define preservation
and restoration before driving a copy operation. Do not claim that
`pbpaste` proves arbitrary rich pasteboard types.

## Setup evidence

See [bootstrap evidence](evidence/bootstrap.md) for exercised guardrails
and their boundaries. These checks are development-tool evidence, not
product-surface verification.
