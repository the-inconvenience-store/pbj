# File references and stdin attachments

Status: actual Finder paste is **UNEXERCISED AFTER RENAME**. Source-install/name
proof does not exercise this clipboard or receiving-application outcome.
[Stdin attachment CLI and Mail paste](../../evidence/issue-2-attachments/report.json)
are **VERIFIED**; the body-only screenshot was inspected.

## User outcome

Copy multiple selected file references and paste both actual files into Finder,
without reading their contents or requiring the writer to stay alive.
Alternatively, materialize arbitrary stdin as a named attachment and paste it
into a generated, unsent Mail compose window.

## How to reach it

Follow Launch and Doctor; authorize the general clipboard and a generated Finder
window. The driver creates its own sources and an empty destination under /tmp.

## How to drive it

```sh
python3 scripts/verify-pbjcopy.py --allow-clipboard --feature files \
  --evidence ".scratch/pbjcopy-files-$(date +%Y%m%d-%H%M%S)"
```

For stdin attachments, obtain separate Mail compose-window consent, then run:

```sh
python3 scripts/verify-pbjcopy.py --allow-clipboard --allow-mail --feature attachments \
  --evidence ".scratch/pbjcopy-attachments-$(date +%Y%m%d-%H%M%S)"
```

## Proof

Invoke the real CLI with two files plus a repeated first file and an unreadable
stdin descriptor. It must succeed without reading stdin. Independent native
readback immediately after process exit must contain both ordered, deduplicated
file URLs. Finder Edit → Paste must create both files with exact source contents.

Attachment proof runs the compiled `--attachment NAME` mode with unknown binary,
PNG, empty input, repeated filenames, and a flag-like filename. After each
process exits, `pbjpaste --files --null` must extract the exact native file
reference; backing-file bytes, filename, and private file/directory permissions
must match. Read/cache failures preserve the prior clipboard and leak no new
backing directories. Clipboard replacement must not remove earlier attachments.

Mail body Cmd+V must expose the named attachment as an accessibility button.
Inspect `mail-attachment.png`, a body-only crop of the exact generated Mail
window, for a file attachment rather than inline payload or filename text.
No message is sent; the generated compose window is discarded.

## Gotchas

A live writer or same-process cached readback can hide a handoff defect. Keep the
post-exit read. Use a fresh empty destination to avoid replacement dialogs.
The automated unique-board tests additionally cover directories, symlinks,
lexical normalization, home expansion, and broken-reference preservation.

Attachment runs override HOME only in CLI child processes, keeping all backing
files in the driver's owned scratch cache. Cleanup removes that scratch after
Mail observation; this is verification cleanup, not product auto-expiry. Never
delete a real user's older attachment caches. Keep unrelated Mail windows and
drafts untouched, and do not interact with app focus during the proof.
