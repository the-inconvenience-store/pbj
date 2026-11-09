#!/usr/bin/env python3
"""Prove a GoReleaser snapshot through a real, isolated Homebrew source install."""

import argparse
import functools
import hashlib
import http.server
import json
import os
from pathlib import Path
import platform
import re
import shutil
import signal
import subprocess
import tarfile
import tempfile
import threading
import uuid

ROOT = Path(__file__).resolve().parent.parent
ARCHIVE = "pbj_0.0.0_source.tar.gz"
TAG = "v0.0.0"


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def redact(value, scratch=None, evidence=None):
    text = str(value)
    for prefix, replacement in ((str(scratch) if scratch else "", "$SCRATCH"),
                                (str(evidence) if evidence else "", "$EVIDENCE"),
                                (str(ROOT), "$REPO"), (str(Path.home()), "$HOME")):
        if prefix:
            text = text.replace(str(Path(prefix).resolve()), replacement)
            text = text.replace(prefix, replacement)
    # Tool output must never retain Git identity or signature material.
    text = re.sub(r"-----BEGIN (?:PGP|SSH) SIGNATURE-----.*?-----END (?:PGP|SSH) SIGNATURE-----",
                  "[signature redacted]", text, flags=re.S)
    text = re.sub(r"(?im)^.*(?:gpg:|signature|signing key|fingerprint|"
                  r"^author:|^committer:|^commit:|^signed-off-by:|^co-authored-by:).*$",
                  "[identity/signature redacted]", text)
    text = re.sub(r"[\w.+-]+@[\w.-]+\.[A-Za-z]{2,}", "[email redacted]", text)
    return text


def environment(scratch):
    env = dict(os.environ)
    for key in list(env):
        if any(word in key.upper() for word in ("TOKEN", "SECRET", "PASSWORD")):
            env.pop(key)
    env.update({
        "HOMEBREW_CACHE": str(scratch / "cache"),
        "HOMEBREW_LOGS": str(scratch / "logs"),
        "HOMEBREW_TEMP": str(scratch / "temp"),
        "HOMEBREW_NO_AUTO_UPDATE": "1",
        "HOMEBREW_NO_INSTALL_CLEANUP": "1",
        "HOMEBREW_NO_ANALYTICS": "1",
        "HOMEBREW_NO_INSTALL_UPGRADE": "1",
        "HOMEBREW_NO_INSTALLED_DEPENDENTS_CHECK": "1",
        "HOMEBREW_NO_ENV_HINTS": "1",
        "HOMEBREW_NO_COLOR": "1",
        "GOCACHE": str(scratch / "go-build"),
        "GOTOOLCHAIN": "local",
        "GIT_CONFIG_GLOBAL": os.devnull,
        "GIT_CONFIG_NOSYSTEM": "1",
        "GIT_TERMINAL_PROMPT": "0",
        "RUBOCOP_CACHE_ROOT": str(scratch / "rubocop"),
        "TMPDIR": str(scratch / "temp"),
        "LC_ALL": "en_US.UTF-8",
    })
    for name in ("cache", "logs", "temp", "go-build", "rubocop"):
        (scratch / name).mkdir()
    return env


