"""Exercise every advertised CLI command against two real local daemons.

Run: python tests/cli/commands.py PATH_TO_UQDA PATH_TO_UQDACTL
No privileged TUN interface or live configuration is used.
"""
import json
import os
from pathlib import Path
import socket
import subprocess
import sys
import tempfile
import time

NODE, CONTROL = (str(Path(arg).resolve()) for arg in sys.argv[1:3])


def port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def run(binary, args, expected=0):
    result = subprocess.run([binary, *args], capture_output=True, text=True, encoding="utf-8", timeout=40)
    assert result.returncode == expected, (args, result.returncode, result.stdout, result.stderr)
    return result.stdout


with tempfile.TemporaryDirectory(prefix="uqda-cli-") as temporary:
    root = Path(temporary)
    processes, logs, configs = [], [], []
    try:
        for i in range(2):
            cfg = json.loads(run(NODE, ["-genconf", "-json"]))
            cfg.update(IfName="none", MulticastInterfaces=[], Peers=[],
                       AdminListen=f"tcp://127.0.0.1:{port()}",
                       Listen=[f"tcp://127.0.0.1:{port()}"], GroupPassword=os.getenv("UQDA_CLI_TEST_PASSWORD", "cli-test-group"))
            path = root / f"node-{i}.json"
            path.write_text(json.dumps(cfg), encoding="utf-8")
            os.chmod(path, 0o600)
            log = (root / f"node-{i}.log").open("w", encoding="utf-8")
            logs.append(log)
            processes.append(subprocess.Popen([NODE, "-useconffile", str(path)], stdout=log, stderr=log))
            configs.append(cfg)
        endpoint = configs[0]["AdminListen"]
        flags = [f"--endpoint={endpoint}"]
        for _ in range(60):
            result = subprocess.run([CONTROL, *flags, "info", "--json"], capture_output=True, text=True, encoding="utf-8", timeout=10)
            if result.returncode == 0:
                break
            time.sleep(0.1)
        else:
            raise AssertionError("daemon admin socket did not start")
        remote = json.loads(run(CONTROL, [f"--endpoint={configs[1]['AdminListen']}", "info", "--json"]))
        remote_key = remote["key"]
        peer = configs[1]["Listen"][0]
        assert "Peer added" in run(CONTROL, [*flags, "addPeer", f"uri={peer}"])
        for _ in range(60):
            peers = json.loads(run(CONTROL, [*flags, "peers", "--json"]))["peers"]
            if any(p["up"] for p in peers):
                break
            time.sleep(0.1)
        else:
            raise AssertionError("local nodes did not connect")
        # Transport 'up' precedes routing-tree/session readiness (as in
        # src/core.WaitConnected). Check routing first, then allow convergence.
        for _ in range(60):
            trees = [json.loads(run(CONTROL, [f"--endpoint={c['AdminListen']}", "getTree", "--json"]))
                     for c in configs]
            if all(len(tree["tree"]) > 1 for tree in trees):
                break
            time.sleep(0.1)
        else:
            raise AssertionError("routing trees did not converge")
        time.sleep(3)
        json.loads(run(CONTROL, [*flags, "getNodeInfo", f"key={remote_key}", "--json"]))
        commands = json.loads(run(CONTROL, [*flags, "commands", "--json"]))["list"]
        for entry in commands:
            name = entry["command"]
            if name in ("addpeer", "removepeer"):
                continue
            args = [*flags, name, "--json"]
            if "key" in entry.get("fields", []):
                args.append(f"key={remote_key}")
            json.loads(run(CONTROL, args))
            # Text output must also be usable, not just the JSON path.
            text = run(CONTROL, [arg for arg in args if arg != "--json"])
            assert "DEBUG:" not in text, (name, text)
            print(f"PASS {name}")
        for name in ("status", "doctor", "peers", "info", "commands"):
            text = run(NODE, [name, *flags])
            assert "DEBUG:" not in text, (name, text)
            assert all(len(line) <= 80 for line in text.splitlines()), (name, text)
        run(CONTROL, flags)  # One command defaults to the health dashboard.
        for binary in (NODE, CONTROL):
            run(binary, ["help"])
            run(binary, ["version"])
        for args in (["status", "ignored"], ["peers", "unused=value"], ["unknown"]):
            run(CONTROL, [*flags, *args], expected=2)
        run(CONTROL, [*flags, "test", remote["address"]], expected=1)  # No TUN: actionable failure.
        assert "Peer removed" in run(CONTROL, [*flags, "removePeer", f"uri={peer}"])
        for option in ("address", "subnet", "publickey", "checkconf", "normaliseconf", "exportkey"):
            run(NODE, [f"-{option}"], expected=2)
        print(f"PASS all {len(commands)} advertised admin commands, daily shortcuts, JSON and error paths")
    finally:
        for process in processes:
            process.terminate()
        for process in processes:
            try:
                process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=5)
        for log in logs:
            log.close()
        if sys.exc_info()[0] is not None:
            for path in root.glob("*.log"):
                print(path.name, path.read_text(encoding="utf-8"), file=sys.stderr)
