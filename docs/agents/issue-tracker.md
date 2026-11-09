# Issue tracker: GitHub

Issues and specs live in https://github.com/the-inconvenience-store/pbj.
Use the `gh` CLI from this repository.

## Operations

- Confirm the target before publishing: `gh repo view --json nameWithOwner`.
  It must be `the-inconvenience-store/pbj`.
- Create: `gh issue create --title "..." --body "..."`.
- Read: `gh issue view <number> --json number,title,body,labels,comments`.
- List: `gh issue list --state open --json number,title,body,labels`.
- Comment: `gh issue comment <number> --body "..."`.
- Label: `gh issue edit <number> --add-label "..." --remove-label "..."`.
- Close: post the explanation with `gh issue comment`, then `gh issue close`.
  Comment separately even when the issue is already closed.

When a skill says to publish a ticket, create a GitHub issue.
When it says to fetch a ticket, read the GitHub issue.

## Labels

Check `gh label list --limit 100 --json name` before labeled writes.
GitHub rejects issue creation when a requested label does not exist.
Use `docs/agents/triage-labels.md` for triage role mappings.

Wayfinder labels: `wayfinder:map`, `wayfinder:research`,
`wayfinder:prototype`, `wayfinder:grilling`, `wayfinder:task`.

## Relationships and wayfinding

Use native GitHub sub-issues for parent-child relationships and native
issue dependencies for blocking. Check `gh --version` and command help
before using version-dependent flags.

On gh 2.94+:
- Parent: `gh issue create --parent <parent> ...`.
- Blockers: `gh issue create --blocked-by <n>,<n> ...`.
- Add blocker: `gh issue edit <child> --add-blocked-by <blocker>`.

For older gh, use the REST API:
- Read database ID: `gh api repos/the-inconvenience-store/pbj/issues/<n> --jq .id`.
- Add child: `gh api --method POST repos/the-inconvenience-store/pbj/issues/<parent>/sub_issues -F sub_issue_id=<child-db-id>`.
- Add blocker: `gh api --method POST repos/the-inconvenience-store/pbj/issues/<child>/dependencies/blocked_by -F issue_id=<blocker-db-id>`.
- Count open blockers: `gh api repos/the-inconvenience-store/pbj/issues/<n> --jq .issue_dependencies_summary.blocked_by`.

Only when native relationships are unavailable, use a parent task list,
`Part of #<parent>` in child bodies, and `Blocked by: #<n>` for blockers.
A ticket is startable when all blockers are closed.

A wayfinder map is one issue labeled `wayfinder:map`, with
Notes / Decisions-so-far / Fog sections. Children use their type label.
Choose the first open, unassigned child in map order with no open blockers.
Claim with `gh issue edit <n> --add-assignee @me`.
Resolve by commenting, closing, and adding a gist plus link to the map's
Decisions-so-far section.

## Pull requests as a triage surface

**PRs as a request surface: no.**
