"""Local-only Uqda supervisor and deliberately small dashboard control API."""
import fcntl
import hashlib
import json
import os
from pathlib import Path
import secrets
import signal
import socket
import socketserver
import subprocess
import tempfile
import threading
import time
from urllib.parse import parse_qsl, urlsplit
from service_access import ServiceBook, ServiceError, access_details

MAX_MESSAGE = 65536


class ControlError(Exception):
    pass


def atomic_write(path, data):
    path = Path(path)
    fd, temporary = tempfile.mkstemp(prefix=".uqda-", dir=path.parent)
    try:
        os.fchmod(fd, 0o600)
        with os.fdopen(fd, "wb") as stream:
            stream.write(data)
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(temporary, path)
        directory = os.open(path.parent, os.O_RDONLY)
        try:
            os.fsync(directory)
        finally:
            os.close(directory)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


def validate_uris(values, listener=False):
    if not isinstance(values, list) or len(values) > 32:
        raise ControlError("Use at most 32 peer or listener addresses.")
    result = []
    for value in values:
        if not isinstance(value, str) or not 1 <= len(value) <= 1024:
            raise ControlError("Invalid address length.")
        if any(ord(c) < 33 or ord(c) > 126 for c in value):
            raise ControlError("Addresses must not contain whitespace or control characters.")
        try:
            uri = urlsplit(value)
            schemes = {"tls", "tcp", "quic"} if listener else {"tls", "tcp", "quic", "ws", "wss"}
            if uri.scheme not in schemes or not uri.hostname or not uri.port:
                raise ValueError()
            if uri.username is not None or uri.password is not None or uri.fragment:
                raise ValueError()
            if listener and (uri.query or uri.path not in ("", "/") or uri.port < 1024):
                raise ValueError()
            # Password-bearing transport URIs belong in advanced local config, never this UI.
            # Decode parameter names just as Core's URL parser does. Encoded
            # password keys must never make credential-bearing peers editable.
            parameters = parse_qsl(uri.query, keep_blank_values=True, strict_parsing=True, max_num_fields=16)
            if any(key not in {"key", "priority", "maxbackoff", "sni", "origin"} for key, _ in parameters):
                raise ValueError()
        except ValueError:
            raise ControlError("Use a supported URI with a host and port, without embedded credentials.") from None
        if value not in result:
            result.append(value)
    return result


def display_uri(value):
    try:
        uri = urlsplit(value)
        host = uri.hostname or ""
        if ":" in host:
            host = "[" + host + "]"
        return uri.scheme + "://" + host + (":" + str(uri.port) if uri.port else "")
    except (ValueError, TypeError):
        return "Configured peer"


def socket_json(path, request, timeout=4):
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as connection:
        connection.settimeout(timeout)
        connection.connect(str(path))
        connection.sendall(json.dumps(request).encode() + b"\n")
        connection.shutdown(socket.SHUT_WR)
        data = bytearray()
        while True:
            block = connection.recv(4096)
            if not block:
                break
            data.extend(block)
            if len(data) > MAX_MESSAGE:
                raise ControlError("Response too large.")
        return json.loads(data)


