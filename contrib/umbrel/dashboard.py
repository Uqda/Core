"""Unprivileged browser UI. Only a bounded local control protocol crosses to Core."""
from collections import deque
from http.cookies import CookieError, SimpleCookie
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import hmac
import json
import os
from pathlib import Path
import secrets
import threading
import time
from urllib.parse import urlsplit

from control import ControlError, MAX_MESSAGE, socket_json


class Dashboard(ThreadingHTTPServer):
    daemon_threads = True

    def __init__(self, address, control_path, password):
        if not password or len(password) < 16:
            raise ValueError("UQDA_DASHBOARD_PASSWORD must contain at least 16 characters.")
        self.password = password.encode()
        self.control_path = control_path
        self.sessions = {}
        self.failed_logins = deque()
        self.auth_lock = threading.Lock()
        super().__init__(address, Handler)


class Handler(BaseHTTPRequestHandler):
    server_version = "UqdaDashboard"

    def setup(self):
        super().setup()
        self.connection.settimeout(35)

    def log_message(self, _format, *_args):
        # No URLs, bodies, passwords, cookies or node configuration in HTTP logs.
        pass

    def reply(self, status, value, content_type="application/json", cookie=None):
        body = value if isinstance(value, bytes) else json.dumps(value).encode()
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Cache-Control", "no-store")
        self.send_header("X-Content-Type-Options", "nosniff")
        self.send_header("Referrer-Policy", "no-referrer")
        self.send_header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; frame-ancestors 'self'; base-uri 'none'; form-action 'self'")
        if cookie:
            self.send_header("Set-Cookie", cookie)
        self.end_headers()
        self.wfile.write(body)

    def read_json(self):
        if self.headers.get("Content-Type", "").split(";")[0] != "application/json":
            raise ControlError("Use JSON requests.")
        if self.headers.get("Transfer-Encoding"):
            raise ControlError("Unsupported request encoding.")
        try:
            length = int(self.headers.get("Content-Length", "0"))
        except ValueError:
            raise ControlError("Invalid request size.") from None
        if not 1 <= length <= MAX_MESSAGE:
            raise ControlError("Invalid request size.")
        value = json.loads(self.rfile.read(length))
        if not isinstance(value, dict):
            raise ControlError("Invalid request.")
        return value

    def session(self):
        try:
            cookies = SimpleCookie(self.headers.get("Cookie", ""))
            token = cookies["uqda_session"].value
        except (KeyError, ValueError, CookieError):
            return None
        with self.server.auth_lock:
            value = self.server.sessions.get(token)
            if value and value["expires"] > time.monotonic():
                return value
            self.server.sessions.pop(token, None)
        return None

    def same_origin(self):
        origin = self.headers.get("Origin")
        if not origin:
            return False
        try:
            parsed = urlsplit(origin)
        except ValueError:
            return False
        # Trust only the actual Host routed by app_proxy; not forwarded headers.
        return parsed.scheme in ("http", "https") and parsed.netloc == self.headers.get("Host") and parsed.path == ""

    def control(self, action, settings=None):
        request = {"action": action}
        if settings is not None:
            request["settings"] = settings
        try:
            response = socket_json(self.server.control_path, request, timeout=32)
        except (OSError, ValueError, ControlError):
            self.reply(503, {"error": "Core control is unavailable. Try again shortly."})
            return
        if response.get("ok"):
            self.reply(200, response["result"])
        else:
            self.reply(400, {"error": response.get("error", "Request failed.")})

    def do_GET(self):
        path = urlsplit(self.path).path
        assets = {"/": ("index.html", "text/html; charset=utf-8"),
                  "/app.js": ("app.js", "text/javascript; charset=utf-8"),
                  "/style.css": ("style.css", "text/css; charset=utf-8"),
                  "/icon.svg": ("icon.svg", "image/svg+xml")}
        if path in assets:
            name, mime = assets[path]
            self.reply(200, (Path(__file__).parent / "web" / name).read_bytes(), mime)
        elif path == "/healthz":
            self.reply(200, {"ok": True})
        elif path == "/api/session":
            session = self.session()
            self.reply(200, {"authenticated": bool(session), "csrf": session["csrf"] if session else ""})
        elif path == "/api/status":
            if not self.session():
                self.reply(401, {"error": "Sign in to continue."})
            else:
                self.control("status")
        else:
            self.reply(404, {"error": "Not found."})

    def do_POST(self):
        if not self.same_origin():
            self.reply(403, {"error": "Request origin rejected."})
            return
        try:
            values = self.read_json()
        except (ControlError, ValueError, OSError):
            self.reply(400, {"error": "Invalid JSON request."})
            return
        path = urlsplit(self.path).path
        if path == "/api/login":
            self.login(values)
            return
        session = self.session()
        if not session:
            self.reply(401, {"error": "Sign in to continue."})
            return
        if not hmac.compare_digest(self.headers.get("X-Uqda-CSRF", ""), session["csrf"]):
            self.reply(403, {"error": "Refresh the page before trying again."})
            return
        if path == "/api/settings":
            self.control("apply", values)
        elif path == "/api/restart":
            self.control("restart")
        elif path == "/api/logout":
            with self.server.auth_lock:
                for token, value in list(self.server.sessions.items()):
                    if value is session:
                        del self.server.sessions[token]
            self.reply(200, {"ok": True}, cookie="uqda_session=; Path=/; HttpOnly; SameSite=Strict; Max-Age=0")
        else:
            self.reply(404, {"error": "Not found."})

    def login(self, values):
        supplied = values.get("password")
        if not isinstance(supplied, str) or len(supplied) > 1024:
            self.reply(400, {"error": "Invalid password."})
            return
        with self.server.auth_lock:
            now = time.monotonic()
            while self.server.failed_logins and self.server.failed_logins[0] < now - 60:
                self.server.failed_logins.popleft()
            if len(self.server.failed_logins) >= 10:
                self.reply(429, {"error": "Too many attempts. Try again in one minute."})
                return
            if not hmac.compare_digest(supplied.encode(), self.server.password):
                self.server.failed_logins.append(now)
                self.reply(401, {"error": "Incorrect password."})
                return
            self.server.sessions = {key: value for key, value in self.server.sessions.items() if value["expires"] > now}
            if len(self.server.sessions) >= 32:
                self.server.sessions.pop(next(iter(self.server.sessions)))
            token = secrets.token_urlsafe(32)
            csrf = secrets.token_urlsafe(32)
            self.server.sessions[token] = {"csrf": csrf, "expires": now + 8 * 3600}
        secure = "; Secure" if urlsplit(self.headers.get("Origin", "")).scheme == "https" else ""
        self.reply(200, {"authenticated": True, "csrf": csrf},
                   cookie="uqda_session=" + token + "; Path=/; HttpOnly; SameSite=Strict; Max-Age=28800" + secure)


def main():
    password = os.environ.pop("UQDA_DASHBOARD_PASSWORD", "")
    server = Dashboard(("0.0.0.0", int(os.environ.get("PORT", "8080"))),
                       os.environ.get("UQDA_CONTROL_SOCKET", "/run/uqda-control/control.sock"), password)
    server.serve_forever()


if __name__ == "__main__":
    main()
