"""Portable policy tests. Real host TUN and remote HTTP live in the Docker gate."""
import hashlib
import json
from pathlib import Path
import socket
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "contrib" / "umbrel"))
from service_access import (ServiceBook, ServiceError, access_details, node_address,
                            probe_umbrel, umbrel_access, validate_service)

ADDRESS = "200:1234::abcd"


class ServiceTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory(prefix="uqda-service-test-")
        self.addCleanup(self.directory.cleanup)
        self.book = ServiceBook(self.directory.name, lambda path, data: path.write_bytes(data))

    def add(self, **changes):
        values = {"revision": self.book.snapshot()["revision"], "name": "My files", "kind": "https", "port": 8443}
        values.update(changes)
        return self.book.change(values)

    def test_persistent_metadata_and_removal_only_touch_address_book(self):
        config = Path(self.directory.name) / "uqda.conf"
        config.write_bytes(b"unchanged private config")
        state = self.add(name="ملفاتي")
        other = ServiceBook(self.directory.name, self.book.writer)
        self.assertEqual(other.snapshot(), state)
        self.assertEqual(state["items"][0]["name"], "ملفاتي")
        self.assertNotEqual(state["revision"], hashlib.sha256(b"").hexdigest())
        other.change({"revision": state["revision"], "id": state["items"][0]["id"]}, remove=True)
        self.assertEqual(other.snapshot()["items"], [])
        self.assertEqual(config.read_bytes(), b"unchanged private config")

    def test_reject_stale_writes_and_duplicate_ports(self):
        old = self.book.snapshot()["revision"]
        state = self.add()
        before = self.book.path.read_bytes()
        for values in ({"revision": old, "name": "stale", "kind": "http", "port": 8000},
                       {"revision": state["revision"], "name": "duplicate", "kind": "https", "port": 8443}):
            with self.assertRaises(ServiceError):
                self.book.change(values)
        with self.assertRaises(ServiceError):
            self.book.change({"revision": state["revision"], "id": "not-found"}, remove=True)
        self.assertEqual(before, self.book.path.read_bytes())

    def test_limit_twelve_services(self):
        for port in range(8000, 8012):
            self.add(port=port)
        with self.assertRaises(ServiceError):
            self.add(port=9000)

    def test_invalid_fields_ports_and_names(self):
        base = {"name": "files", "kind": "https", "port": 443}
        for changes in ({"name": " "}, {"name": "a" * 65}, {"name": "bad\nname"},
                        {"name": "\ud800"}, {"port": True}, {"port": 0}, {"port": 65536},
                        {"port": "443"}, {"port": 443.0}, {"kind": "javascript"}, {"kind": []},
                        {"host": "127.0.0.1"}, {"command": "id"}, {"password": "secret"}):
            with self.subTest(changes=changes), self.assertRaises(ServiceError):
                validate_service({**base, **changes})

    def test_only_canonical_uqda_identity_enters_recipes(self):
        service = validate_service({"name": "<img onerror=alert(1)>", "kind": "https", "port": 443})
        details = access_details(service, "0200:1234:0:0:0:0:0:abcd")
        self.assertEqual(details["endpoint"], f"https://[{ADDRESS}]:443/")
        self.assertNotIn(service["name"], details["service_command"])
        self.assertNotIn("--insecure", details["service_command"])
        for address in ("::1", "127.0.0.1", "fe80::1%eth0", "200::1%eth0", "host.example", "200::1;id", None):
            with self.subTest(address=address), self.assertRaises(ServiceError):
                node_address(address)

    def test_each_service_type_has_correct_recipe(self):
        for kind in ("http", "https", "ssh", "tcp"):
            service = validate_service({"name": "service", "kind": kind, "port": 1234})
            details = access_details(service, ADDRESS)
            self.assertEqual(details["network_command"], f"sudo uqda test {ADDRESS}")
            expected = {"http": "curl ", "https": "curl ", "ssh": "ssh -p 1234 USER@", "tcp": "nc -6 "}[kind]
            self.assertTrue(details["service_command"].startswith(expected))

    def test_probe_only_saved_port_on_own_identity_no_payload(self):
        item = self.add()["items"][0]
        with patch("service_access.socket.socket") as factory:
            connection = factory.return_value.__enter__.return_value
            result = self.book.probe({"id": item["id"]}, ADDRESS)
            factory.assert_called_once_with(socket.AF_INET6, socket.SOCK_STREAM)
            connection.settimeout.assert_called_once_with(2)
            connection.connect.assert_called_once_with((ADDRESS, 8443))
            connection.send.assert_not_called()
            connection.sendall.assert_not_called()
            self.assertTrue(result["tcp_reachable"])
            self.assertFalse(result["remote_verified"])
            self.assertEqual(result["scope"], "local")

    def test_failed_probe_is_not_remote_success(self):
        item = self.add()["items"][0]
        with patch("service_access.socket.socket") as factory:
            factory.return_value.__enter__.return_value.connect.side_effect = OSError("private failure details")
            result = self.book.probe({"id": item["id"]}, ADDRESS)
        self.assertFalse(result["tcp_reachable"])
        self.assertFalse(result["remote_verified"])
        self.assertNotIn("private failure details", json.dumps(result))

    def test_probe_cannot_supply_hosts_ports_shell_or_unknown_ids(self):
        item = self.add()["items"][0]
        for values in ({"id": item["id"], "host": "127.0.0.1"}, {"id": item["id"], "port": 22},
                       {"id": "unknown"}, {"command": "id"}, None):
            with patch("service_access.socket.socket") as factory, self.assertRaises(ServiceError):
                self.book.probe(values, ADDRESS)
            factory.assert_not_called()

    def test_corrupt_or_oversized_store_rejected(self):
        for raw in (b"not-json", b"{}", b"[{}]", b" " * 16385):
            self.book.path.write_bytes(raw)
            with self.assertRaises(ServiceError):
                self.book.snapshot()
        item = {"id": "0" * 16, "name": "ssh", "kind": "ssh", "port": 22}
        self.book.path.write_text(json.dumps([item, item]))
        with self.assertRaises(ServiceError):
            self.book.snapshot()

    def test_umbrel_addresses_require_private_group_and_tun(self):
        for private, tun in ((False, True), (False, False), (True, False)):
            access = umbrel_access(ADDRESS, private, tun)
            self.assertFalse(access["enabled"])
            self.assertNotIn("http_url", access)
            self.assertFalse(access["remote_verified"])
        self.assertFalse(umbrel_access(None, True, True)["enabled"])
        access = umbrel_access(ADDRESS, True, True)
        self.assertEqual(access["http_url"], f"http://[{ADDRESS}]/")
        self.assertEqual(access["https_url"], f"https://[{ADDRESS}]/")

    def test_umbrel_probe_only_fixed_public_ports_not_internal_server(self):
        with patch("service_access.local_tcp", return_value=True) as connect:
            result = probe_umbrel({}, ADDRESS, True, True)
        self.assertEqual([call.args for call in connect.call_args_list], [(ADDRESS, 80), (ADDRESS, 443), (ADDRESS, 2000)])
        self.assertEqual(result["scope"], "local")
        self.assertFalse(result["remote_verified"])
        for values, private, tun in (({"host": "127.0.0.1"}, True, True), ({"port": 22080}, True, True),
                                     ({}, False, True), ({}, True, False), (None, True, True)):
            with patch("service_access.local_tcp") as connect, self.assertRaises(ServiceError):
                probe_umbrel(values, ADDRESS, private, tun)
            connect.assert_not_called()


if __name__ == "__main__":
    unittest.main()
