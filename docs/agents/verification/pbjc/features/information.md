# Command information

## User outcome

Discover copy options, version, and common pasteboard types without copying.
Ambiguous information/copy requests fail as usage errors.

## How to reach it

Follow the surface Launch and Doctor; no clipboard consent is needed for this
feature's informational and invalid invocations.

## How to drive it

```sh
python3 scripts/verify-pbjc.py --feature metadata \
  --evidence ".scratch/pbjc-metadata-$(date +%Y%m%d-%H%M%S)"
```

## Proof

The real binary's `--help` and native `-help` expose registered Cobra options;
version and type listing exit 0. Mixed help/copy options, an invalid board, and
mixed file/fallback modes exit 2 without stdout output.

## Gotchas

The type list is not an allowlist. Do not pin exact help wording or formatting.
The driver uses the built binary, not `go run`'s wrapper exit status.
