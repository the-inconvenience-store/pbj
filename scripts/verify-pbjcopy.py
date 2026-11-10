#!/usr/bin/env python3
"""Verification-owned driver. Clipboard writes require explicit authorization."""

import argparse
import base64
import hashlib
import json
import os
from pathlib import Path
import platform
import shutil
import struct
import subprocess
import tempfile
import zlib
from urllib.parse import unquote, urlparse

ROOT = Path(__file__).resolve().parent.parent
SCRIPTS = ROOT / "scripts"
PLAIN = "public.utf8-plain-text"


def invoke(arguments, *, data=None, stdin=None, env=None, timeout=40):
    return subprocess.run(arguments, input=data, stdin=stdin, capture_output=True,
                          env=env, timeout=timeout, cwd=ROOT)


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def native_script(name, *arguments):
    result = invoke(["/usr/bin/osascript", str(SCRIPTS / name), *map(str, arguments)])
    require(result.returncode == 0, result.stderr.decode(errors="replace"))
    return result.stdout.decode()


def observe(type_name):
    result = invoke(["/usr/bin/osascript", "-l", "JavaScript",
                     str(SCRIPTS / "inspect-pasteboard.js"), type_name])
    require(result.returncode == 0, "native clipboard observation failed")
    values = json.loads(result.stdout)
    if type_name == "--types":
        return values
    return [None if value is None else base64.b64decode(value) for value in values]


def fixture_png():
    def chunk(kind, data):
        return (struct.pack(">I", len(data)) + kind + data +
                struct.pack(">I", zlib.crc32(kind + data)))

    colors = [(255, 40, 40, 255), (40, 80, 255, 255),
              (40, 220, 80, 255), (255, 220, 40, 255)]
    rows = []
    for y in range(256):
        row = bytearray([0])
        for x in range(256):
            row.extend(colors[(y >= 128) * 2 + (x >= 128)])
        rows.append(row)
    header = struct.pack(">IIBBBBB", 256, 256, 8, 6, 0, 0, 0)
    return (b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", header) +
            chunk(b"IDAT", zlib.compress(b"".join(rows))) + chunk(b"IEND", b""))


