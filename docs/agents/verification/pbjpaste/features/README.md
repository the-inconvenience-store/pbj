# pbjpaste feature map

**Information VERIFIED:** [fresh read-only proof](../../evidence/rename-2026-11-09/pbjpaste-information/report.json).
[Source installation/name proof](../../evidence/rename-2026-11-09/report.json) is also
VERIFIED. [Filesystem paste and all clipboard extraction/preservation groups](../../evidence/issue-1-filesystem-paste/report.json)
are **VERIFIED** on macOS/arm64 for the general pasteboard.
Use [surface instructions](../README.md) for Launch, Doctor, Evidence, Cleanup,
and Isolation.

| Capability | User outcome | Driver group |
| --- | --- | --- |
| [Information](information.md) | Discover generated help and version without reading the clipboard | `information` |
| [Representations and locale](representations.md) | Extract default/preferred text or rich data, or exact bytes with no conversion | `representations` |
| [Metadata, selection, and boards](metadata-and-items.md) | Inspect ordered item types, choose a zero-based item, and select a named system board | `representations` + `files` |
| [File paths](files.md) | Extract ordered file paths, safely including spaces and newline names | `files` |
| [Filesystem paste](filesystem.md) | Copy referenced files/directories without overwriting and keep the clipboard unchanged | `filesystem` |
| [Safe failures and preservation](failures.md) | Reject invalid/ambiguous requests with useful exits and preserve seeded data | `failures` + `files` + `representations` |

`--feature all` covers every group. Every fixture-writing group requires
`--allow-clipboard`; standalone `information` and Doctor do not. All successful
reads and failed extraction/usage cases after a seed independently recheck
native metadata, generation, and generated payload bytes.

Board-selection coverage is restricted to general (`-pboard general` and
`--pboard general --item 0`). The find/ruler/font paths are **unexercised** under
general-only consent; neither their metadata nor private payloads are read.
