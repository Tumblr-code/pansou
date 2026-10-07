#!/usr/bin/env python3
"""Exercise only local health and graceful shutdown; never run a provider search."""
import argparse
import json
import os
import signal
import socket
import subprocess
import tempfile
import time
import urllib.error
import urllib.request
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("binary", type=Path)
args = parser.parse_args()
with socket.socket() as probe:
    probe.bind(("127.0.0.1", 0))
    port = probe.getsockname()[1]
with tempfile.TemporaryDirectory(prefix="pansou-smoke-") as directory:
    env = dict(os.environ, PORT=str(port), CACHE_PATH=directory, AUTH_ENABLED="false",
               ASYNC_PLUGIN_ENABLED="false", CHANNELS="offline-placeholder")
    with tempfile.TemporaryFile() as log:
        process = subprocess.Popen([str(args.binary.resolve())], cwd=directory, env=env,
                                   stdout=log, stderr=log)
        try:
            deadline = time.monotonic() + 20
            while True:
                if process.poll() is not None:
                    raise RuntimeError("backend exited before local health was ready")
                try:
                    with urllib.request.urlopen(
                        f"http://127.0.0.1:{port}/api/health", timeout=1
                    ) as response:
                        assert json.load(response)["status"] == "ok"
                    break
                except (OSError, urllib.error.URLError):
                    if time.monotonic() >= deadline:
                        raise RuntimeError("local health deadline exceeded") from None
                    time.sleep(0.1)
            process.send_signal(signal.SIGTERM)
            assert process.wait(timeout=15) == 0
        finally:
            if process.poll() is None:
                process.kill()
                process.wait()
print("Isolated local health and shutdown passed; no provider search executed")
