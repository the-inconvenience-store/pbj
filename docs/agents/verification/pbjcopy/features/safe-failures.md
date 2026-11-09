# Failure preservation

## User outcome

A failed read, invalid text, conflicting fallback, or missing file leaves the
previous clipboard content intact when publication has not begun.

## How to reach it

Follow Launch and Doctor; authorize clipboard replacement. The driver seeds a
known generated sentinel instead of reading or saving private clipboard data.

## How to drive it

```sh
python3 scripts/verify-pbjcopy.py --allow-clipboard --feature failures \
  --evidence ".scratch/pbjcopy-failures-$(date +%Y%m%d-%H%M%S)"
```

## Proof

Malformed UTF-8, both explicit/detected plain-text collisions, an actual stdin
read error from a directory descriptor, and a missing file must fail. The same
sentinel must remain readable after every failed invocation.

## Gotchas

This is a pre-publication guarantee. Do not promise atomic replacement or rollback
on a native failure after publication starts. There is no automatic retry.
