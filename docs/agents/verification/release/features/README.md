# Release feature map

Live outcome: **pending parent orchestration** for every capability below. No
GitHub publication or Homebrew installation is claimed by these instructions.
Use the [surface](../README.md) for consent, launch, Doctor, evidence, and cleanup.

| Capability | User outcome | Proof route |
| --- | --- | --- |
| [Author-free notes](notes.md) | Read exact-tag release notes without contributor identities or Git signatures | Cog real CLI + exact published body inspection |
| [Version ownership](versions.md) | Get a deliberate stable Cog version/tag and matching installed command versions | Isolated Cog bump + helper validation + source driver |
| [Release assets](assets.md) | Download versioned source and optional two-command macOS archives | Actual GoReleaser snapshot/archive inspection; published download separately |
| [Source-built Homebrew installation](source-install.md) | Build/install/test both native commands from checksummed source without signing/quarantine workarounds | `scripts/verify-release.py` actual install |
| [Publication and recovery boundaries](publication.md) | Publish the formula separately and retry it without recreating release assets | GitHub workflow/API operational evidence, explicitly authorized |

The source driver is intentionally not an all-features pass: its report lists the
remaining notes, bump, remote publication, architecture, and OS coverage gaps.
The public tap repository exists; secret availability and its authorization are
not inferred from repository existence. Remote release/formula publication remains
unproved until real authorized actions and durable results are captured.

The parent orchestrator owns the complete checks, snapshot creation from intended
committed source, actual installation, evidence inspection, and final verdicts.
No live commands, builds, or tests were run while authoring this surface.
