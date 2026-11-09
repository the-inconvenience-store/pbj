# Publication and recovery boundaries

## User outcome

A deliberate stable tag publishes a GitHub release first, then a separately
checksummed source formula. If only the tap job fails, the operator can retry it
without reuploading release assets. Identical formula publication is a no-op;
changed formula publication uses the current file SHA and a non-personal bot
identity.

## How to reach it

Follow [release operations](../../../../releases.md). The public
`the-inconvenience-store/homebrew-tap` exists. Its `v0.1.0` default-branch write
through `HOMEBREW_TAP_TOKEN` was verified without retrieving the credential's value.
The release job uses only its scoped automatic GitHub token; only the tap job
receives the separate tap credential. Human provisioning/authorization is outside
the local install proof and outside this driver's permissions.

## How to drive it

The parent first inspects the real helper's stateful HTTP-boundary tests and the
workflow configuration. Those support the contract but cannot prove production
credentials or GitHub publication. Before any authorized live run, establish token
availability through the approved secret mechanism without exposing its value.

For explicitly authorized publication, an operator intentionally pushes the
Cog-created stable tag as described in release operations. Observe the tag job's
full-history checkout, complete checks, exact-tag Cog notes, GoReleaser release,
and subsequent separate Homebrew publication job. Inspect the actual GitHub
release body/asset list, download its source asset, recalculate SHA256, and read
`Formula/pbj.rb` from the tap's default branch. The formula must refer to that
release's source asset, not a binary archive or a locally served snapshot.

If the release succeeded but tap publication failed, use **Re-run failed jobs**.
An alternative, **only with explicit authorization**, is the documented tap-only
workflow dispatch:

```sh
gh workflow run release.yml --repo the-inconvenience-store/pbj --ref main -f tag=v0.1.0
```

Inspect the completed job and durable formula state. Repeat identical publication
only when authorized and observe no additional tap commit. Do not trigger a live
failure deliberately, change secrets, or republish an old version as a test.

## Proof

The release exists as stable, notes match its committed exact-tag Cog section, and
the named source asset is downloadable/checksummed. The tap formula's version,
source URL, and SHA256 match that actual asset. Source installation is proved
separately; the workflow does not publish bottles, casks, signed binaries, or
notarization. Tap retry/dispatch does not execute GoReleaser or modify previously
uploaded assets. Same-content publication leaves the tap commit unchanged;
changed-content publication updates against the current Contents API SHA.
Capture public object IDs, checksums, sanitized job conclusions, and absence of
extra commits, not credential/authentication headers or personal commit metadata.

## Gotchas

**v0.1.0 VERIFIED:** [successful hosted release and tap jobs, downloaded assets,
and durable formula state](../../evidence/v0.1.0-publication/report.json).
The formula matches the actual source checksum and uses a bot author/committer.
The credential's value was never retrieved. Recovery dispatch and identical
publication/no-op have not been exercised; first publication does not prove them.
Never move/recreate the published tag. A failed release job needs release-state
inspection, not a tap-only retry. Older-tag retries reject a newer tap version.
