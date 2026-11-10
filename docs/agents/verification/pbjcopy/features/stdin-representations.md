# Stdin representations

## User outcome

Copy text and enriched content, choosing explicit representations and an optional
plain-text fallback without implicit format conversion.

## How to reach it

Follow Launch and Doctor; obtain permission to replace the general clipboard.

## How to drive it

```sh
python3 scripts/verify-pbjcopy.py --allow-clipboard --feature representations \
  --evidence ".scratch/pbjcopy-representations-$(date +%Y%m%d-%H%M%S)"
```

## Proof

After each real pbjcopy process exits, independent AppKit readback must match exact
fixture bytes: Unicode text, literal HTML fragments as ordinary text, HTML plus
explicit fallback, empty fallback, custom binary types, PNG/JPEG/TIFF/PDF/RTF/EPS
signature recognition, empty input, and locale-decoded Latin-1 text. Repeated
identifiers must not add representations; explicit types bypass detection and decoding.

Structured fixtures must advertise only `public.json`, `public.html`, or
`public.xml` unless `--plain` explicitly adds `public.utf8-plain-text`.
Check complete JSON objects/arrays, declared XML, HTML documents, XHTML's HTML
precedence, exact raw payload extraction through `bin/pbjpaste --type`, explicit
overrides, and fallback text. Scalars, fragments, incomplete documents, malformed
JSON/XML, and multiple XML roots must remain plain text. A fallback conflict
after malformed structured input must preserve the generated sentinel clipboard.

## Gotchas

Only PNG is a complete generated image in this recipe; the other minimal rich
fixtures prove signature routing, not document validity or app interoperability.
`pbpaste` alone does not prove rich types. Locale changes affect only child
invocations, not the shell or other applications.