class Driver:
    def __init__(self, scratch, evidence=None):
        self.scratch = scratch
        self.evidence = evidence
        self.env = environment(scratch)
        self.actions = []
        self.brew = shutil.which("brew")
        self.tap = "pbj-local/release-" + uuid.uuid4().hex[:12]
        self.tap_path = None
        self.formula = self.tap + "/pbj"
        self.cellar = None
        self.prefix = None
        self.install_started = False
        self.before = None
        self.server = None
        self.server_thread = None
        self.links = {}

    def run(self, args, *, timeout=120, check=True):
        args = list(map(str, args))
        action = {"command": [redact(arg, self.scratch, self.evidence) for arg in args],
                  "timeout_seconds": timeout}
        self.actions.append(action)
        process = subprocess.Popen(args, cwd=ROOT, env=self.env, stdin=subprocess.DEVNULL,
                                   stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                   start_new_session=True)
        try:
            stdout, stderr = process.communicate(timeout=timeout)
        except subprocess.TimeoutExpired:
            os.killpg(process.pid, signal.SIGTERM)
            try:
                stdout, stderr = process.communicate(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL)
                stdout, stderr = process.communicate(timeout=5)
            action["timed_out"] = True
            action["exit"] = process.returncode
            action["stdout"] = redact(stdout.decode(errors="replace"), self.scratch, self.evidence)
            action["stderr"] = redact(stderr.decode(errors="replace"), self.scratch, self.evidence)
            raise RuntimeError("subprocess deadline exceeded: " + " ".join(action["command"]))
        action.update(exit=process.returncode,
                      stdout=redact(stdout.decode(errors="replace"), self.scratch, self.evidence),
                      stderr=redact(stderr.decode(errors="replace"), self.scratch, self.evidence))
        require(not check or process.returncode == 0,
                "command failed; see action log: " + " ".join(action["command"]))
        return process.returncode, stdout, stderr

    def brew_run(self, *args, **kwargs):
        return self.run([self.brew, *args], **kwargs)

    def helper(self, *args):
        return self.run(["mise", "exec", "--", "go", "run", "./scripts/release", *args],
                        timeout=180)

    def doctor(self):
        require(platform.system() == "Darwin", "macOS is required")
        require(self.brew is not None, "Homebrew must already be installed")
        require(shutil.which("mise") is not None, "mise must already be installed")
        versions = {}
        for name, command in (
                ("brew", [self.brew, "--version"]),
                ("mise", ["mise", "--version"]),
                ("cog", ["mise", "exec", "--", "cog", "--version"]),
                ("goreleaser", ["mise", "exec", "--", "goreleaser", "--version"]),
                ("go", ["mise", "exec", "--", "go", "version"])):
            _, stdout, _ = self.run(command)
            require(bool(stdout.strip()), name + " returned no version")
            versions[name] = redact(stdout.decode().strip(), self.scratch, self.evidence)
        self.helper("--help")
        self.helper("validate", "--tag", TAG)
        self.prefix = Path(self.brew_run("--prefix")[1].decode().strip())
        self.cellar = Path(self.brew_run("--cellar")[1].decode().strip())
        require(self.prefix.is_dir() and os.access(self.prefix, os.W_OK),
                "Homebrew prefix is not writable; do not change permissions")
        require(self.cellar.is_dir() and os.access(self.cellar, os.W_OK),
                "Homebrew Cellar is not writable; do not change permissions")
        # Reading help detects isolation support without installing anything.
        help_output = self.brew_run("install", "--help")[1].decode()
        return {"system": "macOS", "architecture": platform.machine(), "tools": versions,
                "skip_link_supported": "--skip-link" in help_output,
                "helper": "help and stable snapshot-tag validation succeeded",
                "scope": "readiness only; no formula/tap install, release, or clipboard access"}

    def inventory(self):
        return set(self.brew_run("list", "--formula", "--versions")[1].decode().splitlines())

    def link_state(self):
        result = {}
        for name in ("pbjc", "pbjp"):
            path = self.prefix / "bin" / name
            if path.is_symlink():
                result[name] = {"kind": "symlink", "target": os.readlink(path)}
            elif path.exists():
                require(path.is_file(), "unexpected non-file command at Homebrew prefix")
                result[name] = {"kind": "file", "sha256": hashlib.sha256(path.read_bytes()).hexdigest()}
            else:
                result[name] = {"kind": "absent"}
        return result

    def source(self, source):
        require(source.name == ARCHIVE and source.is_file(),
                "provide the real GoReleaser snapshot archive " + ARCHIVE)
        expected = {"go.mod", "go.sum", "cmd/pbjc/main.go", "cmd/pbjp/main.go"}
        with tarfile.open(source, "r:gz") as archive:
            names = set()
            for member in archive:
                require(member.name.startswith("pbj_0.0.0/"), "source archive prefix differs")
                relative = Path(member.name).parts[1:]
                require(".." not in relative, "source archive has unsafe path")
                names.add("/".join(relative))
            require(expected <= names, "source archive lacks real Go entry points/modules")
        checksum = hashlib.sha256()
        with source.open("rb") as stream:
            for chunk in iter(lambda: stream.read(1024 * 1024), b""):
                checksum.update(chunk)
        digest = checksum.hexdigest()
        self.actions.append({"action": "inspect supplied GoReleaser source archive",
                             "basename": source.name, "sha256": digest,
                             "source_prefix": "pbj_0.0.0/", "entries": len(names),
                             "provenance": "parent supplies actual GoReleaser snapshot; not a GitHub release"})
        return digest

    def formula_info(self):
        data = json.loads(self.brew_run("info", "--json=v2", "--formula", self.formula)[1])
        rows = data.get("formulae", [])
        require(len(rows) == 1, "Homebrew loader did not resolve one formula")
        return rows[0]

    def drive(self, source, host):
        digest = self.source(source)
        # Formula names share Cellar paths across taps, even when taps are unique.
        require(not os.path.lexists(self.cellar / "pbj"),
                "existing pbj Cellar path: abort rather than replace/uninstall the user's installation")
        self.links = self.link_state()
        if not host["skip_link_supported"]:
            require(all(row["kind"] == "absent" for row in self.links.values()),
                    "brew lacks --skip-link and existing pbjc/pbjp links conflict; abort")
        self.before = self.inventory()
        original = self.scratch / "generated" / "pbj.rb"
        original.parent.mkdir()
        self.helper("formula", "--tag", TAG, "--source", str(source), "--output", str(original))
        generated = original.read_text()
        canonical = ("https://github.com/the-inconvenience-store/pbj/releases/download/" +
                     TAG + "/" + ARCHIVE)
        require(generated.count('url "' + canonical + '"') == 1,
                "generator's canonical source URL differs")
        require(digest in generated, "generator did not checksum the actual archive")
        (self.evidence / "generated-pbj.rb").write_text(generated)
        # An equivalent tap-new without Git avoids inheriting personal Git identity.
        repository = Path(self.brew_run("--repository")[1].decode().strip())
        tap_parent = repository / "Library" / "Taps" / "pbj-local"
        tap_parent.mkdir(exist_ok=True)
        candidate = tap_parent / ("homebrew-" + self.tap.split("/")[1])
        candidate.mkdir(exist_ok=False)
        self.tap_path = candidate
        formula_dir = self.tap_path / "Formula"
        formula_dir.mkdir()
        formula_path = formula_dir / "pbj.rb"
        formula_path.write_text(generated)
        loaded = self.formula_info()
        require(loaded["versions"]["stable"] == "0.0.0", "Homebrew loaded a different version")
        require(loaded["urls"]["stable"]["url"] == canonical,
                "Homebrew loader changed the generated source URL")
        require(loaded["urls"]["stable"]["checksum"] == digest,
                "Homebrew loader changed the generated checksum")
        self.brew_run("style", str(formula_path), timeout=600)
        served = self.scratch / "served"
        served.mkdir()
        shutil.copyfile(source, served / ARCHIVE)

        class QuietHandler(http.server.SimpleHTTPRequestHandler):
            def log_message(self, format, *args):
                pass

        handler = functools.partial(QuietHandler, directory=str(served))
        self.server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), handler)
        self.server.daemon_threads = True
        self.server_thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.server_thread.start()
        local_url = "http://127.0.0.1:" + str(self.server.server_port) + "/" + ARCHIVE
        adapted = generated.replace('url "' + canonical + '"', 'url "' + local_url + '"', 1)
        require(adapted.replace(local_url, canonical, 1) == generated,
                "local formula adaptation changed more than the source URL")
        formula_path.write_text(adapted)
        (self.evidence / "local-pbj.rb").write_text(adapted)
        loaded = self.formula_info()
        require(loaded["urls"]["stable"]["url"] == local_url and
                loaded["urls"]["stable"]["checksum"] == digest and
                loaded["versions"]["stable"] == "0.0.0", "local loader result differs")
        self.brew_run("style", str(formula_path), timeout=600)
        options = ["--skip-link"] if host["skip_link_supported"] else []
        self.install_started = True
        self.brew_run("install", "--formula", "--build-from-source", *options,
                      self.formula, timeout=1800)
        keg = self.cellar / "pbj" / "0.0.0"
        receipt_path = keg / "INSTALL_RECEIPT.json"
        receipt = json.loads(receipt_path.read_text())
        require(receipt.get("poured_from_bottle") is False,
                "receipt does not prove a source build")
        require(receipt.get("built_as_bottle") is False,
                "receipt unexpectedly describes a bottle build")
        require(receipt.get("source", {}).get("tap") == self.tap,
                "installed receipt belongs to a different tap")
        (self.evidence / "install-receipt.json").write_text(
            redact(json.dumps(receipt, indent=2), self.scratch, self.evidence) + "\n")
        self.brew_run("test", self.formula, timeout=180)
        for name in ("pbjc", "pbjp"):
            executable = keg / "bin" / name
            require(executable.is_file() and os.access(executable, os.X_OK),
                    "installed executable is missing")
            _, stdout, stderr = self.run([executable, "--version"])
            require(stdout.decode().strip() == name + " 0.0.0" and not stderr,
                    "installed version does not match explicit formula tag")
            _, stdout, stderr = self.run([executable, "--help"])
            require(b"Usage:" in stdout and not stderr, "installed help smoke failed")
            code, stdout, stderr = self.run([executable, "--pbj-release-invalid"], check=False)
            require(code == 2 and not stdout and bool(stderr), "invalid-argument smoke failed")
        self.actions.append({"action": "actual installed binaries smoke", "clipboard_access": False,
                             "version": "0.0.0", "poured_from_bottle": False,
                             "generated_formula_changed_fields": ["url"],
                             "local_only_tap": self.tap})

    def cleanup(self):
        errors = []
        added = []
        if self.server is not None:
            self.server.shutdown()
            self.server.server_close()
            self.server_thread.join(timeout=5)
            if self.server_thread.is_alive():
                errors.append("loopback server thread did not stop")
        if self.install_started:
            try:
                # Never uninstall a keg whose durable ownership record is missing/different.
                keg_root = self.cellar / "pbj"
                if keg_root.exists():
                    receipts = list(keg_root.glob("*/INSTALL_RECEIPT.json"))
                    require(bool(receipts) and all(
                        json.loads(path.read_text()).get("source", {}).get("tap") == self.tap
                        for path in receipts), "cannot safely identify the installed keg as owned")
                    self.brew_run("uninstall", "--formula", self.formula, timeout=180)
                    require(not keg_root.exists(), "owned pbj Cellar path remains")
            except Exception as error:
                errors.append(redact(error, self.scratch, self.evidence))
        if self.before is not None:
            try:
                added = sorted(self.inventory() - self.before)
            except Exception as error:
                errors.append(redact(error, self.scratch, self.evidence))
        if self.links:
            try:
                require(self.link_state() == self.links, "original Homebrew command links changed")
            except Exception as error:
                errors.append(redact(error, self.scratch, self.evidence))
        if self.tap_path is not None:
            try:
                # Exact random directory owned by this process; never untap a user tap.
                shutil.rmtree(self.tap_path)
                try:
                    self.tap_path.parent.rmdir()
                except OSError:
                    pass  # A pre-existing/shared tap namespace must remain.
            except Exception as error:
                errors.append(redact(error, self.scratch, self.evidence))
        if self.evidence is not None:
            try:
                for path in (self.scratch / "logs").rglob("*"):
                    if path.is_file():
                        destination = self.evidence / "homebrew-logs" / path.relative_to(self.scratch / "logs")
                        destination.parent.mkdir(parents=True, exist_ok=True)
                        destination.write_text(redact(path.read_text(errors="replace"), self.scratch, self.evidence))
            except OSError as error:
                errors.append("log evidence capture failed: " + redact(error, self.scratch, self.evidence))
        try:
            shutil.rmtree(self.scratch)
        except OSError as error:
            errors.append(redact(error, self.scratch, self.evidence))
        return {"errors": errors, "scratch_removed": not self.scratch.exists(),
                "tap_removed": self.tap_path is None or not self.tap_path.exists(),
                "added_dependencies_retained": added,
                "dependency_policy": "never uninstall dependencies; no upgrades/autoremove/analytics",
                "other_side_effects": "Homebrew style may install its own Ruby/bundler tooling",
                "clipboard_untouched": True}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--doctor", action="store_true", help="read-only tool/helper readiness; no installs")
    parser.add_argument("--source", type=Path, help="actual GoReleaser pbj_0.0.0_source.tar.gz snapshot")
    parser.add_argument("--allow-install", action="store_true", help="consent to owned source install and possible retained dependencies")
    parser.add_argument("--evidence", type=Path, help="new evidence directory outside cleanup-owned scratch")
    args = parser.parse_args()
    evidence = None
    if not args.doctor:
        require(args.allow_install, "obtain install consent and pass --allow-install")
        require(args.source is not None and args.evidence is not None, "provide --source and --evidence NEW_DIR")
        evidence = args.evidence.resolve()
        require(not any(part.startswith("pbj-release-verify-") for part in evidence.parts),
                "evidence must be outside cleanup-owned release scratch")
        evidence.mkdir(parents=True, exist_ok=False)
    scratch = Path(tempfile.mkdtemp(prefix="pbj-release-verify-", dir="/tmp"))
    driver = Driver(scratch, evidence)
    report = {"verdict": "INCONCLUSIVE", "actions": driver.actions,
              "scope": "local source formula installation only; never GitHub publication",
              "unexercised": ["Cog version bump/changelog/tag and exact-tag author-free notes",
                              "GitHub release and optional binary asset publication",
                              "remote Homebrew tap publication and scoped secret availability",
                              "macOS 13 runtime and non-native architecture"]}
    passed = False
    try:
        report["host"] = driver.doctor()
        if not args.doctor:
            report["verdict"] = "NOT VERIFIED"
            driver.drive(args.source.resolve(), report["host"])
        passed = True
    except Exception as error:
        report["error"] = redact(error, scratch, evidence)
    finally:
        report["cleanup"] = driver.cleanup()
        if passed and not report["cleanup"]["errors"]:
            report["verdict"] = "VERIFIED" if not args.doctor else "INCONCLUSIVE"
        if evidence is not None:
            (evidence / "report.json").write_text(json.dumps(report, indent=2) + "\n")
    if args.doctor:
        print(json.dumps({"doctor": "ready" if passed and not report["cleanup"]["errors"] else "not ready",
                          "host": report.get("host"), "error": report.get("error"),
                          "cleanup": report["cleanup"]}, indent=2))
    else:
        print(json.dumps({"verdict": report["verdict"], "evidence": "$EVIDENCE/report.json",
                          "cleanup": report["cleanup"]}, indent=2))
    return 0 if passed and not report["cleanup"]["errors"] else 1


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as error:
        print(json.dumps({"verdict": "INCONCLUSIVE", "error": redact(error)}))
        raise SystemExit(1)
