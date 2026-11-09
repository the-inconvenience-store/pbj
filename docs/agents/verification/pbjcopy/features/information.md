# Command information

## User outcome

Discover copy options, version, and common pasteboard types without copying.
Ambiguous information/copy requests fail as usage errors.

## How to reach it

Follow the surface Launch and Doctor; no clipboard consent is needed for this
feature's informational and invalid invocations.

## How to drive it

```sh
python3 scripts/verify-pbjcopy.py --feature metadata \
  --evidence ".scratch/pbjcopy-metadata-$(date +%Y%m%d-%H%M%S)"
```

## Proof

The real binary's `--help` and native `-help` expose registered Cobra options;
version and type listing exit 0. Mixed help/copy options, an invalid board, and
mixed file/fallback modes exit 2 without stdout output.

**VERIFIED after rename:** [current metadata/usage run](../../evidence/rename-2026-11-09/pbjcopy-information/report.json).

## Gotchas

The type list is not an allowlist. Do not pin exact help wording or formatting.
The driver uses the built binary, not `go run`'s wrapper exit status.
