# pbjp feature map

Live outcome: **VERIFIED** for all five mapped capabilities on the general
clipboard, macOS 26.5.1 / arm64. The [124-action report](../../evidence/pbjp-2026-11-09/report.json)
and [notes](../../evidence/pbjp-2026-11-09/README.md) record the exercised paths and
limits. Use [surface instructions](../README.md) for Launch, Doctor, Evidence,
Cleanup, and Isolation.

| Capability | User outcome | Driver group |
| --- | --- | --- |
| [Information](information.md) | Discover generated help and version without reading the clipboard | `information` |
| [Representations and locale](representations.md) | Extract default/preferred text or rich data, or exact bytes with no conversion | `representations` |
| [Metadata, selection, and boards](metadata-and-items.md) | Inspect ordered item types, choose a zero-based item, and select a named system board | `representations` + `files` |
| [File paths](files.md) | Extract ordered file paths, safely including spaces and newline names | `files` |
| [Safe failures and preservation](failures.md) | Reject invalid/ambiguous requests with useful exits and preserve seeded data | `failures` + `files` + `representations` |

`--feature all` covers every group. Every fixture-writing group requires
`--allow-clipboard`; standalone `information` and Doctor do not. All successful
reads and failed extraction/usage cases after a seed independently recheck
native metadata, generation, and generated payload bytes.

Board-selection coverage is restricted to general (`-pboard general` and
`--pboard general --item 0`). The find/ruler/font paths are **unexercised** under
general-only consent; neither their metadata nor private payloads are read.
