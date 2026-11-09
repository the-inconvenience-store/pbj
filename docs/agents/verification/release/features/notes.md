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

The parent exercises the real Cog CLI against the isolated privacy fixture and
then inspects notes at the intended existing release tag. For the first release:

```sh
mise exec -- cog changelog --at v0.1.0
```

Inspect `CHANGELOG.md` produced by the real bump too. For authorized publication,
follow [release operations](../../../../releases.md): the workflow renders these
exact-tag notes outside `dist/` and the worktree, then supplies GoReleaser's
`--release-notes`. Capture a sanitized comparison against the published release
body after publication. Do not disable GoReleaser's changelog pipe; that also
ignores supplied notes.

## Proof

The real Cog output contains intended change subjects and the correct version
range, but none of the fixture's author/co-author/signature markers. Earlier-tag
notes do not accidentally contain newer commits. The changelog and published
body follow the same author-free template. Keep only sanitized evidence: record
that synthetic privacy markers were absent, not real identities. Local tests or
snapshot notes do not establish the published body's contents.

## Gotchas

Status is **pending** until the parent records the real CLI result. A template
cannot sanitize a real name deliberately placed inside a commit subject. Review
subjects as well as generated notes. Never capture raw signed Git logs or tokens
in reports. A live GitHub release body is separately unavailable until an actual
release exists and is explicitly authorized; local notes are not that proof.
