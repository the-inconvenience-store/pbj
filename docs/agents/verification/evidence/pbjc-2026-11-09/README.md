# pbjc live evidence — 2026-11-09

**Run verdict: VERIFIED. Verification maintenance outcome: changed.**

The [report](report.json) records 35 observed actions through the compiled arm64
CLI on macOS 26.5.1. Every mapped feature received a live pass. The driver waited
for each pbjc process to exit before independent AppKit payload observation.

## Public-path proof

- Unicode text, literal markup as plain text, HTML with a separate fallback,
  empty fallback, explicit custom binary types, deduplicated types, empty input,
  and locale decoding matched exact fixture bytes.
- PNG, JPEG, both TIFF byte orders, PDF, RTF, and EPS signature routing matched
  the advertised native types. Only PNG was a complete generated image; the
  minimal other headers prove recognition, not document validity.
- Real stdin read failure, malformed text, fallback collisions, and a missing
  file preserved the previously generated sentinel.
- File mode ignored an unreadable stdin descriptor, removed a duplicate path,
  and preserved both ordered file URLs after writer exit. Actual Finder Edit →
  Paste created both destination files with exact source contents.
- TextEdit Cmd+V saved exact plain text. Rich paste rendered `pbjc bold` with the
  observed `Times-Bold` font and saved RTF.
- Preview New from Clipboard rendered the generated 256×256 four-color image.

Visual evidence was inspected, not inferred from filenames:
[text](textedit-plain.png), [rendered HTML](textedit-rich.png),
[image](preview-image.png). Captures contain only generated fixture windows.

## Handoff defect caught by the live pass

Initial in-process/unique-board checks did not expose a general-board lifetime
problem: the writer-side cache held two file items, but immediate process exit
could leave only one visible externally. A diagnostic wait while the writer was
alive preserved both; waiting after exit did not recover the missing item.

Commit `5378aca` adds an independent native reader-reference synchronization and
item-count round trip after the checked AppKit write. It does not add a sleep,
retry, daemon, payload readback copy, ownership assertion, or rollback. The result
count is deliberately not asserted because another application may replace the
clipboard under the accepted last-writer-wins contract.

Immediate post-exit checks then preserved both references, and the original
Finder scenario passed. The consent-gated `files` driver retains this exact
public-path regression check; ordinary tests do not modify the general clipboard.
A separate subprocess test covers native unique-board ownership after writer exit.

## Automated and architecture evidence

- Complete `mise run check` passed: format check, lint, build, race-enabled test
  suite, module verification, and govulncheck with 0 reachable vulnerabilities.
- After the handoff correction, affected native and command race tests passed;
  the subprocess regression, lint, build, and commit hooks also passed.
- Govulncheck additionally reported 1 uncalled advisory in imported packages and
  13 uncalled advisories in required modules/toolchain; these were not claimed
  remediated.
- Both arm64 and amd64 executables built. Intel native publication and process-
  exit tests ran under Rosetta, not on physical Intel hardware.
- The [amd64 live file proof](../pbjc-2026-11-09-amd64/report.json) additionally
  drives the Intel CLI and real Finder paste on the same Apple Silicon host.

## Cleanup and gaps

The report confirms `scratch_removed: true` and `clipboard_reset: true`.
Generated app documents/windows were closed; shared GUI applications were not
quit. The clipboard was left with `pbjc verification complete`, without recording
or restoring its prior contents. No runtime helper or product process remains.

macOS 13 itself, physical Intel hardware, other receiving applications, and
writes to find/font/ruler pasteboards were not exercised by the live pass.
Architecture declarations and the minimum deployment target are not proof of
those environments. Fresh-checkout installation and hosted CI remain unverified;
CI was explicitly excluded from scope. `pbjp` remains deferred.
