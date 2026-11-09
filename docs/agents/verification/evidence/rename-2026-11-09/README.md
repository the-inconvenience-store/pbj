# Executable rename verification — 2026-11-09

## Verdict

**VERIFIED:** the executable/file/documentation cutover to `pbjcopy` and
`pbjpaste`, including built commands, release archives, and source-built Homebrew
installation. Project/module/tap identity remains `pbj`. No compatibility aliases
were added. No remote push, tag, release, or tap publication occurred.

Implementation commits: `67335d3` and `e0883a4`. The real GoReleaser snapshot
records source commit `e0883a4e41aa32c041469d21da8d88062b1d900c`; later evidence
and status-documentation edits are not claimed inside that archive.

## Observed public paths

- [Source installation](report.json): the actual generated formula installed both
  renamed commands from the real snapshot archive. Homebrew loader/style,
  source compilation, unchanged formula tests, installed version/help/error
  smoke, and cleanup all passed. The [receipt](install-receipt.json) records
  neither a poured nor a built bottle. Only the source URL was adapted to an
  owned loopback server; compare [canonical](generated-pbj.rb) and
  [local transport](local-pbj.rb) formulas.
- [Actual archive execution](archives.json): checksums matched the generated
  manifest; each architecture archive had exactly the two renamed executable
  members. Source entry points are `cmd/pbjcopy/main.go` and
  `cmd/pbjpaste/main.go`. All four real binaries reported their new names and
  `0.0.0`, generated help with the new usage/documentation paths, and exited 2
  with the correct command prefix for an invalid flag. arm64 executed natively;
  x86_64 executed through Rosetta.
- [Copy information](pbjcopy-information/report.json): the renamed verifier
  passed help/native-help, version, type listing, and invalid invocation checks.
- [Paste information](pbjpaste-information/report.json): the renamed verifier
  passed every supported help/version spelling and checked generated options.
  Both information routes ran without clipboard consent or payload access.

The native built commands reported `dev`; versioned snapshot/Homebrew commands
reported `0.0.0`. This distinction is intentional. The snapshot is a local
fixture and must never be published.

## Reproduce

Follow the corresponding Launch and Doctor instructions, then use a new evidence
directory for each driver:

```sh
mise run check
mise run release-check
mise run release-snapshot
python3 scripts/verify-release.py --doctor
python3 scripts/verify-release.py --source dist/pbj_0.0.0_source.tar.gz \
  --allow-install --evidence NEW_DIRECTORY
python3 scripts/verify-pbjcopy.py --feature metadata --doctor
python3 scripts/verify-pbjcopy.py --feature metadata --evidence NEW_DIRECTORY
python3 scripts/verify-pbjpaste.py --doctor
python3 scripts/verify-pbjpaste.py --feature information --evidence NEW_DIRECTORY
```

The full automated gate passed build/typechecking, race tests, format checks,
lint with zero issues, module verification, and vulnerability audit with no
vulnerabilities found. Release-check passed GoReleaser validation, actionlint,
release tests, and Python process/cache-cleanup regressions. Each executable's
identity test was observed failing before its implementation change and passing
afterward. Independent Standards and Spec reviews against
`7df51ec29c17bfc40ef4fe43ad5837275147e4b8` both reported no findings.

## Cleanup and limits

The source driver removed its owned install, tap, trust configuration, server,
and scratch with no cleanup errors or new dependency versions. Existing prefix
command links were preserved. Information-driver scratch and temporary archive
extraction were removed. Old generated executable artifacts were removed rather
than left as aliases. The user's clipboard was untouched.

Old captures were retired rather than rewriting historical command output as a
new run. Clipboard data transfer and TextEdit/Preview/Finder paste were not rerun
for this name-only change; their feature cards explicitly retain that gap.
Native tests used isolated named pasteboards. macOS 13 runtime, native Intel
hardware, stable release versions, and authenticated GitHub/tap publication are
not claimed by these local proofs.
