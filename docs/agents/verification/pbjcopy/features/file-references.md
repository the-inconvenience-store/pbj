# File references

Status: actual Finder paste is **UNEXERCISED AFTER RENAME**. Source-install/name
proof does not exercise this clipboard or receiving-application outcome.

## User outcome

Copy multiple selected file references and paste both actual files into Finder,
without reading their contents or requiring the writer to stay alive.

## How to reach it

Follow Launch and Doctor; authorize the general clipboard and a generated Finder
window. The driver creates its own sources and an empty destination under /tmp.

## How to drive it

```sh
python3 scripts/verify-pbjcopy.py --allow-clipboard --feature files \
  --evidence ".scratch/pbjcopy-files-$(date +%Y%m%d-%H%M%S)"
```

## Proof

Invoke the real CLI with two files plus a repeated first file and an unreadable
stdin descriptor. It must succeed without reading stdin. Independent native
readback immediately after process exit must contain both ordered, deduplicated
file URLs. Finder Edit → Paste must create both files with exact source contents.

## Gotchas

A live writer or same-process cached readback can hide a handoff defect. Keep the
post-exit read. Use a fresh empty destination to avoid replacement dialogs.
The automated unique-board tests additionally cover directories, symlinks,
lexical normalization, home expansion, and broken-reference preservation.