class Proof:
    def __init__(self, binary, scratch, evidence):
        self.binary = binary
        self.scratch = scratch
        self.evidence = evidence
        self.actions = []
        self.environment = {**os.environ, "LC_ALL": "en_US.UTF-8"}

    def record(self, action, result):
        self.actions.append({"action": action, "observed": result})

    def copy(self, arguments=(), data=b"", expected=0, stdin=None, environment=None):
        result = invoke([str(self.binary), *map(str, arguments)],
                        data=None if stdin is not None else data, stdin=stdin,
                        env=environment or self.environment)
        require(result.returncode == expected, "unexpected pbjcopy exit status")
        require(not result.stdout, "copy or rejected invocation emitted stdout")
        if expected == 0:
            require(not result.stderr, "successful copy emitted stderr")
        return result

    def payload(self, type_name, expected):
        # Do not log unexpected clipboard contents: another app may have copied.
        require(observe(type_name) == [expected], "clipboard fixture mismatch for " + type_name)
        self.record("read " + type_name + " after pbjcopy exited",
                    {"bytes": len(expected), "sha256": hashlib.sha256(expected).hexdigest()})

    def metadata(self):
        for arguments in [["--help"], ["-help"], ["--version"], ["--list-types"]]:
            result = invoke([str(self.binary), *arguments], data=b"")
            require(result.returncode == 0 and not result.stderr, "informational command failed")
            if arguments[-1] in ("--help", "-help"):
                require(b"--type" in result.stdout and b"--file" in result.stdout,
                        "generated help does not expose copy options")
            self.record("pbjcopy " + " ".join(arguments), {"exit": result.returncode})
        for arguments in [["--help", "-t", "public.html"], ["-pboard", "unknown"],
                          ["-f", "missing", "--plain", "text"]]:
            self.copy(arguments, expected=2)
            self.record("reject ambiguous invocation", {"arguments": arguments, "exit": 2})

    def representations(self):
        plain = "pbjcopy plain ☃".encode()
        self.copy(["-pboard", "general"], plain)
        self.payload(PLAIN, plain)
        markup = b"<b>literal source</b>"
        self.copy(data=markup)
        self.payload(PLAIN, markup)
        html = b"<b>pbjcopy bold</b>"
        self.copy(["-t", "public.html", "-p", "pbjcopy bold"], html)
        self.payload("public.html", html)
        self.payload(PLAIN, b"pbjcopy bold")
        self.copy(["-t", "public.html", "--plain", ""], html)
        self.payload(PLAIN, b"")
        raw = b"\xff\x00\x7f"
        self.copy(["-t", "com.example.raw", "-t", "com.example.raw",
                   "-t", "com.example.alternate"], raw)
        self.payload("com.example.raw", raw)
        self.payload("com.example.alternate", raw)
        require(observe("--types") == [["com.example.raw", "com.example.alternate"]],
                "explicit types were not deduplicated or detection was not overridden")
        formats = [("public.png", fixture_png()), ("public.jpeg", b"\xff\xd8\xff"),
                   ("public.tiff", b"II\x2a\x00"), ("public.tiff", b"MM\x00\x2a"),
                   ("com.adobe.pdf", b"%PDF-1.7\n"), ("public.rtf", b"{\\rtf1 hello}"),
                   ("com.adobe.encapsulated-postscript", b"%!PS-Adobe-3.0 EPSF-3.0\n")]
        for type_name, data in formats:
            self.copy(data=data)
            self.payload(type_name, data)
        self.copy(data=b"")
        self.payload(PLAIN, b"")
        latin = {**self.environment, "LC_ALL": "fr_FR.ISO8859-1", "LC_CTYPE": "C", "LANG": "C"}
        self.copy(data=b"caf\xe9", environment=latin)
        self.payload(PLAIN, "café".encode())
        self.structured()

    def structured(self):
        fixtures = [
            ("public.json", b'{"answer":42}', []),
            ("public.json", ' \n[true, "☃"]\t'.encode(), []),
            ("public.xml", b'<?xml version="1.0"?><root/>', []),
            ("public.html", b'<!DOCTYPE html><html><body><br>Hello</body></html>', []),
            ("public.html", b'<?xml version="1.0"?><html><body>Hello</body></html>', []),
            ("public.html", b'<html>Hello</html>', ["--plain", "Hello"]),
            (PLAIN, b'{"answer":42}', ["--type", PLAIN]),
            ("com.example.raw", b'<html>Hello</html>', ["--type", "com.example.raw"]),
        ]
        for type_name, data, arguments in fixtures:
            self.copy(arguments, data)
            expected_types = [type_name] + ([PLAIN] if "--plain" in arguments else [])
            require(observe("--types") == [expected_types], "structured type set differs")
            self.payload(type_name, data)
            if "--plain" in arguments:
                self.payload(PLAIN, b"Hello")
            extracted = invoke([str(ROOT / "bin/pbjpaste"), "--type", type_name])
            require(extracted.returncode == 0 and extracted.stdout == data and not extracted.stderr,
                    "pbjpaste exact structured payload differs")
            self.record("pbjcopy " + " ".join(arguments) + " then pbjpaste --type " + type_name,
                        {"types": expected_types, "exact_payload": True})
        for data in [b"true", b"42", b'"hello"', b'{"broken":}', b"{} trailing",
                     b"<root/>", b"<html>unfinished", b"<b>fragment</b>",
                     b'<?xml broken?><root/>', b'<?xml version="1.0"?><a/><b/>',
                     b'<?xml version="1.0"?><root/><!garbage>',
                     b'<?xml version="1.0"?><root><!garbage></root>',
                     b'<?xml version="1.0"?><root><?xml version="1.0"?></root>',
                     b'<?xml version="1.0"?><root a="1" a="2"/>',
                     b"ordinary text"]:
            self.copy(data=data)
            require(observe("--types") == [[PLAIN]], "ambiguous input was mislabeled")
            self.payload(PLAIN, data)
        self.copy(data=b"generated structured failure sentinel")
        self.copy(["--plain", "conflict"], b'{"broken":}', expected=1)
        self.payload(PLAIN, b"generated structured failure sentinel")

    def failures(self):
        sentinel = b"pbjcopy keep this generated fixture"
        self.copy(data=sentinel)
        for arguments, data, status in [([], b"\xff", 1),
                                         (["--plain", "conflict"], b"original", 1),
                                         (["-t", PLAIN, "--plain", "conflict"], b"original", 1)]:
            self.copy(arguments, data, expected=status)
            self.payload(PLAIN, sentinel)
        descriptor = os.open(self.scratch, os.O_RDONLY)
        try:
            self.copy(expected=1, stdin=descriptor)
        finally:
            os.close(descriptor)
        self.payload(PLAIN, sentinel)
        self.copy(["-f", self.scratch / "missing-file"], expected=1)
        self.payload(PLAIN, sentinel)
        self.record("pre-publication failures", "existing generated clipboard payload preserved")

    def files(self):
        first = self.scratch / "pbjcopy A.txt"
        second = self.scratch / "pbjcopy-B.txt"
        first.write_bytes(b"pbjcopy file A")
        second.write_bytes(b"pbjcopy file B")
        descriptor = os.open(self.scratch, os.O_RDONLY)
        try:
            self.copy(["-f", first, "-f", second, "-f", first], stdin=descriptor)
        finally:
            os.close(descriptor)
        values = observe("public.file-url")
        require(all(value is not None for value in values), "file URL representation is missing")
        locations = [urlparse(value.decode()) for value in values]
        require(all(location.scheme == "file" for location in locations), "reference is not a file URL")
        paths = [unquote(location.path) for location in locations]
        require(paths == [str(first), str(second)], "file references did not survive exit in selected order")
        self.record("pbjcopy -f FIRST -f SECOND -f FIRST with unreadable stdin",
                    "two deduplicated ordered file URLs persisted after process exit")
        destination = self.scratch / "pbjcopy-destination"
        destination.mkdir()
        native_script("paste-into-finder.applescript", destination, first.name, second.name)
        require((destination / first.name).read_bytes() == b"pbjcopy file A", "Finder first file differs")
        require((destination / second.name).read_bytes() == b"pbjcopy file B", "Finder second file differs")
        self.record("Finder Edit → Paste", "both files created with exact generated contents; window closed")

    def apps(self):
        capture = SCRIPTS / "capture-verification-window.js"
        plain = self.scratch / (self.scratch.name + "-plain.txt")
        plain.write_text("")
        self.copy(data="pbjcopy plain ☃".encode())
        native_script("paste-into-textedit.applescript", plain, "pbjcopy plain ☃", "plain",
                      self.evidence / "textedit-plain.png", capture)
        require(plain.read_text() == "pbjcopy plain ☃", "TextEdit saved plain text differs")
        self.record("TextEdit Cmd+V", "exact plain text saved; generated document closed")
        rich = self.scratch / (self.scratch.name + "-rich.rtf")
        rich.write_text(r"{\rtf1\ansi\deff0{\fonttbl{\f0 Helvetica;}}\f0\fs24 }")
        self.copy(["-t", "public.html", "-p", "pbjcopy bold"], b"<b>pbjcopy bold</b>")
        rendered = native_script("paste-into-textedit.applescript", rich, "pbjcopy bold", "rich",
                                 self.evidence / "textedit-rich.png", capture)
        text, font = rendered.rstrip("\n").split("\x1f")
        require(text.strip() == "pbjcopy bold" and "bold" in font.lower(), "TextEdit did not render HTML as bold")
        self.record("TextEdit rich Cmd+V", {"text": text.strip(), "font": font, "saved": "RTF", "closed": True})
        image = self.scratch / (self.scratch.name + "-image.png")
        image.write_bytes(fixture_png())
        self.copy(data=image.read_bytes())
        native_script("paste-into-preview.applescript", image, self.evidence / "preview-image.png", capture)
        self.payload("public.png", image.read_bytes())
        self.record("Preview File → New from Clipboard", "generated 256×256 four-color image rendered; both generated windows closed")

    def attachments(self):
        home = self.scratch / "attachment-home"
        home.mkdir()
        environment = {**self.environment, "HOME": str(home)}
        cache = home / "Library/Caches"
        fixtures = [("same.bin", b"\x00\xff\x7f"), ("same.bin", fixture_png()),
                    ("empty", b""), ("-help", b"literal filename"),
                    ("pbjcopy-owned.bin", b"pbjcopy attachment\x00\xff\n")]
        retained = []
        for name, data in fixtures:
            self.copy(["--attachment", name], data, environment=environment)
            references = observe("public.file-url")
            require(len(references) == 1 and references[0] is not None, "missing attachment reference")
            location = urlparse(references[0].decode())
            require(location.scheme == "file" and not location.netloc, "attachment is not a local file URL")
            path = Path(unquote(location.path))
            require(path.is_relative_to(cache) and path.name == name, "attachment path or filename differs")
            require(path.read_bytes() == data, "backing bytes differ after pbjcopy exit")
            require(path.stat().st_mode & 0o777 == 0o600, "attachment file is not private")
            require(path.parent.stat().st_mode & 0o777 == 0o700, "attachment directory is not private")
            require(all(path != previous for previous, _ in retained), "attachment reused a backing path")
            extracted = invoke([str(self.binary.with_name("pbjpaste")), "--files", "--null"])
            require(extracted.returncode == 0 and extracted.stdout == os.fsencode(path) + b"\0",
                    "pbjpaste did not extract the exact attachment reference")
            retained.append((path, data))
            self.record("pbjcopy --attachment " + name + " then pbjpaste --files --null",
                        {"filename": name, "bytes": len(data), "sha256": hashlib.sha256(data).hexdigest(),
                         "file_mode": "0600", "directory_mode": "0700", "persisted_after_exit": True})
        before = set(cache.iterdir())
        descriptor = os.open(self.scratch, os.O_RDONLY)
        try:
            self.copy(["--attachment", "read-fails.bin"], expected=1, stdin=descriptor, environment=environment)
        finally:
            os.close(descriptor)
        bad_home = self.scratch / "not-a-directory"
        bad_home.write_bytes(b"owned preparation failure fixture")
        self.copy(["--attachment", "cache-fails.bin"], b"not published", expected=1,
                  environment={**environment, "HOME": str(bad_home)})
        require(set(cache.iterdir()) == before, "failed preparation leaked attachment directories")
        require(observe("public.file-url") == references, "failed preparation replaced the clipboard")
        self.record("attachment read/cache preparation failures", "prior reference preserved; no leaked directories")
        self.copy(data=b"generated clipboard reset before retention check")
        for path, data in retained:
            require(path.read_bytes() == data, "clipboard replacement removed an older attachment")
        path, data = retained[-1]
        self.copy(["--attachment", path.name], data, environment=environment)
        references = observe("public.file-url")
        app_path = Path(unquote(urlparse(references[0].decode()).path))
        require(app_path.is_relative_to(cache) and app_path.read_bytes() == data, "app fixture differs")
        outcome = native_script("paste-attachment-into-mail.applescript",
                                self.scratch.name + "-attachment", path.name,
                                self.evidence / "mail-attachment.png", SCRIPTS / "capture-verification-window.js")
        self.record("Mail body Cmd+V after pbjcopy exit", outcome.strip())
        self.record("attachment cleanup policy", "backing files retained across clipboard replacement; owned cache removed with scratch")


