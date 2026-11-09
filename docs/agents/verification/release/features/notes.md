# Author-free exact-tag notes

## User outcome

Release readers see useful Conventional Commit changes for their selected version
without author names, emails, usernames, co-author trailers, or Git signatures.
Cog owns both the changelog and the release notes; GoReleaser does not generate a
competing body.

## How to reach it

Use the intended committed source with full Git history/tags. The owned template
is `scripts/changelog.tera`, selected by `cog.toml`. Prepare a disposable local Git
fixture when proving a new version; never bump/tag the user's real checkout just
to test this capability. Synthetic fixture commit identities must be non-personal.

## How to drive it

Exercise the real Cog CLI against an isolated privacy fixture, then run the actual
release helper from that fixture to select the committed section for its tag:

```sh
mise exec -- go run ./scripts/release notes --tag v0.1.0
```

Inspect `CHANGELOG.md` produced by the real bump too. For authorized publication,
follow [release operations](../../../../releases.md): the workflow selects this
exact-tag section outside `dist/` and the worktree, then supplies GoReleaser's
`--release-notes`. Capture a sanitized comparison against the published release
body after publication. Do not disable GoReleaser's changelog pipe; that also
ignores supplied notes.

## Proof

The real Cog output contains intended change subjects, but none of the fixture's
author/co-author/signature markers. The helper selects the committed exact-tag
section; earlier-tag notes contain neither newer changes nor Cog's later version
bump commit. Publication uses that same author-free section. Record whether synthetic
privacy markers were absent, never real identities. Local tests or
snapshot notes do not establish the published body's contents.

## Gotchas

**Local VERIFIED:** [real Cog/helper proof](../../evidence/release-2026-11-09-verified/notes-proof.json).
Published body inspection has not run. A template cannot sanitize a real name
deliberately placed in a commit subject; review subjects too. Never capture raw
signed Git logs or tokens. Local notes are not proof of a published release body.
