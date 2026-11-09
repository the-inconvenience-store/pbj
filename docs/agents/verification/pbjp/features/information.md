# Discover help and version

## User outcome

Users can inspect Cobra-generated paste options with `--help`, `-h`, or native
`-help`, and version with `--version` or `-v`. These standalone commands do not
need clipboard data and may not be combined with extraction options or operands.

## How to reach it

Build both binaries and pass [Doctor](../README.md#doctor). No fixture, clipboard
consent, or application window is needed.

Source entry points: `cmd/pbjp`, `internal/command.Command.Paste`, and Cobra option
registration/validation in `internal/command`. The verification-owned recipe is
`Proof.information` and `doctor` in `scripts/verify-pbjp.py`.

## How to drive it

```sh
python3 scripts/verify-pbjp.py --feature information \
  --evidence ".scratch/pbjp-information-$(date +%Y%m%d-%H%M%S)"
```

## Proof

Every spelling exits 0 with nonempty stdout and empty stderr. Help exposes type,
list-types, item, files, null, pboard, and prefer options. Doctor reports both
binary versions and the host OS/architecture. This group performs no native
clipboard calls and publishes no fixture. Its report records hashes of expected
information output and `clipboard_untouched: true`.

## Gotchas

Information options are standalone, not a way to suppress an invalid invocation.
Mixed information/extraction options, repeated flags, and operands are driven by
the [failure group](failures.md), which uses an authorized generated fixture to
prove preservation. A passing Doctor is not extraction evidence.
