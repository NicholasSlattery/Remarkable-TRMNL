#!/usr/bin/env python3
import json
import os
import signal
import socket
import struct
import subprocess
import sys
import tempfile
import time
from pathlib import Path

MSG_INIT = 1
MSG_SAVE = 2
MSG_NEXT = 5
MSG_STATE = 101
MSG_IMAGE = 102
MSG_STATUS = 103
SYSTEM_TERMINATE = 0xFFFFFFFF


def send(conn, kind, payload=""):
    data = payload.encode()
    conn.send(struct.pack("<II", kind, len(data)))
    # AppLoad's C++ sender emits a second sequence packet even for an empty
    # QString. The production backend must consume it.
    conn.send(data)


def receive(conn, timeout=15):
    conn.settimeout(timeout)
    header = conn.recv(8)
    if len(header) != 8:
        raise RuntimeError(f"short header: {len(header)}")
    kind, length = struct.unpack("<II", header)
    payload = conn.recv(length).decode() if length else ""
    return kind, json.loads(payload or "{}")


def wait_for(conn, expected, timeout=30):
    deadline = time.monotonic() + timeout
    seen = []
    while time.monotonic() < deadline:
        kind, payload = receive(conn, max(1, deadline - time.monotonic()))
        seen.append((kind, payload))
        if kind == expected:
            return payload, seen
    raise TimeoutError(f"message {expected} not received; saw {seen}")


def main():
    if len(sys.argv) != 3:
        raise SystemExit("usage: appload_harness.py BACKEND MOCK_SERVER")
    backend, mock_bin = map(Path, sys.argv[1:])
    with tempfile.TemporaryDirectory(prefix="trmnl-integration-") as td:
        root = Path(td)
        home = root / "home"
        home.mkdir(mode=0o700)
        sock_path = str(root / "appload.sock")
        listener = socket.socket(socket.AF_UNIX, socket.SOCK_SEQPACKET)
        listener.bind(sock_path)
        listener.listen(1)
        mock = subprocess.Popen([str(mock_bin), "-listen", "127.0.0.1:19988"], stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)
        try:
            for _ in range(50):
                try:
                    with socket.create_connection(("127.0.0.1", 19988), timeout=.1):
                        break
                except OSError:
                    time.sleep(.1)
            else:
                raise RuntimeError("mock server did not start")
            env = dict(os.environ, HOME=str(home))
            proc = subprocess.Popen([str(backend), sock_path], env=env)
            conn, _ = listener.accept()
            try:
                send(conn, MSG_INIT)
                state, _ = wait_for(conn, MSG_STATE)
                assert state["api_key_configured"] is False
                config = state["config"]
                config.update({"api_key": "local-test", "base_url": "http://127.0.0.1:19988", "device_id": "", "minimum_refresh_seconds": 60})
                send(conn, MSG_SAVE, json.dumps(config))
                wait_for(conn, MSG_STATE)
                send(conn, MSG_NEXT)
                image, seen = wait_for(conn, MSG_IMAGE)
                image_path = Path(image["path"].removeprefix("file://"))
                assert image_path.is_file() and image_path.stat().st_size > 1000
                assert image_path.suffix == ".png"
                cfg_path = home / ".config/trmnl-remarkable/config.json"
                assert cfg_path.stat().st_mode & 0o777 == 0o600
                saved = json.loads(cfg_path.read_text())
                assert saved["api_key"] == "local-test"
                send(conn, SYSTEM_TERMINATE)
                proc.wait(timeout=10)
                assert proc.returncode == 0
                print(json.dumps({"ok": True, "image": str(image_path), "bytes": image_path.stat().st_size, "messages_seen": len(seen)}))
            finally:
                conn.close()
                if proc.poll() is None:
                    proc.send_signal(signal.SIGTERM)
                    proc.wait(timeout=10)
        finally:
            mock.terminate()
            mock.wait(timeout=10)
            listener.close()


if __name__ == "__main__":
    main()
