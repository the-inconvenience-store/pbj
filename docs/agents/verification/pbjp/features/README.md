# pbjp feature map

All live outcomes are **pending** until the named driver group has run and its
report has been inspected. Use [surface instructions](../README.md) for shared
Launch, Doctor, Evidence, Cleanup, and Isolation rules.

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
