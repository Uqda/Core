"""Structural checks independent of PyYAML/Docker or Umbrel's injected services."""
from pathlib import Path
import re
import sys

root = Path(__file__).resolve().parents[2]
version = (root / "contrib/umbrel/VERSION").read_text().strip()
manifest = (root / "uqda-network/umbrel-app.yml").read_text()
compose = (root / "uqda-network/docker-compose.yml").read_text()
assert re.fullmatch(r"\d+\.\d+\.\d+-umbrel\.[1-9]\d*", version)
assert "id: uqda-network\n" in manifest
assert 'version: "' + version + '"' in manifest
assert "deterministicPassword: true" in manifest
assert "UQDA_DASHBOARD_PASSWORD: ${APP_PASSWORD}" in compose
assert "APP_HOST: uqda-network_dashboard_1" in compose
assert "APP_PORT: 8080" in compose
assert "privileged:" not in compose and "docker.sock" not in compose
assert "${APP_DATA_DIR}/data/config:/etc/uqda" in compose
assert "ports:" not in compose
assert compose.count("image: ghcr.io/uqda/core:" + version) == 2
if "--require-digest" in sys.argv:
    refs = re.findall(r"image: (\S+)", compose)
    assert all(re.fullmatch(r"ghcr\.io/uqda/core:" + re.escape(version) + r"@sha256:[0-9a-f]{64}", image) for image in refs)
print("PASS: community store structure, version, auth, persistence and privilege boundaries")