class Supervisor:
    def __init__(self, config_dir, runtime_dir, binary="/usr/bin/uqda", tun_name="uqda0"):
        self.config_dir = Path(config_dir)
        self.runtime_dir = Path(runtime_dir)
        self.config_path = self.config_dir / "uqda.conf"
        self.admin_path = self.runtime_dir / "admin.sock"
        self.binary = binary
        self.tun_name = tun_name
        self.process = None
        self.lock = threading.RLock()
        self.closing = False
        self.services = ServiceBook(self.config_dir, atomic_write)

    def command(self, *args):
        try:
            result = subprocess.run([self.binary, *args], capture_output=True, timeout=12, check=False)
        except (OSError, subprocess.TimeoutExpired):
            raise ControlError("Core could not validate the configuration.") from None
        if result.returncode:
            # Core validation errors can contain private key or peer material.
            raise ControlError("Core rejected the configuration. No settings were applied.")
        return result.stdout

    def validate(self, config):
        fd, candidate = tempfile.mkstemp(prefix=".validate-", dir=self.config_dir)
        try:
            os.fchmod(fd, 0o600)
            with os.fdopen(fd, "w") as stream:
                json.dump(config, stream)
            self.command("-useconffile", candidate, "-checkconf")
        finally:
            os.unlink(candidate)

    def initialize(self):
        self.config_dir.mkdir(parents=True, exist_ok=True, mode=0o700)
        os.chmod(self.config_dir, 0o700)
        self.lock_file = open(self.config_dir / ".supervisor.lock", "a")
        try:
            fcntl.flock(self.lock_file, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            raise ControlError("Another supervisor is using this node configuration.") from None
        self.runtime_dir.mkdir(parents=True, exist_ok=True, mode=0o700)
        os.chmod(self.runtime_dir, 0o700)
        if self.config_path.is_symlink():
            raise ControlError("Configuration must be a regular file.")
        if self.config_path.exists():
            original = self.config_path.read_bytes()
            # Core performs HJSON parsing and retains externally stored identity references.
            config = json.loads(self.command("-useconffile", str(self.config_path), "-normaliseconf", "-json"))
            if not (self.config_dir / "uqda.conf.before-umbrel").exists():
                atomic_write(self.config_dir / "uqda.conf.before-umbrel", original)
        else:
            config = json.loads(self.command("-genconf", "-json"))
            config.update(Peers=[], Listen=[], MulticastInterfaces=[],
                          GroupPassword=secrets.token_urlsafe(32), NodeInfoPrivacy=True,
                          IfName=self.tun_name, IfMTU=1280)
        # The raw admin API is never placed in the dashboard's shared directory.
        config["AdminListen"] = "unix://" + str(self.admin_path)
        self.validate(config)
        atomic_write(self.config_path, (json.dumps(config, indent=2) + "\n").encode())

    def read_config(self):
        return json.loads(self.config_path.read_bytes())

    def revision(self):
        return hashlib.sha256(self.config_path.read_bytes()).hexdigest()

    def settings(self):
        config = self.read_config()
        peers = config.get("Peers", [])
        # Preserve unsupported advanced peers without returning credentials to the browser.
        editable = True
        try:
            validate_uris(peers)
            validate_uris(config.get("Listen", []), listener=True)
        except ControlError:
            editable = False
        return {"revision": self.revision(), "peers": peers if editable else [display_uri(p) for p in peers],
                "listen": config.get("Listen", []) if editable else [],
                "private": bool(config.get("GroupPassword")), "editable": editable}

    def admin(self, name):
        result = socket_json(self.admin_path, {"request": name})
        if result.get("status") != "success":
            raise ControlError("Core is not ready.")
        return result["response"]

    def status(self):
        with self.lock:
            response = {"running": self.process is not None and self.process.poll() is None,
                        "settings": self.settings(), "ready": False}
            try:
                identity = self.admin("getSelf")
                peers = self.admin("getPeers").get("peers", [])
                tun = self.admin("getTun")
                response.update(ready=True, identity={key: identity.get(key) for key in
                                ("address", "subnet", "key", "build_name", "build_version")},
                                tun={key: tun.get(key) for key in ("enabled", "name", "mtu")},
                                peers=[{**{key: p.get(key, 0) for key in
                                       ("up", "inbound", "address", "bytes_recvd", "bytes_sent", "uptime")},
                                        "remote": display_uri(p.get("remote", ""))} for p in peers])
            except (OSError, ValueError, KeyError, ControlError):
                response["message"] = "Core is starting or unavailable. Check the app's container logs if this continues."
            response["services"] = self.services.snapshot()
            address = response.get("identity", {}).get("address")
            if address:
                response["services"]["items"] = [{**item, **access_details(item, address)}
                                                 for item in response["services"]["items"]]
            return response

    def service_action(self, action, values):
        with self.lock:
            if action == "service_probe":
                state = self.status()
                if not state["ready"] or not state.get("tun", {}).get("enabled"):
                    raise ControlError("Core and its TUN interface must be ready before testing a service.")
                return self.services.probe(values, state["identity"]["address"])
            self.services.change(values, remove=action == "service_remove")
            return self.status()

    def start(self):
        if self.process is not None and self.process.poll() is None:
            return
        # Keep daemon logs out of HTTP responses; operators can use Umbrel's logs.
        self.process = subprocess.Popen([self.binary, "-useconffile", str(self.config_path)])

    def stop(self):
        if self.process is not None and self.process.poll() is None:
            self.process.terminate()
            try:
                self.process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait(timeout=5)
        self.process = None

    def wait_ready(self):
        deadline = time.monotonic() + 12
        while time.monotonic() < deadline:
            if self.process is None or self.process.poll() is not None:
                return False
            try:
                self.admin("getSelf")
                tun = self.admin("getTun")
                if self.read_config().get("IfName") == "none" or tun.get("enabled"):
                    return True
            except (OSError, ValueError, KeyError, ControlError):
                pass
            time.sleep(0.1)
        return False

    def apply(self, values):
        with self.lock:
            if not isinstance(values, dict) or set(values) - {"revision", "peers", "listen", "mode", "group_password", "confirm_public"}:
                raise ControlError("Unknown setting.")
            if values.get("revision") != self.revision():
                raise ControlError("Settings changed. Refresh before saving again.")
            if not self.settings()["editable"]:
                raise ControlError("This configuration contains advanced addresses. Preserve it and edit locally.")
            config = self.read_config()
            config["Peers"] = validate_uris(values.get("peers"))
            config["Listen"] = validate_uris(values.get("listen"), listener=True)
            mode = values.get("mode")
            if mode not in ("private", "public"):
                raise ControlError("Select private group or public network mode.")
            password = values.get("group_password", "")
            if not isinstance(password, str) or len(password.encode()) > 1024:
                raise ControlError("Invalid group password.")
            if mode == "public":
                if values.get("confirm_public") is not True:
                    raise ControlError("Confirm public mode before removing group protection.")
                config["GroupPassword"] = ""
            elif password:
                if len(password) < 16:
                    raise ControlError("Use a randomly generated group password of at least 16 characters.")
                config["GroupPassword"] = password
            elif not config.get("GroupPassword"):
                raise ControlError("Enter the shared password for your private group.")
            self.validate(config)
            original = self.config_path.read_bytes()
            atomic_write(self.config_dir / "uqda.conf.previous", original)
            self.stop()
            try:
                atomic_write(self.config_path, (json.dumps(config, indent=2) + "\n").encode())
                self.start()
                if not self.wait_ready():
                    raise ControlError("Core could not start with those settings. Previous settings restored.")
            except Exception:
                self.stop()
                atomic_write(self.config_path, original)
                self.start()
                raise ControlError("Core could not apply those settings. Previous settings restored.") from None
            return self.status()

    def restart(self):
        with self.lock:
            self.stop()
            self.start()
            if not self.wait_ready():
                raise ControlError("Core did not become ready after restarting.")
            return self.status()

    def monitor(self):
        while not self.closing:
            time.sleep(2)
            with self.lock:
                if not self.closing and (self.process is None or self.process.poll() is not None):
                    self.start()


class ControlHandler(socketserver.StreamRequestHandler):
    def handle(self):
        self.connection.settimeout(5)
        try:
            data = self.rfile.readline(MAX_MESSAGE + 1)
            if len(data) > MAX_MESSAGE:
                raise ControlError("Request too large.")
            request = json.loads(data)
            if not isinstance(request, dict) or set(request) - {"action", "settings"}:
                raise ControlError("Invalid request.")
            action = request.get("action")
            supervisor = self.server.supervisor
            if action == "status":
                value = supervisor.status()
            elif action == "apply":
                value = supervisor.apply(request.get("settings"))
            elif action == "restart":
                value = supervisor.restart()
            elif action in {"service_add", "service_remove", "service_probe"}:
                value = supervisor.service_action(action, request.get("settings"))
            else:
                raise ControlError("Unsupported action.")
            response = {"ok": True, "result": value}
        except (ControlError, ServiceError) as error:
            response = {"ok": False, "error": str(error)}
        except Exception:
            response = {"ok": False, "error": "The local control service could not complete the request."}
        try:
            self.wfile.write(json.dumps(response).encode() + b"\n")
        except OSError:
            pass


class ControlServer(socketserver.ThreadingUnixStreamServer):
    daemon_threads = True
    request_queue_size = 16


def main():
    os.umask(0o077)
    supervisor = Supervisor(os.environ.get("UQDA_CONFIG_DIR", "/etc/uqda"),
                            os.environ.get("UQDA_RUNTIME_DIR", "/run/uqda-core"))
    supervisor.initialize()
    shared = Path(os.environ.get("UQDA_CONTROL_DIR", "/run/uqda-control"))
    shared.mkdir(parents=True, exist_ok=True)
    gid = int(os.environ.get("UQDA_DASHBOARD_GID", "1000"))
    os.chown(shared, 0, gid)
    os.chmod(shared, 0o750)
    address = shared / "control.sock"
    if address.exists():
        address.unlink()
    server = ControlServer(str(address), ControlHandler)
    server.supervisor = supervisor
    os.chown(address, 0, gid)
    os.chmod(address, 0o660)
    supervisor.start()
    threading.Thread(target=supervisor.monitor, daemon=True).start()

    def shutdown(_signum, _frame):
        supervisor.closing = True
        threading.Thread(target=server.shutdown, daemon=True).start()

    signal.signal(signal.SIGTERM, shutdown)
    signal.signal(signal.SIGINT, shutdown)
    try:
        server.serve_forever(poll_interval=0.2)
    finally:
        supervisor.closing = True
        server.server_close()
        with supervisor.lock:
            supervisor.stop()
        address.unlink(missing_ok=True)


if __name__ == "__main__":
    main()
