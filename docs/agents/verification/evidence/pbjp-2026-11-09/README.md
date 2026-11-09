# pbjp live evidence — 2026-11-09

**Run verdict: VERIFIED. Verification maintenance outcome: changed.**

The [report](report.json) records 124 public actions on macOS 26.5.1 / arm64.
All five mapped capabilities received a live pass on the general clipboard.
Fixture publication used built `bin/pbjc`; extraction used built `bin/pbjp` in
separate processes. An independent AppKit observer compared known generated
payloads, advertised types, and native change count around reads and failures.

## Reproduce

From the repository root, with explicit consent to replace the general clipboard:

```sh
mise run build
python3 scripts/verify-pbjp.py --doctor
python3 scripts/verify-pbjp.py --allow-clipboard --feature all \
  --evidence .scratch/pbjp-new-proof
```

Use a new evidence directory. The recorded run used
`docs/agents/verification/evidence/pbjp-2026-11-09` instead of the scratch example.
No prior clipboard contents were inspected, saved, or restored.

## Observed results

- Cobra-generated help and version spellings succeeded without clipboard reads.
- Exact custom binary bytes preserved NUL, invalid UTF-8, CR/LF, and trailing NUL.
  Missing/conforming-but-unadvertised type requests failed without substitution.
- HTML and its separate plain fallback remained distinct; default text, RTF/PS
  preferences, native aliases, and missing-preference fallback selected the
  expected fixture bytes.
- Locale precedence LC_ALL → LC_CTYPE → LANG → C produced exact expected Latin-1
  or ASCII output. Exact extraction bypassed locale; unrepresentable text failed
  before stdout.
- Per-item JSON matched independent native metadata. Item filtering retained the
  original index. Ambiguous multi-item default/exact requests failed; explicit
  item selection returned only the chosen URL representation.
- File paths preserved order, spaces, embedded newlines in NUL mode, and the final
  record terminator. Ordinary newline mode rejected an unsafe second path before
  any stdout. A deleted target's reference remained extractable without stat or
  file-content reads.
- Empty data, a truly zero-item clipboard, unsupported default formats, missing
  data/indexes, repeated/conflicting flags, operands, and true-only boolean
  validation all followed the documented exit and output rules.
- Generated native bytes, types, and change count were unchanged after every
  seeded extraction, inspection, and expected failure.

## Promised representation edge

A native probe disproved a test fixture's assumption that publishing only RTF
means plain text is unadvertised. AppKit adds derived UTF-8/UTF-16 types. That
incidental test was removed rather than pinning platform-generated content.

The live report independently observed `advertised: true` with
`payload_available: false` for derived UTF-8 after the writer exited. Metadata
listing still succeeded. Exact UTF-8 extraction returned exit 1, zero stdout,
and preserved the fixture. Default extraction skipped that unavailable promise
and returned all 32 stored RTF bytes with exit 0. Exact RTF also matched.

**Boundary Discipline** drove the distinction: native unavailability is classified
separately, and only default preference selection may continue to the next type.
Native failures, generation changes, malformed text, and encoding loss do not
silently fall back; exact selection never does.

## Supporting checks

- Final format check, golangci-lint (0 issues), build of both commands, complete
  race-enabled Go test suite, and module verification passed.
- Govulncheck reported 0 reachable vulnerabilities; 1 uncalled advisory in
  imported packages and 13 in required modules/toolchain remain reported.
- One committed-diff review from `331b484` found 0 Standards and 0 Spec findings.
  It was source-only; the live pass closed its unavailable-flavor evidence gap.
- The amd64 executable built, reported its version, and native pasteboard tests
  passed under Rosetta on this Apple Silicon host. This is not physical Intel
  testing or a full Intel CLI drive.
- Final arm64 Mach-O metadata declares macOS 13.0 minimum / SDK 26.5 and links
  only system libraries/frameworks. That is not a run on macOS 13.
- Source and current executable private-identity scans passed. Builds use
  `-trimpath`; new commits use the configured pseudonymous identity.

## Cleanup and limits

The report confirms `clipboard_reset: true`, `scratch_removed: true`, and no
cleanup errors. Its unique generated fixture directory was removed; evidence
remains here. The clipboard contains `pbjp verification complete`. No application
windows, persistent product processes, or runtime helpers were created.

Find/ruler/font boards were neither read nor written: consent covered general
only. macOS 13 and physical Intel hardware were not exercised. The serial live
pass does not inject generation races (native tests cover replacement rejection),
and identical multi-type fixture bytes do not independently prove rich-only
EPS-versus-RTF priority. That priority is explicit in source. No GUI receiving-app
paste is needed to prove this CLI extraction surface. Fresh-checkout installation
and hosted CI were not exercised; CI remains outside scope.
