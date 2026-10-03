"""Native Linux Docker gate: actual host TUN, split service permissions and identity."""
import http.client
import json
import os
from pathlib import Path
import subprocess
import time

COMPOSE = ["docker", "compose", "-f", "contrib/umbrel/compose.dev.yml"]
PASSWORD = os.environ["UQDA_TEST_PASSWORD"]
cookie = ""
csrf = ""


def compose(*args):
    return subprocess.check_output([*COMPOSE, *args], text=True)


def request(path, value=None):
    connection = http.client.HTTPConnection("127.0.0.1", 8926, timeout=35)
    headers = {"Cookie": cookie}
    if value is not None:
        headers.update({"Content-Type": "application/json", "Origin": "http://127.0.0.1:8926", "X-Uqda-CSRF": csrf})
    connection.request("GET" if value is None else "POST", path, None if value is None else json.dumps(value), headers)
    response = connection.getresponse()
    body = json.loads(response.read())
    status, response_headers = response.status, dict(response.getheaders())
    connection.close()
    return status, body, response_headers


def ready():
    deadline = time.monotonic() + 60
    while time.monotonic() < deadline:
        try:
            status, body, _ = request("/api/status")
            if status == 200 and body.get("ready"):
                return body
        except (OSError, ValueError):
            pass
        time.sleep(1)
    raise AssertionError("Core/dashboard did not become ready")


try:
    compose("up", "-d", "--build")
    for _ in range(30):
        try:
            status, body, headers = request("/api/login", {"password": PASSWORD})
            if status == 200:
                cookie, csrf = headers["Set-Cookie"].split(";")[0], body["csrf"]
                break
        except (OSError, ValueError):
            pass
        time.sleep(1)
    else:
        raise AssertionError("Dashboard login failed")
    state = ready()
    assert state["tun"]["enabled"] and state["tun"]["name"] == "uqda0"
    address = state["identity"]["address"]
    if os.environ.get("UQDA_BROWSER_TEST") == "1":
        subprocess.run(["node", "tests/umbrel/browser_smoke.cjs"], check=True)
        state = ready()
    assert compose("exec", "-T", "dashboard", "id", "-u").strip() == "1000"
    assert compose("exec", "-T", "dashboard", "python3", "-c",
                   "from pathlib import Path; assert not Path('/etc/uqda/uqda.conf').exists(); assert not Path('/run/uqda-core/admin.sock').exists(); print('isolated')").strip() == "isolated"
    assert compose("exec", "-T", "core", "stat", "-c", "%a", "/etc/uqda/uqda.conf").strip() == "600"
    config = {"revision": state["settings"]["revision"], "peers": [], "listen": [], "mode": "private",
              "group_password": "docker-private-group-test-0123456789"}
    assert request("/api/settings", config)[0] == 200
    compose("restart", "core")
    assert ready()["identity"]["address"] == address
    # Recreate the containers as an image update would, retaining the same data.
    compose("up", "-d", "--force-recreate")
    # Sessions are deliberately invalidated by dashboard recreation.
    for _ in range(30):
        try:
            status, body, headers = request("/api/login", {"password": PASSWORD})
            if status == 200:
                cookie, csrf = headers["Set-Cookie"].split(";")[0], body["csrf"]
                break
        except (OSError, ValueError):
            pass
        time.sleep(1)
    else:
        raise AssertionError("Dashboard re-login failed")
    assert ready()["identity"]["address"] == address
    print("PASS: host TUN, login, split permissions, settings, restart and container recreation")
finally:
    compose("logs", "--no-color")
    compose("down", "--volumes")
