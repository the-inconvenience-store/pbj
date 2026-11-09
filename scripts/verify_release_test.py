"""Regression coverage for release-verifier process ownership; no Homebrew writes."""

import contextlib
import json
import os
from pathlib import Path
import runpy
import signal
import sys
import tempfile
import unittest


Driver = runpy.run_path(str(Path(__file__).with_name("verify-release.py")))["Driver"]

# Filling stdout before interrupting requires Driver.run to have started draining
# its pipe. No sleeps or timing assumptions are needed to reach communicate().
CHILD = r"""
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import time

grandchild = subprocess.Popen([sys.executable, "-c", "import time; time.sleep(60)"])
def terminate(signum, frame):
    grandchild.wait(timeout=5)
    raise SystemExit(0)
signal.signal(signal.SIGTERM, terminate)
Path(sys.argv[1]).write_text(json.dumps({"pid": os.getpid()}))
sys.stdout.buffer.write(b"!" * (2 * 1024 * 1024))
sys.stdout.buffer.flush()
os.kill(os.getppid(), signal.SIGINT)
time.sleep(60)
"""


class ProcessOwnershipTest(unittest.TestCase):
    def test_interrupt_reaps_the_owned_process_group_before_cleanup(self):
        with tempfile.TemporaryDirectory(prefix="pbj-interrupt-", dir="/tmp") as directory:
            root = Path(directory)
            scratch = root / "scratch"
            scratch.mkdir()
            driver = Driver(scratch)
            marker = root / "child.json"
            try:
                with self.assertRaises(KeyboardInterrupt):
                    driver.run([sys.executable, "-c", CHILD, marker], timeout=20)
                pid = json.loads(marker.read_text())["pid"]
                with self.assertRaises(ChildProcessError, msg="owned child was not reaped"):
                    os.waitpid(pid, os.WNOHANG)
                with self.assertRaises(ProcessLookupError, msg="owned process group is still alive"):
                    os.killpg(pid, 0)
            finally:
                # Red runs must not leave their deliberately interrupted children.
                if marker.exists():
                    pid = json.loads(marker.read_text())["pid"]
                    with contextlib.suppress(ProcessLookupError):
                        os.killpg(pid, signal.SIGKILL)
                    with contextlib.suppress(ChildProcessError):
                        os.waitpid(pid, 0)
                cleanup = driver.cleanup()
                self.assertEqual(cleanup["errors"], [])


if __name__ == "__main__":
    unittest.main()
