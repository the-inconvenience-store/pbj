# pbjp verification

Live outcome: **VERIFIED** on macOS 26.5.1 / arm64, general clipboard only.
The [124-action report](../evidence/pbjp-2026-11-09/report.json) records each public
action, exact output check, preservation observation, and successful cleanup.

## Surface

The public surface is `bin/pbjp`, launched from `cmd/pbjp` and implemented by
`internal/command.Command.Paste`. `internal/pasteboard` owns native captures,
exact representation access, and locale encoding; the command decodes file URL
paths. Fixture publication uses the real `bin/pbjc`, not an internal Go setter. This is a CLI
extraction change: no TextEdit, Preview, Finder, screenshot, or GUI automation is
needed to prove its output.

## Launch

From the repository root on macOS 13+ with mise and Apple Command Line Tools:

```sh
mise trust
mise install
mise run bootstrap
mise run build
bin/pbjc --version
bin/pbjp --version
```

Readiness: both built binaries report versions and exit 0. The verifier also needs
Python 3, `/usr/bin/osascript`, and the macOS pasteboard service in the active user
session. There is no server, port, daemon, or long-running teardown handle. The
binaries do not require Python or osascript in normal use.

## Doctor

```sh
python3 scripts/verify-pbjp.py --doctor
```

Expected: JSON with `"doctor": "ready"`, macOS version, architecture, and the
`pbjc` and `pbjp` binary versions. Doctor invokes only each binary's `--version`
and `--help` plus the OS version query. It does not list clipboard types, fetch
payloads, publish fixtures, open applications, or change permissions. Doctor
checks launch/information readiness, not proof that extraction works.

After any surprising result, run Doctor again before driving. Do not substitute
unit tests for a failed public-path proof.

## Drive

Obtain explicit consent to replace only the general clipboard for the full proof.
Do not inspect, back up, or restore the old clipboard. Run:

```sh
python3 scripts/verify-pbjp.py --allow-clipboard --feature all \
  --evidence ".scratch/pbjp-proof-$(date +%Y%m%d-%H%M%S)"
```

The evidence directory must be new and outside `/tmp/pbjp-verify-*` scratch.
Select `--feature information|representations|files|failures` for a focused proof.
Only `information` is read-only and may omit `--allow-clipboard`:

```sh
python3 scripts/verify-pbjp.py --feature information \
  --evidence ".scratch/pbjp-information-$(date +%Y%m%d-%H%M%S)"
```

The driver publishes generated fixtures through `bin/pbjc` and consumes them
through `bin/pbjp`, waiting for each process to exit. A verification-owned native
osascript observer checks only known generated payload types and metadata. It
checks the seeded native change count before reading data and again afterward,
and checks native bytes/metadata before and after each extraction or failure.
A concurrent copy invalidates the proof instead of becoming report data.

For the genuinely zero-item case only, the driver explicitly calls native
`clearContents` after checking its own seeded fixture. This is fixture setup, not
an alternative extraction path; `pbjp` itself must produce empty default output,
`[]` metadata, and missing-data errors. Empty plain text is exercised separately.
Every subprocess has a 40-second timeout. No pre-existing process is stopped.

## Evidence

`report.json` contains host/architecture and binary names/versions; feature choice;
public actions and expected exits; expected payload lengths/SHA-256 hashes;
native per-item types; preservation assertions; and cleanup results. It records
no prior clipboard payloads and never prints unexpected extraction bytes or
native payloads. Repository, home, and scratch prefixes in exception messages are
redacted. Generated file URL/path bytes are compared in memory, not saved as
personal absolute paths in reports.

A result is `VERIFIED` only after every selected assertion and cleanup succeeds.
A failed drive yields `NOT VERIFIED`; readiness/setup exceptions yield
`INCONCLUSIVE`. Inspect the actual report, including `clipboard_reset`,
`clipboard_reset_boards`, `scratch_removed`, and errors. Do not mark unselected
features covered. No report or live outcome is fabricated by this artifact.

The find, ruler, and font pasteboards are **unexercised**: consent covers only
general. The driver neither reads nor writes their metadata or payloads and
records this coverage gap in `unexercised`.

The driver distinguishes text from preferred RTF/EPS using rich-plus-plain
fixtures. It also observes an advertised-but-unavailable derived text flavor:
metadata listing succeeds, exact extraction fails, and default selection falls
back to stored RTF. macOS may instead supply readable derived text on another
host; the fixture checks that against independently known text.

`pbjc` publishes identical bytes when given multiple raw type labels; the shared
EPS/RTF fixture proves byte preservation, not a ranking between distinct EPS
and RTF payloads. Source inspection supports that priority order. Generation-race
rejection has native test coverage; the serial live proof does not inject races.

## Cleanup

The driver owns one unique `/tmp/pbjp-verify-*` directory containing generated
files. Its `finally` cleanup writes `pbjp verification complete` through `bin/pbjc`
to the general pasteboard if touched, independently checks the generated
completion text, and removes only that scratch directory. General clipboard state
is not backed up. A read-only information run touches no clipboard. Reports remain
in the requested evidence directory, outside cleanup-owned scratch.

No application window or process is started beyond short-lived subprocesses, so
there is no GUI teardown or process shutdown. If cleanup fails, the report must
not say `VERIFIED`; inspect its named cleanup errors before retrying. Do not
recover a failed run by deleting evidence or by reading/restoring old contents.

## Isolation

Run this driver serially with all clipboard tests, pbjc drivers, clipboard
managers, and manual copying. The general clipboard is shared even though scratch
and evidence directories are unique. Disable external clipboard mutation for the
run rather than stopping pre-existing processes. Every fixture-writing group
touches only general; find/ruler/font selection remains unexercised because no
consent was granted to inspect or overwrite those boards. A native arm64
observation does not prove Intel, Rosetta, or the minimum macOS version; record
those hosts/builds separately.

## Feature map

[Every pbjp user-facing capability belongs here](features/README.md). Add an
independently reachable proof recipe for each new capability and keep the driver,
source entry points, and map aligned. The existing pbjc map and historical
evidence remain separate and unchanged.
