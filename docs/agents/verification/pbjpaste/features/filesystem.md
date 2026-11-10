# Paste referenced filesystem objects

## User outcome

`pbjpaste --to DIR` copies selected local file references into an existing
folder, preserving supported native metadata and symlinks, without overwriting
existing names or modifying the clipboard. A batch stops at its first failure;
earlier completed objects remain.

## How to reach it

Build both binaries and pass the surface Doctor. Obtain permission to replace
only the general clipboard. The driver owns all source and destination fixtures
inside one unique `/tmp/pbjpaste-verify-*` directory; no GUI application is needed.
Run as a normal macOS user so mode-000 unreadable-directory fixtures are meaningful.
The metadata fixture uses macOS's `/usr/bin/xattr`.

Source entry points: `internal/command.Command.Paste`, snapshot dispatch in
`paste.go`, URL validation in `paste_files.go`, and staged native copying in
`paste_to_darwin.go`. Publication uses the real `pbjcopy --file` CLI.

## How to drive it

```sh
python3 scripts/verify-pbjpaste.py --allow-clipboard --feature filesystem \
  --evidence ".scratch/pbjpaste-filesystem-$(date +%Y%m%d-%H%M%S)"
```

## Proof

After the real paste process exits, verify exact binary and empty file contents,
recursive directory contents, names containing spaces/Unicode/newlines, symlink
identity/target text, executable permissions, modification time, and an extended
attribute. `--item` must copy only the selected directory. Every success/failure
checks native clipboard generation, per-item types, and generated file URL bytes.

Repeated paste and existing file/directory/symlink destinations must fail without
changing existing contents. Missing and nondirectory destinations, a missing
source after an earlier success, a destination inside its source through a
symlink, an unreadable subtree, and a FIFO must fail. Prior successful objects
remain after a later missing source; the failing tree is not published and
normal cleanup leaves no staging directory.
Destination and raw source URL paths containing `symlink/..` must select the
filesystem-resolved parent, not the lexical parent. Resolving source parents
must still preserve a final symlink rather than copy its target.

## Gotchas

This route writes only inside owned scratch and cleans it afterward. Never use
real user sources or destinations as collision/permission fixtures. Source
permissions are restored in `finally` before scratch cleanup. No source is frozen
while copying, no multi-item transaction is claimed, and a forced interruption
can leave owned staging; never sweep unrelated `.pbjpaste-*` directories.
Metadata proof is host/filesystem-specific and does not establish preservation
of ownership, set-id bits, hard-link relationships, or all ACL variants.
