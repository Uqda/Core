"""A service address book, not a publisher, proxy or arbitrary network scanner."""
import hashlib
import ipaddress
import json
from pathlib import Path
import re
import secrets
import socket


class ServiceError(Exception):
    pass


def validate_service(value, stored=False):
    fields = {"name", "kind", "port"} | ({"id"} if stored else set())
    if not isinstance(value, dict) or set(value) != fields:
        raise ServiceError("Use a service name, type and TCP port only.")
    name = value["name"]
    if not isinstance(name, str) or not 1 <= len(name.strip()) <= 64 or any(ord(c) < 32 or ord(c) == 127 or 0xD800 <= ord(c) <= 0xDFFF for c in name):
        raise ServiceError("Use a service name of 1 to 64 characters without control characters.")
    if not isinstance(value["kind"], str) or value["kind"] not in {"https", "http", "ssh", "tcp"}:
        raise ServiceError("Select HTTPS, HTTP, SSH or TCP.")
    port = value["port"]
    if type(port) is not int or not 1 <= port <= 65535:
        raise ServiceError("Use a TCP port from 1 to 65535.")
    result = {"name": name.strip(), "kind": value["kind"], "port": port}
    if stored:
        if not isinstance(value["id"], str) or not re.fullmatch(r"[0-9a-f]{16}", value["id"]):
            raise ServiceError("Invalid saved service identifier.")
        result["id"] = value["id"]
    return result


def node_address(address):
    try:
        value = ipaddress.IPv6Address(address)
        if value.scope_id is not None or value not in ipaddress.IPv6Network("200::/7"):
            raise ValueError()
        return str(value)
    except (ValueError, TypeError):
        raise ServiceError("A ready Uqda IPv6 identity is required.") from None


def access_details(service, address):
    """Only daemon-derived literal addresses and validated ports enter commands."""
    address = node_address(address)
    port, kind = service["port"], service["kind"]
    endpoint = f"[{address}]:{port}"
    command = f"nc -6 -vz -w 5 {address} {port}"
    if kind in {"https", "http"}:
        endpoint = f"{kind}://{endpoint}/"
        command = f"curl --connect-timeout 5 --max-time 10 --head '{endpoint}'"
    elif kind == "ssh":
        command = f"ssh -p {port} USER@{address}"
    return {"endpoint": endpoint, "network_command": f"sudo uqda test {address}",
            "service_command": command}


class ServiceBook:
    def __init__(self, directory, writer):
        self.path = Path(directory) / "services.json"
        self.writer = writer

    def snapshot(self):
        if self.path.is_symlink():
            raise ServiceError("Service list must be a regular file.")
        raw = b""
        if self.path.exists():
            if not self.path.is_file() or self.path.stat().st_size > 16384:
                raise ServiceError("Invalid saved service list.")
            raw = self.path.read_bytes()
        try:
            values = json.loads(raw) if raw else []
            if not isinstance(values, list) or len(values) > 12:
                raise ServiceError("Use at most 12 services.")
            items = [validate_service(item, stored=True) for item in values]
            if len({item["id"] for item in items}) != len(items):
                raise ServiceError("Duplicate saved service identifier.")
        except (ValueError, UnicodeError):
            raise ServiceError("Invalid saved service list.") from None
        return {"revision": hashlib.sha256(raw).hexdigest(), "items": items}

    def change(self, values, remove=False):
        expected = {"revision", "id"} if remove else {"revision", "name", "kind", "port"}
        if not isinstance(values, dict) or set(values) != expected:
            raise ServiceError("Invalid service request.")
        state = self.snapshot()
        if values["revision"] != state["revision"]:
            raise ServiceError("Service list changed. Refresh before saving again.")
        if remove:
            items = [item for item in state["items"] if item["id"] != values["id"]]
            if len(items) == len(state["items"]):
                raise ServiceError("Service not found.")
        else:
            if len(state["items"]) >= 12:
                raise ServiceError("Use at most 12 services.")
            item = validate_service({key: values[key] for key in ("name", "kind", "port")})
            if any(saved["port"] == item["port"] and saved["kind"] == item["kind"] for saved in state["items"]):
                raise ServiceError("That service type and port are already saved.")
            items = [*state["items"], {**item, "id": secrets.token_hex(8)}]
        self.writer(self.path, (json.dumps(items, ensure_ascii=True, indent=2) + "\n").encode())
        return self.snapshot()

    def probe(self, values, address):
        if not isinstance(values, dict) or set(values) != {"id"}:
            raise ServiceError("Select one saved service. Custom probe targets are not allowed.")
        service = next((item for item in self.snapshot()["items"] if item["id"] == values["id"]), None)
        if service is None:
            raise ServiceError("Service not found.")
        # No DNS, user-supplied hosts, HTTP requests, credentials or forwarding.
        # Only a saved TCP port on THIS daemon's own overlay identity is tested.
        address = node_address(address)
        reachable = False
        with socket.socket(socket.AF_INET6, socket.SOCK_STREAM) as connection:
            connection.settimeout(2)
            try:
                connection.connect((address, service["port"]))
                reachable = True
            except OSError:
                pass
        return {"id": service["id"], "tcp_reachable": reachable, "scope": "local",
                "remote_verified": False, **access_details(service, address)}
