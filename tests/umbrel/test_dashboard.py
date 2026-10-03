"""HTTP boundaries and configuration lifecycle using the real released daemon.

Set UQDA_TEST_BINARY to a built or checksum-verified release binary.
These tests intentionally disable TUN; the Docker smoke gate tests host TUN.
"""
import http.client
import json
import os
from pathlib import Path
import sys
import tempfile
import threading
import time
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "contrib" / "umbrel"))
from control import (ControlError, ControlHandler, ControlServer, Supervisor,
                     atomic_write, display_uri, socket_json, validate_uris)
from dashboard import Dashboard

BINARY = os.environ.get("UQDA_TEST_BINARY", str(Path(__file__).resolve().parents[2] / "uqda"))
PASSWORD = "test-dashboard-password-123456"


class LifecycleTests(unittest.TestCase):
    def setUp(self):
        if not Path(BINARY).is_file():
            self.fail("Set UQDA_TEST_BINARY to a real Uqda binary.")
        self.directory = tempfile.TemporaryDirectory(prefix="uqda-ui-")
        self.root = Path(self.directory.name)
        self.supervisor = Supervisor(self.root / "config", self.root / "runtime", BINARY, tun_name="none")
        self.supervisor.initialize()
        self.supervisor.start()
        self.assertTrue(self.supervisor.wait_ready())

    def tearDown(self):
        self.supervisor.stop()
        self.supervisor.lock_file.close()
        self.directory.cleanup()

    def values(self, **updates):
        values = {"revision": self.supervisor.revision(), "peers": [], "listen": [], "mode": "private"}
        values.update(updates)
        return values

    def test_new_node_is_isolated_and_private(self):
        config = self.supervisor.read_config()
        self.assertEqual(config["Peers"], [])
        self.assertEqual(config["Listen"], [])
        self.assertEqual(config["MulticastInterfaces"], [])
        self.assertGreaterEqual(len(config["GroupPassword"]), 32)
        self.assertEqual(self.supervisor.config_path.stat().st_mode & 0o777, 0o600)

    def test_restart_and_settings_retain_identity_and_unknown_fields(self):
        config = self.supervisor.read_config()
        key = config["PrivateKey"]
        config["NodeInfo"] = {"operator": "preserve-me"}
        atomic_write(self.supervisor.config_path, json.dumps(config).encode())
        before = self.supervisor.status()["identity"]["address"]
        self.supervisor.apply(self.values(group_password="new-shared-secret-0123456789"))
        self.supervisor.restart()
        after = self.supervisor.read_config()
        self.assertEqual(key, after["PrivateKey"])
        self.assertEqual(before, self.supervisor.status()["identity"]["address"])
        self.assertEqual(after["NodeInfo"], {"operator": "preserve-me"})
        self.assertEqual(after["GroupPassword"], "new-shared-secret-0123456789")
        self.assertTrue((self.supervisor.config_dir / "uqda.conf.previous").exists())

    def test_existing_configuration_survives_supervisor_upgrade(self):
        original = self.supervisor.config_path.read_bytes()
        address = self.supervisor.status()["identity"]["address"]
        self.supervisor.stop()
        self.supervisor.lock_file.close()
        self.supervisor = Supervisor(self.root / "config", self.root / "runtime", BINARY)
        self.supervisor.initialize()
        self.supervisor.start()
        self.assertTrue(self.supervisor.wait_ready())
        self.assertEqual(address, self.supervisor.status()["identity"]["address"])
        self.assertEqual(original, (self.supervisor.config_dir / "uqda.conf.before-umbrel").read_bytes())

    def test_status_never_returns_private_key_or_group_secret(self):
        config = self.supervisor.read_config()
        status = json.dumps(self.supervisor.status())
        self.assertNotIn(config["PrivateKey"], status)
        self.assertNotIn(config["GroupPassword"], status)
        self.assertNotIn("PrivateKey", status)
        self.assertNotIn("GroupPassword", status)

    def test_public_mode_requires_explicit_confirmation(self):
        original = self.supervisor.config_path.read_bytes()
        with self.assertRaises(ControlError):
            self.supervisor.apply(self.values(mode="public"))
        self.assertEqual(original, self.supervisor.config_path.read_bytes())
        self.supervisor.apply(self.values(mode="public", confirm_public=True))
        self.assertEqual(self.supervisor.read_config()["GroupPassword"], "")
        with self.assertRaises(ControlError):
            self.supervisor.apply(self.values(mode="private"))

    def test_services_persist_separately_without_changing_node_or_settings(self):
        original = self.supervisor.config_path.read_bytes()
        address = self.supervisor.status()["identity"]["address"]
        state = self.supervisor.service_action("service_add", {
            "revision": self.supervisor.services.snapshot()["revision"], "name": "My files", "kind": "https", "port": 8443})
        self.assertEqual(original, self.supervisor.config_path.read_bytes())
        self.assertEqual(self.supervisor.services.path.stat().st_mode & 0o777, 0o600)
        self.assertEqual(state["services"]["items"][0]["endpoint"], f"https://[{address}]:8443/")
        self.supervisor.restart()
        self.assertEqual(self.supervisor.status()["services"], state["services"])
        with self.assertRaisesRegex(ControlError, "TUN interface"):
            self.supervisor.service_action("service_probe", {"id": state["services"]["items"][0]["id"]})
        self.supervisor.service_action("service_remove", {
            "revision": state["services"]["revision"], "id": state["services"]["items"][0]["id"]})
        self.assertEqual(self.supervisor.status()["services"]["items"], [])
        self.assertEqual(original, self.supervisor.config_path.read_bytes())

    def test_stale_revision_and_invalid_input_do_not_change_config(self):
        original = self.supervisor.config_path.read_bytes()
        for values in [self.values(revision="stale"), self.values(peers=["tls://user:secret@example.com:443"]),
                       self.values(group_password="weak"), self.values(PrivateKey="attacker")]:
            with self.assertRaises(ControlError):
                self.supervisor.apply(values)
            self.assertEqual(original, self.supervisor.config_path.read_bytes())

    def test_start_failure_rolls_back_exact_bytes(self):
        original = self.supervisor.config_path.read_bytes()
        with patch.object(self.supervisor, "wait_ready", return_value=False):
            with self.assertRaisesRegex(ControlError, "Previous settings restored"):
                self.supervisor.apply(self.values(group_password="rollback-test-secret-01234"))
        self.assertEqual(original, self.supervisor.config_path.read_bytes())
        self.assertTrue(self.supervisor.wait_ready())

    def test_duplicate_supervisor_is_rejected(self):
        other = Supervisor(self.root / "config", self.root / "other-runtime", BINARY)
        try:
            with self.assertRaisesRegex(ControlError, "Another supervisor"):
                other.initialize()
        finally:
            other.lock_file.close()

    def test_advanced_peer_credentials_are_not_returned_or_overwritten(self):
        config = self.supervisor.read_config()
        config["Peers"] = ["tls://user:private-password@example.com:443?secret=hidden"]
        atomic_write(self.supervisor.config_path, json.dumps(config).encode())
        settings = self.supervisor.settings()
        self.assertFalse(settings["editable"])
        self.assertEqual(settings["peers"], ["tls://example.com:443"])
        with self.assertRaises(ControlError):
            self.supervisor.apply(self.values())

    def test_encoded_peer_credentials_are_not_exposed_or_overwritten(self):
        for query in ("%70assword=encoded-private-value", "pass%77ord=encoded-private-value",
                      "token=encoded-private-value", "%73ecret=encoded-private-value"):
            with self.subTest(query=query):
                config = self.supervisor.read_config()
                config["Peers"] = ["tls://example.com:443?" + query]
                atomic_write(self.supervisor.config_path, json.dumps(config).encode())
                before = self.supervisor.config_path.read_bytes()
                settings = self.supervisor.settings()
                self.assertFalse(settings["editable"])
                self.assertNotIn("encoded-private-value", json.dumps(settings))
                with self.assertRaises(ControlError):
                    self.supervisor.apply(self.values())
                self.assertEqual(before, self.supervisor.config_path.read_bytes())


