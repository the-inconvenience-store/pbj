#!/usr/bin/env python3
"""Drive built pbjpaste against generated pbjcopy fixtures; never save the old clipboard."""

import argparse
import base64
import hashlib
import json
import os
from pathlib import Path
import platform
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parent.parent
PLAIN = "public.utf8-plain-text"
HTML = "public.html"
RTF = "public.rtf"
EPS = "com.adobe.encapsulated-postscript"
RAW = "com.example.pbjpaste-verification.raw"
FILE = "public.file-url"
ENV = {**os.environ, "LC_ALL": "en_US.UTF-8"}
FEATURES = ("information", "representations", "files", "filesystem", "failures")

# Metadata is safe before a fixture write. Payload observation is enabled only
# for explicitly named generated representations and the seeded change count.
INSPECT = r'''
ObjC.import("AppKit");
ObjC.import("Foundation");
function run(args) {
    const names = {general: $.NSPasteboardNameGeneral, find: $.NSPasteboardNameFind,
                   ruler: $.NSPasteboardNameRuler, font: $.NSPasteboardNameFont};
    const board = $.NSPasteboard.pasteboardWithName(names[args[0]]);
    const count = Number(board.changeCount);
    const expected = JSON.parse(args[1]);
    if (expected !== null && count !== expected) throw Error("fixture generation changed");
    const items = board.pasteboardItems;
    const types = [];
    for (let i = 0; i < Number(items.count); i++) {
        const item = items.objectAtIndex(i);
        const row = [];
        for (let j = 0; j < Number(item.types.count); j++) {
            row.push(ObjC.unwrap(item.types.objectAtIndex(j)));
        }
        types.push(row);
    }
    const requested = JSON.parse(args[2]);
    const payloads = {};
    for (const type of requested) {
        payloads[type] = [];
        for (let i = 0; i < Number(items.count); i++) {
            const data = items.objectAtIndex(i).dataForType($(type));
            payloads[type].push(data.isNil() ? null :
                ObjC.unwrap(data.base64EncodedStringWithOptions(0)));
        }
    }
    if (Number(board.changeCount) !== count) throw Error("fixture generation changed");
    return JSON.stringify({count: count, types: types, payloads: payloads});
}
'''


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def invoke(arguments, *, data=b"", env=None):
    return subprocess.run(list(map(str, arguments)), input=data, capture_output=True,
                          env=env or ENV, cwd=ROOT, timeout=40)


def digest(data):
    return {"bytes": len(data), "sha256": hashlib.sha256(data).hexdigest()}


def redact(value, scratch=None):
    text = str(value)
    for prefix, replacement in ((str(scratch) if scratch else "", "$SCRATCH"),
                                (str(ROOT), "$REPO"), (str(Path.home()), "$HOME")):
        if prefix:
            text = text.replace(prefix, replacement)
    return text


def inspect(board, count=None, types=()):
    require(board == "general", "live verification consent covers only the general pasteboard")
    result = invoke(["/usr/bin/osascript", "-l", "JavaScript", "-e", INSPECT,
                     board, json.dumps(count), json.dumps(list(types))])
    require(result.returncode == 0, "native generated-fixture observation failed")
    return json.loads(result.stdout)


def doctor():
    require(platform.system() == "Darwin", "macOS is required")
    versions = {}
    for name in ("pbjcopy", "pbjpaste"):
        for flag in ("--version", "--help"):
            result = invoke([ROOT / "bin" / name, flag])
            require(result.returncode == 0 and not result.stderr and bool(result.stdout),
                    "build bin/" + name + "; its read-only " + flag + " must succeed")
            if flag == "--version":
                versions[name] = redact(result.stdout.decode().strip())
    system = invoke(["/usr/bin/sw_vers", "-productVersion"])
    require(system.returncode == 0, "macOS version query failed")
    return {"system": "macOS", "macOS": system.stdout.decode().strip(),
            "architecture": platform.machine(), "binaries": versions}


