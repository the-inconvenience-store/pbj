# Domain docs

## Layout and reading rules

This is a single-context repository.

Before exploring a domain area:
- Read root `GLOSSARY.md`, if present.
- Read relevant decisions under `docs/adr/`, if present.

If these artifacts are absent, proceed silently. Create them lazily through
domain modeling when terminology or decisions are actually resolved.

## Vocabulary

Use glossary terms in issues, proposals, code, and tests.
If a needed concept is missing, reconsider the terminology or note the gap
for domain modeling.

## Decisions

Surface conflicts with existing ADRs explicitly, identifying the decision
and why it may need reopening; do not silently override it.
