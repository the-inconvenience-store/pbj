# Receiving applications

Status: **UNEXERCISED AFTER RENAME**. This is the proof recipe, not a new observed
result; the pending source-install/name proof does not exercise application paste.

## User outcome

Paste ordinary text, rendered HTML, and an image into real native applications.
Finder file pasting has its own mapped feature.

## How to reach it

Follow Launch and Doctor in the graphical desktop. Authorize clipboard changes,
generated documents/windows, Accessibility actions, and own-window capture.

## How to drive it

```sh
python3 scripts/verify-pbjcopy.py --allow-clipboard --feature apps \
  --evidence ".scratch/pbjcopy-apps-$(date +%Y%m%d-%H%M%S)"
```

## Proof

TextEdit Cmd+V must save exact Unicode plain text. In a rich-text document it must
render the HTML fixture as bold text, save RTF, and expose a bold font attribute.
Preview New from Clipboard must render the generated 256×256 four-color PNG.
Inspect each generated-window screenshot, then confirm all test windows closed.

## Gotchas

Open TextEdit fixtures through Launch Services, not a bare AppleScript open event,
so its sandbox can access them. Resolve WindowServer capture IDs independently;
application scripting IDs are not necessarily capture IDs. Do not overwrite or
close pre-existing user documents. Receiving apps choose their own representation;
this recipe proves named outcomes, not compatibility with every application.
