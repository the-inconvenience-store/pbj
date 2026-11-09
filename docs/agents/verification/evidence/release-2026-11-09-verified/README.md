# Release verification — 2026-11-09

## Verdict and scope

**VERIFIED: local source-built Homebrew installation, Cog/helper note selection,
and GoReleaser snapshot archives.** No production tag, GitHub release, or tap
formula was published. Authenticated publication remains unrun pending explicit
first-release approval; configured secret metadata is not authorization proof.

The snapshot records source commit `1c93d9f2ed5bd714f8c07a164111b25dd7e10ff0`.
Later evidence/documentation edits are not claimed inside that archive. The
`0.0.0` version is a local fixture, never an asset/formula to publish.

## Observed behavior

- [Source-install report](report.json): the actual release helper generated a
  formula using the real GoReleaser source checksum. Homebrew loaded and styled
  the canonical formula, then built the unchanged recipe with only its download
  URL redirected to an owned loopback server. `brew test --force` tested the
  unlinked installation; both installed commands passed version/help/error smoke.
- [Durable receipt](install-receipt.json): the keg belonged to the unique local
  tap, with both `poured_from_bottle` and `built_as_bottle` false. The
  [canonical](generated-pbj.rb) and [local-transport](local-pbj.rb) formulas
  preserve identical build/test/version/checksum logic.
- [Actual Cog/helper CLI proof](notes-proof.json): a disposable Git fixture
  produced `v0.1.0` and `v0.1.1`; the helper selected the committed `v0.1.0`
  section unchanged after the later bump. Synthetic author/coauthor markers were
  absent. Regenerating with `cog changelog --at` would instead include the later
  version-bump commit; publication therefore selects committed notes.
- [Tag validation](tag-validation.json): stable `v0.1.0` succeeded; a prerelease
  and a missing `v` prefix failed through the actual release CLI.
- [Archive proof](assets-proof.json): all three real archive checksums matched
  GoReleaser's manifest. Both binary archives contained both commands. All four
  executables reported `0.0.0` and displayed help: arm64 natively, x86_64 through
  Rosetta. No Developer ID authority was present; arm64 had Go's ad-hoc signature
  and x86_64 was unsigned.

## Commands

```sh
mise run check
mise run release-check
mise run release-snapshot
python3 scripts/verify-release.py --doctor
python3 scripts/verify-release.py \
  --source dist/pbj_0.0.0_source.tar.gz \
  --allow-install \
  --evidence docs/agents/verification/evidence/release-2026-11-09-verified
```

Choose a new evidence directory for another run; this directory is preserved
proof, not disposable scratch. Archive extraction/execution and Cog/helper fixture
operations are recorded in their separate JSON artifacts.

The complete Go gate passed: build/typechecking, race suite, formatting, lint
with zero issues, module verification, and `govulncheck` with no vulnerabilities
found. Release-check passed GoReleaser schema validation, actionlint, real Cog
regressions, and the Python child-process/read-only-cache cleanup regressions.
Go was pinned to 1.27.3 after the HTTPS publisher made standard-library
vulnerabilities reachable on 1.27.0; the audit was not suppressed.

## Cleanup and earlier attempts

The final source run removed its owned keg, tap, scratch/cache/log/temp and
scratch-only trust configuration with no cleanup errors. Existing command links
were unchanged and no clipboard payload was read or written. Cog Git fixtures,
compiled helper, archive extraction directories and inspection caches were
removed. Shared cache permissions were not changed.

Earlier attempts are preserved in the neighboring `release-2026-11-09`,
`release-2026-11-09-source`, and `release-2026-11-09-install` directories. They
exposed canonical tar-root parsing, explicit Homebrew tap trust, unlinked testing,
and read-only module-cache cleanup. Unrelated installed-formula inventories are
redacted; dependency differences remain recorded. The failed install attempt's
owned scratch residue was subsequently removed with the bounded cleanup helper.
Its original report describes cleanup at that attempt's end, not the later
manual cleanup.

Homebrew added Go 1.27.3 alongside existing 1.27.0 during that earlier source
build; both versions were retained. Homebrew also populated its own Ruby/bundler
tooling. The old report's claim of no dependency upgrades was incorrect: Homebrew
can update build dependencies despite automatic-update/install-upgrade controls.
The driver and instructions now explicitly report and retain those changes.
The final run added no further formula dependency versions.

The loopback HTTP server logged a client disconnect during the successful run.
Checksum, installation, formula tests, command observations and cleanup all
passed; no server exception was suppressed to obtain the verdict.

## Review and remaining gaps

The committed implementation was reviewed against fixed point
`6efc03cdc2bd03ad9f60806e5e42cc5113470d7a`: Standards had no findings; Spec found
one interruption cleanup defect. A real child/grandchild regression failed
before the fix and passed after process-group termination/reaping was added.
Subsequent real-surface failures were fixed and re-exercised, not papered over.

`HOMEBREW_TAP_TOKEN` presence was confirmed through GitHub secret metadata without
retrieving its value. The public tap is reachable. Still unexercised: production
credential authorization, tag-triggered hosted workflow, published release body
and downloads, actual tap create/update/no-op, and recovery dispatch. No Intel
hardware or macOS 13 runtime claim is made. These gaps require an authorized
stable release and its durable remote results, not another passing local test.
