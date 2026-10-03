"""Two independent network namespaces: actual overlay TCP and private-group rejection.

Called only inside the disposable Linux Docker gate, never against live nodes.
"""
import hashlib
import json
import socket
import subprocess
import time

COMPOSE = ["docker", "compose", "-f", "contrib/umbrel/compose.dev.yml"]
PAYLOAD_SIZE = 1024 * 1024
EXPECTED = hashlib.sha256(b"u" * PAYLOAD_SIZE).hexdigest()


def execute(service, code):
    return subprocess.check_output([*COMPOSE, "exec", "-T", service, "python3", "-c", code],
                                   text=True, timeout=40)


def control(service, action, settings=None):
    request = {"action": action}
    if settings is not None:
        request["settings"] = settings
    value = json.loads(execute(service, "from control import socket_json; import json; "
                              "print(json.dumps(socket_json('/run/uqda-control/control.sock', " +
                              repr(request) + ", timeout=32)))"))
    assert value.get("ok"), value.get("error", "Control operation failed")
    return value["result"]


def wait_for(check):
    deadline = time.monotonic() + 45
    while time.monotonic() < deadline:
        try:
            result = check()
            if result:
                return result
        except (OSError, ValueError, subprocess.SubprocessError, AssertionError):
            pass
        time.sleep(1)
    raise AssertionError("Two-node network did not become ready")


def configure(service, password, peers, listeners):
    state = control(service, "status")
    return control(service, "apply", {"revision": state["settings"]["revision"],
                   "peers": peers, "listen": listeners, "mode": "private", "group_password": password})


def probe(sender, receiver, address, allowed=True):
    server_code = """
import hashlib, json, os, socket
with socket.socket(socket.AF_INET6, socket.SOCK_STREAM) as server:
    server.settimeout(35)
    server.bind((ADDRESS, 0))
    server.listen(1)
    print(json.dumps({'port': server.getsockname()[1], 'pid': os.getpid()}), flush=True)
    connection, _ = server.accept()
    with connection:
        connection.settimeout(15)
        digest, total = hashlib.sha256(), 0
        while True:
            block = connection.recv(65536)
            if not block: break
            digest.update(block)
            total += len(block)
        assert total == 1048576
        print(digest.hexdigest(), flush=True)
""".replace("ADDRESS", repr(address))
    process = subprocess.Popen([*COMPOSE, "exec", "-T", receiver, "python3", "-u", "-c", server_code],
                               stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    pid = None
    try:
        ready = json.loads(process.stdout.readline())
        pid = ready["pid"]
        client = """
import socket
with socket.socket(socket.AF_INET6, socket.SOCK_STREAM) as connection:
    connection.settimeout(8)
    try:
        connection.connect((ADDRESS, PORT))
    except (TimeoutError, OSError):
        assert not ALLOWED, 'Expected reachable overlay service'
    else:
        assert ALLOWED, 'Different private-group password allowed overlay TCP'
        connection.sendall(b'u' * 1048576)
        connection.shutdown(socket.SHUT_WR)
""".replace("ADDRESS", repr(address)).replace("PORT", str(ready["port"])).replace("ALLOWED", repr(allowed))
        execute(sender, client)
        if allowed:
            output, error = process.communicate(timeout=20)
            assert process.returncode == 0, error
            assert output.strip() == EXPECTED, "Overlay payload hash mismatch"
    finally:
        if pid is not None and process.poll() is None:
            execute(receiver, "import os, signal\ntry: os.kill(" + str(pid) + ", signal.SIGTERM)\nexcept ProcessLookupError: pass")
        if process.poll() is None:
            process.communicate(timeout=10)


def probe_udp(sender, receiver, address):
    code = """
import hashlib, socket
with socket.socket(socket.AF_INET6, socket.SOCK_DGRAM) as server:
    server.settimeout(12)
    server.bind((ADDRESS, 0))
    print(server.getsockname()[1], flush=True)
    data, remote = server.recvfrom(2048)
    assert data == b'd' * 1024
    server.sendto(hashlib.sha256(data).digest(), remote)
""".replace("ADDRESS", repr(address))
    server = subprocess.Popen([*COMPOSE, "exec", "-T", receiver, "python3", "-u", "-c", code],
                              stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    try:
        port = int(server.stdout.readline())
        execute(sender, "import socket, hashlib\nwith socket.socket(socket.AF_INET6, socket.SOCK_DGRAM) as client:\n"
                " client.settimeout(8)\n client.connect((" + repr(address) + ", " + str(port) + "))\n"
                " client.send(b'd' * 1024)\n assert client.recv(2048) == hashlib.sha256(b'd' * 1024).digest()")
    finally:
        _, error = server.communicate(timeout=15)
    assert server.returncode == 0, error


def run_network_test(compose):
    secret = "disposable-network-test-secret-0123456789"
    core_before = control("core", "status")["identity"]["address"]
    # The peer uses a bridge network, its own TUN and volatile config. It neither
    # copies the host identity nor installs services on any persistent server.
    compose("up", "-d", "peer")
    try:
        wait_for(lambda: control("peer", "status").get("ready"))
        with socket.socket() as reservation:
            reservation.bind(("0.0.0.0", 0))
            port = reservation.getsockname()[1]
        listeners = ["tls://0.0.0.0:" + str(port)]
        peers = ["tls://host.docker.internal:" + str(port)]
        configure("core", secret, [], listeners)
        peer = configure("peer", secret, peers, [])
        address = peer["identity"]["address"]
        assert address != core_before, "Test peer must have its own identity"
        wait_for(lambda: any(p["up"] for p in control("core", "status")["peers"]))
        probe("core", "peer", address)
        probe("peer", "core", core_before)
        probe_udp("core", "peer", address)
        probe_udp("peer", "core", core_before)
        configure("peer", "different-private-group-secret-0123456789", peers, [])
        wait_for(lambda: any(p["up"] for p in control("peer", "status")["peers"]))
        probe("core", "peer", address, allowed=False)
        configure("peer", secret, peers, [])
        wait_for(lambda: any(p["up"] for p in control("peer", "status")["peers"]))
        probe("core", "peer", address)
        assert control("core", "status")["identity"]["address"] == core_before
        print("PASS: two independent nodes, bidirectional IPv6 TCP/hash and UDP, wrong-group rejection and recovery")
    finally:
        compose("rm", "-s", "-f", "peer")
        configure("core", secret, [], [])