def doctor(binary, graphical):
    result = invoke([str(binary), "--version"], data=b"")
    require(result.returncode == 0, "build bin/pbjcopy before verification")
    if graphical:
        access = invoke(["/usr/bin/osascript", "-l", "JavaScript", "-e",
                         'ObjC.import("ApplicationServices"); $.AXIsProcessTrusted()'])
        require(access.returncode == 0 and access.stdout.strip() == b"true", "Accessibility permission is required")
        capture = invoke(["/usr/bin/swift", "-e", "import CoreGraphics; print(CGPreflightScreenCaptureAccess())"])
        require(capture.returncode == 0 and capture.stdout.strip() == b"true", "Screen Recording permission is required")
    return {"version": result.stdout.decode().strip(), "architecture": platform.machine(),
            "macOS": invoke(["/usr/bin/sw_vers", "-productVersion"]).stdout.decode().strip()}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, default=Path("bin/pbjcopy"))
    parser.add_argument("--feature", choices=["all", "metadata", "representations", "failures", "files", "apps", "attachments"], default="all")
    parser.add_argument("--doctor", action="store_true")
    parser.add_argument("--allow-clipboard", action="store_true")
    parser.add_argument("--allow-mail", action="store_true",
                        help="Consent to paste into and discard a generated, unsent Mail compose window")
    parser.add_argument("--evidence", type=Path)
    args = parser.parse_args()
    binary = (ROOT / args.binary).resolve()
    host = doctor(binary, args.feature in ("all", "files", "apps", "attachments"))
    if args.doctor:
        print(json.dumps({"doctor": "ready", **host}, indent=2))
        return
    require(args.feature == "metadata" or args.allow_clipboard, "obtain consent, then pass --allow-clipboard")
    require(args.feature not in ("all", "attachments") or args.allow_mail,
            "obtain Mail compose-window consent, then pass --allow-mail")
    require(args.evidence is not None, "provide --evidence outside cleanup-owned scratch state")
    evidence = args.evidence.resolve()
    evidence.mkdir(parents=True, exist_ok=False)
    scratch = Path(tempfile.mkdtemp(prefix="pbjcopy-verify-", dir="/tmp"))
    proof = Proof(binary, scratch, evidence)
    report = {"host": host, "feature": args.feature, "verdict": "NOT VERIFIED", "actions": proof.actions}
    try:
        for feature in ("metadata", "representations", "failures", "files", "apps", "attachments"):
            if args.feature in ("all", feature):
                getattr(proof, feature)()
        report["verdict"] = "VERIFIED"
    except Exception as failure:
        report["error"] = str(failure).replace(str(ROOT), "$REPO").replace(str(scratch), "$SCRATCH")
        raise
    finally:
        cleanup_errors = []
        if args.allow_clipboard:
            try:
                reset = invoke([str(binary)], data=b"pbjcopy verification complete", env=proof.environment)
                report["clipboard_reset"] = reset.returncode == 0
            except subprocess.SubprocessError:
                report["clipboard_reset"] = False
            if not report["clipboard_reset"]:
                cleanup_errors.append("clipboard reset failed")
        try:
            shutil.rmtree(scratch)
        except OSError:
            cleanup_errors.append("scratch removal failed")
        report["scratch_removed"] = not scratch.exists()
        if cleanup_errors:
            report["verdict"] = "NOT VERIFIED"
            report["cleanup_errors"] = cleanup_errors
        (evidence / "report.json").write_text(json.dumps(report, indent=2) + "\n")
    print(json.dumps({"verdict": report["verdict"], "evidence": str(args.evidence),
                      "actions": len(proof.actions), "scratch_removed": report["scratch_removed"]}, indent=2))
    require(report["verdict"] == "VERIFIED", "verification or cleanup failed; inspect the evidence report")


if __name__ == "__main__":
    main()