class HttpTests(LifecycleTests):
    # Reuse lifecycle fixture, not the lifecycle test methods.
    def setUp(self):
        super().setUp()
        self.socket_path = str(self.root / "control.sock")
        self.control = ControlServer(self.socket_path, ControlHandler)
        self.control.supervisor = self.supervisor
        self.control_thread = threading.Thread(target=self.control.serve_forever, daemon=True)
        self.control_thread.start()
        self.web = Dashboard(("127.0.0.1", 0), self.socket_path, PASSWORD)
        self.web_thread = threading.Thread(target=self.web.serve_forever, daemon=True)
        self.web_thread.start()
        self.port = self.web.server_address[1]
        self.origin = "http://127.0.0.1:" + str(self.port)
        self.cookie = ""
        self.csrf = ""

    def tearDown(self):
        self.web.shutdown()
        self.web.server_close()
        self.control.shutdown()
        self.control.server_close()
        self.web_thread.join()
        self.control_thread.join()
        super().tearDown()

    def request(self, path, values=None, origin=None, csrf=None):
        connection = http.client.HTTPConnection("127.0.0.1", self.port, timeout=35)
        headers = {"Cookie": self.cookie, "X-Uqda-CSRF": self.csrf if csrf is None else csrf}
        method, body = "GET", None
        if values is not None:
            method, body = "POST", json.dumps(values)
            headers.update({"Content-Type": "application/json", "Origin": self.origin if origin is None else origin,
                            "X-Uqda-CSRF": self.csrf if csrf is None else csrf})
        connection.request(method, path, body, headers)
        response = connection.getresponse()
        data = response.read()
        status, response_headers = response.status, dict(response.getheaders())
        connection.close()
        return status, json.loads(data) if response_headers["Content-Type"] == "application/json" else data, response_headers

    def login(self):
        status, value, headers = self.request("/api/login", {"password": PASSWORD})
        self.assertEqual(status, 200)
        self.cookie = headers["Set-Cookie"].split(";")[0]
        self.csrf = value["csrf"]
        self.assertIn("HttpOnly", headers["Set-Cookie"])
        self.assertIn("SameSite=Strict", headers["Set-Cookie"])

    def test_authentication_and_logout(self):
        self.assertEqual(self.request("/api/status")[0], 401)
        self.assertEqual(self.request("/api/login", {"password": "wrong"})[0], 401)
        self.login()
        status, data, _ = self.request("/api/status")
        self.assertEqual(status, 200)
        self.assertTrue(data["ready"])
        self.assertEqual(self.request("/api/logout", {})[0], 200)
        self.assertEqual(self.request("/api/status")[0], 401)

    def test_origin_and_csrf_reject_cross_site_writes(self):
        self.assertEqual(self.request("/api/login", {"password": PASSWORD}, origin="http://evil.example")[0], 403)
        self.login()
        original = self.supervisor.config_path.read_bytes()
        self.assertEqual(self.request("/api/settings", self.values(), csrf="wrong")[0], 403)
        self.assertEqual(self.request("/api/settings", self.values(), origin="http://evil.example")[0], 403)
        self.assertEqual(original, self.supervisor.config_path.read_bytes())

    def test_authenticated_settings_and_restart(self):
        self.login()
        status, data, _ = self.request("/api/settings", self.values(group_password="http-private-secret-012345"))
        self.assertEqual(status, 200)
        self.assertTrue(data["settings"]["private"])
        self.assertEqual(self.request("/api/restart", {})[0], 200)

    def test_service_routes_require_authentication_origin_and_proof(self):
        self.assertEqual(self.request("/api/services/add", {})[0], 401)
        self.login()
        original = self.supervisor.config_path.read_bytes()
        values = {"revision": self.supervisor.services.snapshot()["revision"], "name": "Web app", "kind": "https", "port": 8443}
        for path in ("/api/services/add", "/api/services/remove", "/api/services/probe", "/api/umbrel/probe"):
            self.assertEqual(self.request(path, values, csrf="wrong")[0], 403)
            self.assertEqual(self.request(path, values, origin="http://evil.example")[0], 403)
        self.assertFalse(self.supervisor.services.path.exists())
        self.assertEqual(self.request("/api/umbrel/probe", {})[0], 400, "TUN-disabled fixture must not probe")
        status, state, _ = self.request("/api/services/add", values)
        self.assertEqual(status, 200)
        item = state["services"]["items"][0]
        self.assertEqual(self.request("/api/services/probe", {"id": item["id"], "host": "127.0.0.1"})[0], 400)
        self.assertEqual(self.request("/api/services/add", values)[0], 400, "stale writes must fail")
        self.assertEqual(self.request("/api/services/remove", {
            "revision": state["services"]["revision"], "id": item["id"]})[0], 200)
        self.assertEqual(original, self.supervisor.config_path.read_bytes())

    def test_cookie_alone_cannot_read_status_or_bootstrap_a_session(self):
        self.login()
        original = self.supervisor.config_path.read_bytes()
        for proof in ("", "wrong", "\u00e9"):
            status, session, _ = self.request("/api/session", csrf=proof)
            self.assertEqual(status, 200)
            self.assertEqual(session, {"authenticated": False, "csrf": ""})
            self.assertEqual(self.request("/api/status", csrf=proof)[0], 401)
            self.assertEqual(self.request("/api/settings", self.values(), csrf=proof)[0], 403)
        self.assertEqual(self.request("/api/status")[0], 200)
        self.assertEqual(self.supervisor.config_path.read_bytes(), original)

    def test_malformed_unicode_password_is_rejected_without_disconnect(self):
        self.assertEqual(self.request("/api/login", {"password": "\ud800"})[0], 400)
        self.assertEqual(self.request("/api/status")[0], 401)
        self.login()

    def test_non_ascii_csrf_is_rejected_without_changing_config(self):
        self.login()
        original = self.supervisor.config_path.read_bytes()
        self.assertEqual(self.request("/api/settings", self.values(), csrf="\u00e9")[0], 403)
        self.assertEqual(self.supervisor.config_path.read_bytes(), original)
        self.assertEqual(self.request("/api/status")[0], 200)

    def test_rate_limit_and_expired_sessions(self):
        self.login()
        with self.web.auth_lock:
            next(iter(self.web.sessions.values()))["expires"] = time.monotonic() - 1
        self.assertEqual(self.request("/api/status")[0], 401)
        for _ in range(10):
            self.assertEqual(self.request("/api/login", {"password": "wrong"})[0], 401)
        self.assertEqual(self.request("/api/login", {"password": PASSWORD})[0], 429)

    def test_assets_have_csp_and_no_external_dependencies(self):
        status, body, headers = self.request("/")
        self.assertEqual(status, 200)
        self.assertIn("script-src 'self'", headers["Content-Security-Policy"])
        self.assertIn(b'/app.js', body)
        self.assertEqual(self.request("/../../etc/passwd")[0], 404)

    def test_local_protocol_rejects_arbitrary_admin_commands(self):
        result = socket_json(self.socket_path, {"action": "addPeer", "uri": "tls://example.com:443"})
        self.assertFalse(result["ok"])


class ValidationTests(unittest.TestCase):
    def test_uris_and_redaction(self):
        self.assertEqual(validate_uris(["tls://[2001:db8::1]:443"]), ["tls://[2001:db8::1]:443"])
        for address in ["http://example.com:443", "tls://example.com", "tls://user:secret@example.com:443",
                        "tls://example.com:443?password=secret", "tls://example.com:443?%70assword=hidden",
                        "tls://example.com:443?token=hidden", "tls://example.com:443\n", "unix:///tmp/admin.sock"]:
            with self.assertRaises(ControlError):
                validate_uris([address])
        self.assertEqual(display_uri("tls://user:secret@example.com:443?password=hidden"), "tls://example.com:443")
        with self.assertRaises(ControlError):
            validate_uris(["tls://[::]:443"], listener=True)


# Don't run inherited lifecycle tests twice through HttpTests.
for name in list(LifecycleTests.__dict__):
    if name.startswith("test_") and name not in HttpTests.__dict__:
        setattr(HttpTests, name, None)


if __name__ == "__main__":
    unittest.main()
