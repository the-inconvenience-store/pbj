# pbjcopy verification

Rename status: **NOT RUN** for the renamed executable. Fresh source-install/name
proof is reserved at `../evidence/rename-2026-11-09/report.json`, pending the
parent's actual run. TextEdit, Preview, and Finder paste outcomes are
**UNEXERCISED AFTER RENAME**; no historical GUI result is claimed as fresh.

## Surface

`cmd/pbjcopy` launches the public Cobra CLI. `internal/command` owns invocation and
input preparation; `internal/pasteboard` owns native publication and ownership.
The receiving applications are TextEdit, Preview, and Finder.

## Launch

From the repository root, after installing mise and Apple Command Line Tools:

```sh
mise trust
mise install
mise run bootstrap
mise run build
pbjcopy --version
```

Readiness: build succeeds and version exits 0. Each invocation is short-lived;
there is no port, server, daemon, authentication, or long-running teardown handle.
The installed executable needs macOS 13+, not Go, Swift, or a runtime helper.

## Doctor

```sh
python3 scripts/verify-pbjcopy.py --doctor
```

Expected: JSON with `"doctor": "ready"`, the binary version, architecture, and
macOS version. Doctor is read-only: it does not copy or read clipboard payloads.
Full application proof requires a logged-in graphical macOS session, an English
Finder menu bar, Python 3, osascript, Swift from the developer tools, Accessibility
permission, and Screen Recording permission for the driving terminal/harness.
Swift is used only to check capture permission, not by the product executable.

After any surprising result, run Doctor again before driving. Permission changes
are human-owned; do not silently bypass a failed check or substitute unit tests.

## Drive

Obtain permission to replace the general clipboard and open generated windows.
Then run:

```sh
python3 scripts/verify-pbjcopy.py --allow-clipboard --feature all \
  --evidence ".scratch/pbjcopy-proof-$(date +%Y%m%d-%H%M%S)"
```

Each evidence directory must be new. Select one mapped feature with
`--feature metadata|representations|failures|files|apps` when a focused proof is
appropriate. Only `metadata` can run without `--allow-clipboard`.

The driver invokes the real built executable, waits for it to exit, and observes
payloads independently through `scripts/inspect-pasteboard.js` and AppKit. It
never calls internal Go functions or synthesizes expected clipboard state.

Application drivers are `scripts/paste-into-textedit.applescript`,
`scripts/paste-into-preview.applescript`, and
`scripts/paste-into-finder.applescript`. TextEdit fixtures open through Launch
Services to grant sandbox access. The driver pastes into generated documents;
Preview uses New from Clipboard; Finder uses its enabled Edit → Paste action.
`scripts/capture-verification-window.js` resolves actual WindowServer IDs rather
than confusing them with application scripting IDs.

## Evidence

`report.json` records public actions, exit statuses, verified payload lengths and
hashes, native type/readback results, saved application results, and cleanup.
Successful app proof also captures `textedit-plain.png`, `textedit-rich.png`, and
`preview-image.png`. Inspect these images; file existence alone is not visual
proof. HTML must render bold, not merely appear as HTML source or fallback text.

Finder proof requires both actual destination files and exact generated contents.
The file proof checks ordered references immediately after the writer exits;
metadata-only or same-process readback cannot replace this lifetime check.

Do not record prior clipboard data, unexpected bytes, personal paths, or unrelated
windows. Failure reports redact repository and scratch prefixes.

## Cleanup

The driver closes its own app documents/windows, removes its unique
`/tmp/pbjcopy-verify-*` directory, and leaves a harmless clipboard completion message.
GUI applications are shared and are not quit. The JSON report and screenshots
remain in the requested evidence directory. Confirm `scratch_removed` and
`clipboard_reset` in the report; failed cleanup is not a complete pass.

If an application driver fails, its error handler closes its known generated
window/document. Inspect any remaining fixture window before retrying; never
close an unrelated window or discard a user's unsaved document.

## Isolation

Only one live driver may run at a time: the general clipboard and app focus are
shared. Unit/integration tests instead use unique named native pasteboards and
release them globally on cleanup. Do not run GUI automation alongside live
clipboard proof or interact with its generated windows during the run.

The native arm64 path and selected Intel/Rosetta checks are separate observations.
A minimum deployment target does not prove behavior on macOS 13 itself.

## Feature map

[Every user-facing capability belongs in the feature map](features/README.md).
Add a feature recipe when a new independently provable outcome is introduced.