class Proof:
    def __init__(self, scratch):
        self.scratch = scratch
        self.actions = []
        self.touched = set()
        self.board = "general"
        self.state = None
        self.expected = {}

    def record(self, action, **observed):
        self.actions.append({"action": action, "observed": observed})

    def seed(self, data=b"", arguments=(), expected=None, board="general"):
        require(board == "general", "live verification consent covers only the general pasteboard")
        self.touched.add(board)
        result = invoke([ROOT / "bin/pbjcopy", "--pboard", board, *arguments], data=data)
        require(result.returncode == 0 and not result.stdout and not result.stderr,
                "pbjcopy generated-fixture seed failed")
        self.board = board
        self.state = inspect(board)
        self.expected = expected if expected is not None else {PLAIN: [data]}
        for type_name, values in self.expected.items():
            require(len(self.state["types"]) == len(values), "seeded item count differs")
            require(all(value is None or type_name in self.state["types"][index]
                        for index, value in enumerate(values)), "seeded type is missing")
        self.preserved()
        self.record("pbjcopy seed", board=board, items=len(self.state["types"]),
                    payloads={key: [None if value is None else digest(value) for value in values]
                              for key, values in self.expected.items()})

    def preserved(self):
        require(self.state is not None, "no generated fixture has been seeded")
        observed = inspect(self.board, self.state["count"], self.expected)
        require(observed["types"] == self.state["types"], "fixture metadata changed")
        for key, values in self.expected.items():
            encoded = [None if value is None else base64.b64encode(value).decode()
                       for value in values]
            require(observed["payloads"][key] == encoded, "generated fixture bytes changed")

    def paste(self, arguments=(), *, expected=b"", status=0, env=None, label=None,
              contains=None):
        self.preserved()
        result = invoke([ROOT / "bin/pbjpaste", *arguments], env=env)
        require(result.returncode == status, "unexpected pbjpaste exit status for " +
                (label or " ".join(map(str, arguments)) or "default extraction"))
        require(result.stdout == expected, "unexpected pbjpaste stdout for " +
                (label or " ".join(map(str, arguments)) or "default extraction"))
        if status == 0:
            require(not result.stderr, "successful pbjpaste emitted stderr")
        else:
            require(bool(result.stderr), "failed pbjpaste omitted diagnostic")
        if contains is not None:
            require(contains in result.stderr, "pbjpaste diagnostic lacks required guidance")
        self.preserved()
        self.record(label or ("pbjpaste " + " ".join(map(str, arguments))),
                    exit=status, stdout=digest(expected), preserved=True,
                    environment={key: env[key] for key in ("LC_ALL", "LC_CTYPE", "LANG")
                                 if env is not None and key in env})
        return result

    def listing(self, arguments=()):
        self.preserved()
        result = invoke([ROOT / "bin/pbjpaste", "--list-types", *arguments])
        require(result.returncode == 0 and not result.stderr, "type listing failed")
        rows = [{"item": index, "types": types}
                for index, types in enumerate(self.state["types"])]
        if "--item" in arguments:
            index = int(arguments[arguments.index("--item") + 1])
            rows = [rows[index]]
        require(json.loads(result.stdout) == rows, "type listing differs from native metadata")
        self.preserved()
        self.record("pbjpaste --list-types " + " ".join(map(str, arguments)),
                    items=rows, preserved=True)

    def empty_board(self):
        # Clear only our known generated state, never inspect the previous user data.
        self.preserved()
        script = r'''
ObjC.import("AppKit");
function run(args) {
    const board = $.NSPasteboard.generalPasteboard;
    if (Number(board.changeCount) !== Number(args[0])) throw Error("fixture generation changed");
    board.clearContents;
    return "cleared";
}
'''
        result = invoke(["/usr/bin/osascript", "-l", "JavaScript", "-e", script,
                         str(self.state["count"])])
        require(result.returncode == 0, "generated empty-board fixture setup failed")
        self.state = inspect("general")
        self.expected = {}
        require(self.state["types"] == [], "empty-board fixture still has items")
        self.record("native clearContents of generated fixture", items=0)
        self.listing()
        self.paste(label="empty clipboard default is empty success")
        self.paste(["--type", PLAIN], status=1)
        self.paste(["--files"], status=1)
        self.paste(["--item", "0"], status=1)

    def information(self):
        # This feature remains read-only and can run without clipboard consent.
        for flag in ("--help", "-h", "-help", "--version", "-v"):
            result = invoke([ROOT / "bin/pbjpaste", flag])
            require(result.returncode == 0 and not result.stderr and bool(result.stdout),
                    "pbjpaste information failed")
            if "help" in flag or flag == "-h":
                for option in (b"--type", b"--list-types", b"--item", b"--files",
                               b"--null", b"--pboard", b"--prefer", b"--to"):
                    require(option in result.stdout, "generated help omitted a paste option")
            self.record("pbjpaste " + flag, exit=0, stdout=digest(result.stdout), clipboard_read=False)

    def representations(self):
        raw = b"\x00\xff\x7fraw\r\n\x00"
        self.seed(raw, ["-t", RAW], {RAW: [raw]})
        self.listing()
        self.paste(["-t", RAW], expected=raw)
        self.paste(["--type", "public.data"], status=1, label="exact type has no conformance fallback")
        self.paste(expected=b"", label="unsupported default format is empty success")
        html, plain = b"<b>pbjpaste generated bold</b>", b"pbjpaste generated bold"
        self.seed(html, ["-t", HTML, "--plain", plain.decode()], {HTML: [html], PLAIN: [plain]})
        self.listing()
        self.paste(expected=plain)
        self.paste(["--type", HTML], expected=html)
        self.paste(["--type", PLAIN], expected=plain)
        self.paste(["--type", "public.text"], status=1,
                   label="exact generic text type does not use advertised HTML or UTF8 text")
        self.paste(["--prefer", "ps"], expected=plain, label="missing preferred PS falls back to text")
        rtf = b"{\\rtf1\\ansi pbjpaste generated rich}"
        eps = b"%!PS-Adobe-3.0 EPSF-3.0\n%%Title: pbjpaste generated\n"
        self.seed(rtf, ["-t", RTF, "--plain", "pbjpaste rich fallback"],
                  {RTF: [rtf], PLAIN: [b"pbjpaste rich fallback"]})
        self.paste(expected=b"pbjpaste rich fallback")
        self.paste(["--prefer", "rtf"], expected=rtf)
        self.paste(["--prefer", "ps"], expected=b"pbjpaste rich fallback")
        self.seed(eps, ["-t", EPS, "--plain", "pbjpaste EPS fallback"],
                  {EPS: [eps], PLAIN: [b"pbjpaste EPS fallback"]})
        self.paste(["-Prefer", "ps"], expected=eps)
        self.paste(["--prefer", "txt"], expected=b"pbjpaste EPS fallback")
        self.paste(["--prefer", "ascii"], expected=b"pbjpaste EPS fallback")
        self.seed(rtf, ["-t", RTF], {RTF: [rtf]})
        derived = inspect(self.board, self.state["count"], [PLAIN])["payloads"][PLAIN][0]
        plain = b"pbjpaste generated rich"
        advertised = PLAIN in self.state["types"][0]
        if derived is not None:
            require(derived == base64.b64encode(plain).decode(),
                    "generated RTF-derived text differs from the known fixture")
        self.record("native RTF-derived flavor availability",
                    advertised=advertised, payload_available=derived is not None)
        self.listing()  # Must succeed even when an advertised payload is unavailable.
        if advertised and derived is not None:
            self.paste(expected=plain, label="default uses readable advertised derived text")
        else:
            self.paste(["--type", PLAIN], status=1,
                       label="exact unavailable or unadvertised derived text never falls back")
            self.paste(expected=rtf,
                       label="default skips unavailable derived text and reads stored RTF")
        self.paste(["--type", RTF], expected=rtf)
        self.seed(eps, ["-t", EPS], {EPS: [eps]})
        self.paste(expected=eps)
        self.paste(["--prefer", "rtf"], expected=eps, label="missing RTF preference falls back to EPS")
        self.seed(eps, ["-t", EPS, "-t", RTF], {EPS: [eps], RTF: [eps]})
        self.paste(expected=eps, label="default preserves shared EPS and RTF representation bytes")
        self.paste(["--prefer", "rtf"], expected=eps, label="RTF preference preserves shared rich bytes")
        text = "pbjpaste café".encode()
        self.seed(text)
        self.paste(expected=text)
        latin = {**ENV, "LC_ALL": "fr_FR.ISO8859-1", "LC_CTYPE": "C", "LANG": "C"}
        self.paste(expected=b"pbjpaste caf\xe9", env=latin, label="LC_ALL overrides LC_CTYPE and LANG")
        ctype = {**ENV, "LC_ALL": "", "LC_CTYPE": "fr_FR.ISO8859-1", "LANG": "C"}
        self.paste(expected=b"pbjpaste caf\xe9", env=ctype, label="LC_CTYPE overrides LANG")
        language = {**ENV, "LC_ALL": "", "LC_CTYPE": "", "LANG": "fr_FR.ISO8859-1"}
        self.paste(expected=b"pbjpaste caf\xe9", env=language, label="LANG locale fallback")
        self.paste(["--type", PLAIN], expected=text, env=latin, label="exact bytes bypass locale")
        self.paste(env={**ENV, "LC_ALL": "C"}, status=1, label="unrepresentable locale text fails without partial output")
        self.seed(b"pbjpaste ASCII")
        no_locale = {key: value for key, value in ENV.items() if key not in ("LC_ALL", "LC_CTYPE", "LANG")}
        self.paste(expected=b"pbjpaste ASCII", env=no_locale, label="unset locale falls back to C")
        self.seed(b"")
        self.paste()
        self.paste(["--type", PLAIN])
        self.listing()
        self.empty_board()
        value = b"pbjpaste generated general"
        self.seed(value)
        self.paste(["-pboard", "general"], expected=value)
        self.paste(["--pboard", "general", "--item", "0"], expected=value)
        self.listing(["--pboard", "general", "--item", "0"])

    def files(self):
        first = self.scratch / "generated file A.txt"
        second = self.scratch / "generated-file-B.txt"
        newline = self.scratch / "generated\nfile-C.txt"
        for index, path in enumerate((first, second, newline)):
            path.write_bytes(("pbjpaste generated file " + str(index)).encode())
        urls = [path.as_uri().encode() for path in (first, second)]
        self.seed(arguments=["--file", first, "--file", second], expected={FILE: urls})
        self.listing()
        self.listing(["--item", "1"])
        self.paste(["--files"], expected=(str(first) + "\n" + str(second) + "\n").encode(),
                   label="ordered multi-item file paths with spaces")
        self.paste(["--files", "--null"], expected=(str(first) + "\0" + str(second) + "\0").encode(),
                   label="ordered file paths with trailing NUL")
        self.paste(["--files", "--item", "1"], expected=(str(second) + "\n").encode(),
                   label="file subset by item")
        self.paste(["--files", "--null", "--item", "0"], expected=(str(first) + "\0").encode(),
                   label="NUL file subset by item")
        self.paste(status=1, label="ambiguous default multi-item extraction rejected")
        self.paste(["--type", FILE], status=1, label="ambiguous exact multi-item extraction rejected")
        self.paste(["--type", FILE, "--item", "1"], expected=urls[1],
                   label="exact selected file URL bytes")
        self.paste(["--item", "1"], label="selected default without a supported format is empty success")
        self.paste(["--item", "2"], status=1)
        self.paste(["--list-types", "--item", "2"], status=1)
        self.paste(["--files", "--item", "2"], status=1)
        self.seed(arguments=["--file", first, "--file", newline],
                  expected={FILE: [first.as_uri().encode(), newline.as_uri().encode()]})
        self.paste(["--files"], status=1, contains=b"--null", label="newline path rejected before any stdout")
        self.paste(["--files", "--null"], expected=(str(first) + "\0" + str(newline) + "\0").encode(),
                   label="newline path preserved by NUL separation")
        self.paste(["--files", "--item", "0"], expected=(str(first) + "\n").encode(),
                   label="selected safe path ignores other newline path")
        newline.unlink()
        self.paste(["--files", "--null"], expected=(str(first) + "\0" + str(newline) + "\0").encode(),
                   label="file reference extraction does not require target existence")


    def filesystem(self):
        source = self.scratch / "filesystem-sources"
        source.mkdir()
        binary = source / "space ☃\nfile.bin"
        binary.write_bytes(bytes(range(256)))
        empty = source / "empty"
        empty.write_bytes(b"")
        tree = source / "tree"
        (tree / "nested").mkdir(parents=True)
        payload = tree / "nested/payload"
        payload.write_bytes(b"owned directory payload")
        payload.chmod(0o751)
        os.utime(payload, (1700000000, 1700000000))
        metadata = invoke(["/usr/bin/xattr", "-w", "com.example.pbjpaste-verification", "owned metadata", payload])
        require(metadata.returncode == 0, "could not prepare owned extended-attribute fixture")
        (tree / "dangling").symlink_to("missing-target")
        link = source / "tree-link"
        link.symlink_to("tree")

        def seed(paths):
            arguments = [argument for path in paths for argument in ("--file", path)]
            urls = [(path.as_uri() + ("/" if path.is_dir() else "")).encode() for path in paths]
            self.seed(arguments=arguments, expected={FILE: urls})

        destination = self.scratch / "filesystem-destination"
        destination.mkdir()
        seed([binary, empty, tree, link])
        self.paste(["--to", destination], label="filesystem paste of multiple files, directory, and symlink")
        require((destination / binary.name).read_bytes() == bytes(range(256)), "binary file copy differs")
        require((destination / empty.name).read_bytes() == b"", "empty file copy differs")
        copied = destination / "tree/nested/payload"
        require(copied.read_bytes() == b"owned directory payload", "recursive file content differs")
        require(copied.stat().st_mode & 0o777 == 0o751 and copied.stat().st_mtime_ns == 1700000000000000000,
                "native file mode or timestamp was not preserved")
        metadata = invoke(["/usr/bin/xattr", "-p", "com.example.pbjpaste-verification", copied])
        require(metadata.returncode == 0 and metadata.stdout.rstrip(b"\n") == b"owned metadata",
                "native extended attribute was not preserved")
        require(os.readlink(destination / "tree/dangling") == "missing-target" and
                os.readlink(destination / link.name) == "tree", "symlinks were dereferenced or changed")
        require(set(path.name for path in destination.iterdir()) == {binary.name, empty.name, "tree", link.name},
                "filesystem destination contains extra or missing objects")
        self.record("inspect copied filesystem objects",
                    binary=digest(bytes(range(256))), recursive=digest(b"owned directory payload"),
                    empty=digest(b""), names_with_spaces_unicode_newlines=True,
                    symlinks_preserved=True, mode_mtime_xattr_preserved=True)
        self.paste(["--to", destination], status=1, label="repeat paste refuses existing names")
        require((destination / binary.name).read_bytes() == bytes(range(256)), "repeat paste changed data")
        selected = self.scratch / "filesystem-selected"
        selected.mkdir()
        self.paste(["--to", selected, "--item", "2"], label="filesystem item selection")
        require(set(path.name for path in selected.iterdir()) == {"tree"} and
                (selected / "tree/nested/payload").read_bytes() == b"owned directory payload",
                "filesystem selection copied the wrong items")

        seed([binary])
        for kind in ("file", "directory", "symlink"):
            collision = self.scratch / ("collision-" + kind)
            collision.mkdir()
            target = collision / binary.name
            sentinel = self.scratch / ("keep-" + kind)
            sentinel.write_bytes(b"keep existing bytes")
            if kind == "file":
                target.write_bytes(b"keep existing bytes")
            elif kind == "directory":
                target.mkdir()
                (target / "keep").write_bytes(b"keep existing bytes")
            else:
                target.symlink_to(sentinel)
            self.paste(["--to", collision], status=1, label="refuse existing " + kind + " destination")
            require(set(path.name for path in collision.iterdir()) == {binary.name}, "collision leaked staging")
            retained = target / "keep" if kind == "directory" else target
            require(retained.read_bytes() == b"keep existing bytes", "collision changed existing bytes")
            if kind == "symlink":
                require(target.is_symlink() and os.readlink(target) == str(sentinel), "collision replaced symlink")

        self.paste(["--to", self.scratch / "missing-destination"], status=1, label="reject missing destination")
        self.paste(["--to", binary], status=1, label="reject nondirectory destination")
        missing = source / "removed-after-copy"
        missing.write_bytes(b"will be removed")
        seed([binary, missing, empty])
        missing.unlink()
        partial = self.scratch / "partial-destination"
        partial.mkdir()
        self.paste(["--to", partial], status=1, label="missing source stops batch after earlier successful copy")
        require(set(path.name for path in partial.iterdir()) == {binary.name} and
                (partial / binary.name).read_bytes() == bytes(range(256)), "partial-success contract differs")

        seed([tree])
        alias = self.scratch / "recursive-alias"
        alias.symlink_to(tree / "nested")
        before = set((tree / "nested").iterdir())
        self.paste(["--to", alias], status=1, label="reject destination inside source through symlink")
        require(set((tree / "nested").iterdir()) == before, "recursive rejection changed source tree")
        blocked = tree / "unreadable"
        blocked.mkdir()
        blocked.chmod(0)
        failed = self.scratch / "failed-tree-destination"
        failed.mkdir()
        try:
            self.paste(["--to", failed], status=1, label="unreadable subtree prevents top-level publication")
        finally:
            blocked.chmod(0o700)
        require(not list(failed.iterdir()), "unreadable tree was published or staging leaked")
        special = source / "special"
        special.mkdir()
        os.mkfifo(special / "pipe", 0o600)
        seed([special])
        self.paste(["--to", failed], status=1, label="unsupported FIFO fails without publication")
        require(not list(failed.iterdir()), "special-file copy published data or staging leaked")
        self.record("filesystem failure safety", no_overwrite=True, partial_success_documented=True,
                    failed_object_not_published=True, clipboard_generation_and_bytes_preserved=True)

    def failures(self):
        self.seed(b"pbjpaste generated preservation sentinel")
        self.paste(["--type", RAW], status=1)
        self.paste(["--files"], status=1)
        invalid = [
            ["--type", ""], ["--type"], ["--item", "-1"], ["--item", "no"],
            ["--pboard", "unknown"], ["--prefer", "html"], ["--null"],
            ["operand"], ["--", "operand"], ["--unknown"],
            ["--type", PLAIN, "--files"], ["--type", PLAIN, "--prefer", "txt"],
            ["--files", "--prefer", "txt"], ["--list-types", "--type", PLAIN],
            ["--list-types", "--files"], ["--list-types", "--prefer", "txt"],
            ["--list-types", "--null"], ["--help", "--version"],
            ["--help", "--item", "0"], ["--version", "--pboard", "general"],
            ["--help", "operand"], ["--version", "--list-types"],
            ["--to"], ["--to", ""], ["--to", ".", "--files"],
            ["--to", ".", "--null"], ["--to", ".", "--type", FILE],
            ["--to", ".", "--list-types"], ["--to", ".", "--prefer", "txt"],
            ["--to", ".", "--help"], ["--to", ".", "--version"],
        ]
        for name, value in (("--type", PLAIN), ("--item", "0"),
                            ("--pboard", "general"), ("--prefer", "txt"), ("--to", ".")):
            invalid.append([name, value, name, value])
        for name in ("--files", "--null", "--list-types", "--help", "--version"):
            invalid.append([name, name])
            invalid.append([name + "=false"])
        invalid.extend([["-t", PLAIN, "--type", PLAIN], ["-h", "--help"],
                        ["-v", "--version"], ["-pboard", "general", "--pboard", "general"],
                        ["-Prefer", "txt", "--prefer", "txt"]])
        for arguments in invalid:
            self.paste(arguments, status=2)
        self.paste(expected=b"pbjpaste generated preservation sentinel")

    def cleanup(self):
        errors = []
        reset = b"pbjpaste verification complete"
        for board in sorted(self.touched):
            try:
                result = invoke([ROOT / "bin/pbjcopy", "--pboard", board], data=reset)
                require(result.returncode == 0 and not result.stdout and not result.stderr,
                        "completion fixture write failed")
                metadata = inspect(board)
                observed = inspect(board, metadata["count"], [PLAIN])
                require(observed["payloads"][PLAIN] == [base64.b64encode(reset).decode()],
                        "completion fixture readback failed")
            except Exception:
                errors.append("clipboard reset failed for " + board)
        try:
            shutil.rmtree(self.scratch)
        except OSError:
            errors.append("scratch removal failed")
        return {"clipboard_reset": not any("clipboard" in error for error in errors),
                "clipboard_reset_boards": sorted(self.touched),
                "clipboard_untouched": not self.touched,
                "scratch_removed": not self.scratch.exists(), "errors": errors}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--doctor", action="store_true", help="read-only version/help readiness check")
    parser.add_argument("--allow-clipboard", action="store_true", help="authorize replacing only the general clipboard with generated fixtures")
    parser.add_argument("--evidence", type=Path, help="new evidence directory outside temporary scratch")
    parser.add_argument("--feature", choices=("all", *FEATURES), default="all")
    args = parser.parse_args()
    host = doctor()
    if args.doctor:
        print(json.dumps({"doctor": "ready", "host": host}, indent=2))
        return 0
    require(args.feature == "information" or args.allow_clipboard,
            "obtain general-clipboard consent, then pass --allow-clipboard")
    require(args.evidence is not None, "provide --evidence NEW_DIR")
    evidence = args.evidence.resolve()
    require(not any(part.startswith("pbjpaste-verify-") for part in evidence.parts),
            "evidence must be outside cleanup-owned pbjpaste-verify scratch")
    evidence.mkdir(parents=True, exist_ok=False)
    scratch = Path(tempfile.mkdtemp(prefix="pbjpaste-verify-", dir="/tmp"))
    proof = Proof(scratch)
    report = {"host": host, "feature": args.feature, "verdict": "NOT VERIFIED",
              "actions": proof.actions,
              "unexercised": ["find, ruler, and font pasteboards: general-only clipboard consent"]}
    passed = False
    try:
        for feature in FEATURES:
            if args.feature in ("all", feature):
                getattr(proof, feature)()
        passed = True
    except Exception as error:
        report["error"] = redact(error, scratch)
    finally:
        cleanup = proof.cleanup()
        report.update(cleanup)
        if passed and not cleanup["errors"] and cleanup["scratch_removed"]:
            report["verdict"] = "VERIFIED"
        (evidence / "report.json").write_text(json.dumps(report, indent=2) + "\n")
    print(json.dumps({"verdict": report["verdict"], "evidence": evidence.name + "/report.json",
                      "host": host, "actions": len(proof.actions),
                      "clipboard_reset": report["clipboard_reset"],
                      "scratch_removed": report["scratch_removed"]}, indent=2))
    return 0 if report["verdict"] == "VERIFIED" else 1


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as error:
        print(json.dumps({"verdict": "INCONCLUSIVE", "error": redact(error)}))
        raise SystemExit(1)
