# pbjcopy feature map

**Command information VERIFIED:** [fresh information proof](../../evidence/rename-2026-11-09/pbjcopy-information/report.json).
[Source installation/name proof](../../evidence/rename-2026-11-09/report.json) is also
VERIFIED. [Stdin attachments and Mail paste](../../evidence/issue-2-attachments/report.json),
[representations](../../evidence/issue-2-representations/report.json), and
[safe failures](../../evidence/issue-2-safe-failures/report.json) are **VERIFIED**.
TextEdit, Preview, and Finder application-paste outcomes remain **UNEXERCISED AFTER RENAME**.

| Feature | User outcome | Driver feature |
| --- | --- | --- |
| [Command information](information.md) | Discover options, version, common types, and usage errors | metadata |
| [Stdin representations](stdin-representations.md) | Copy text, detected rich formats, explicit raw types, and fallbacks | representations |
| [Failure preservation](safe-failures.md) | Keep existing copied content when preparation fails | failures |
| [File references and attachments](file-references.md) | Paste selected files in Finder, or materialized stdin as a Mail attachment after pbjcopy exits | files, attachments |
| [Receiving applications](application-paste.md) | Paste ordinary text, rendered HTML, and an image into native apps | apps |

Follow the surface Launch and Doctor before each live pass. The `all` driver
covers every row in one serial clipboard session. Record a verdict per exercised
path and report unsupported/unexercised platforms separately.
