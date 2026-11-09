# Stdin representations

## User outcome

Copy text and enriched content, choosing explicit representations and an optional
plain-text fallback without implicit format conversion.

## How to reach it

Follow Launch and Doctor; obtain permission to replace the general clipboard.

## How to drive it

```sh
python3 scripts/verify-pbjc.py --allow-clipboard --feature representations \
  --evidence ".scratch/pbjc-representations-$(date +%Y%m%d-%H%M%S)"
```

## Proof

After each real pbjc process exits, independent AppKit readback must match exact
fixture bytes: Unicode text, literal HTML as ordinary text, HTML plus explicit
fallback, empty fallback, custom binary types, PNG/JPEG/TIFF/PDF/RTF/EPS signature
recognition, empty input, and locale-decoded Latin-1 text. Repeated identifiers
must not add representations; explicit types bypass detection and decoding.

## Gotchas

Only PNG is a complete generated image in this recipe; the other minimal rich
fixtures prove signature routing, not document validity or app interoperability.
`pbpaste` alone does not prove rich types. Locale changes affect only child
invocations, not the shell or other applications.
