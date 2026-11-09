# Bootstrap evidence — 2026-11-09

Scope: repository setup only; no application implementation or CI.

## Exercised commands and observations

- `gh repo view --json nameWithOwner,url,hasIssuesEnabled`: resolved
  `the-inconvenience-store/pbj`, with Issues enabled.
- Created the four missing triage labels and five wayfinder labels with
  successful `gh label create` calls. Existing `wontfix` was preserved.
- `test -L CLAUDE.md` and `test CLAUDE.md -ef AGENTS.md`: passed.
- `mise install`: all pinned tools installed.
- `mise run bootstrap`: module download completed with no dependencies;
  Lefthook installed pre-commit and commit-msg hooks.
- `mise run check` and `mise run fmt` on the final repository: passed;
  package checks explicitly skipped because no Go packages exist.

## Disposable Go fixture

A temporary package under `guardrail_smoke/` was created and removed.
It was not committed and is not a product implementation.

- Unformatted source made `mise run fmt-check` exit 1 and print the diff.
- `mise run fmt` applied formatting; the next format check passed.
- Comparing an error with `io.EOF` using `==` made `mise run lint` exit 1
  with an `errorlint` diagnostic about wrapped errors.
- Replacing that comparison with `errors.Is` made lint pass with 0 issues.
- `mise run check` compiled the fixture, ran race-enabled test discovery
  (`[no test files]`), verified module checksums, and ran govulncheck.
- Govulncheck reported 0 reachable vulnerabilities and 14 advisories in
  required modules/toolchain whose vulnerable symbols were not called.
  The pinned Go toolchain was not upgraded or claimed free of advisories.

## Hook proof

Lefthook's `run` command was invoked with controlled file inputs; no real
commits, index staging, or user identity changes were performed.

- `pre-commit --file guardrail_smoke/smoke.go`: rejected the lint violation;
  passed after correction. Source bytes were unchanged by the passing hook.
- `pre-commit --file README.md`: skipped because no Go/config files matched.
- Passing Go pre-commit run: about 1.5 seconds; documentation-only no-op:
  about 0.07 seconds. These are fixture measurements, not future budgets.
- `commit-msg` accepted `chore: configure development guardrails` and
  rejected `not a conventional commit`, using a message file whose path
  contained spaces.
- Cocogitto initially panicked because Git `user.name` was absent. The
  successful policy proof used a temporary HOME containing a test Git
  identity; it did not set the user's identity. Configure real Git identity
  before committing.

## Remaining gaps

- No user-facing command exists: product verdict **INCONCLUSIVE**.
- No permanent behavioral tests exist; race-enabled test discovery is not
  evidence of behavior or concurrency safety.
- Fresh-checkout tool installation was not proved in a clone or CI.
- Actual partially staged commits and real Git hook dispatch were not
  exercised; checks are read-only and direct Lefthook paths were exercised.
- Dependabot configuration was written; hosted update execution was not
  observed. Tool-version updates remain manual.
- Module/toolchain advisories remain as described above; no product audit
  can run until code exists.
- No CI by explicit choice. No verified-runtime Agent skills pointer is
  added until a real mapped feature is proved.

Temporary source and message files were removed. No product process was
started, and the clipboard was not accessed or modified.
