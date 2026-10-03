"""Export only the installable community-store files, pinned to a verified image."""
import argparse
from pathlib import Path
import re
import zipfile

ROOT = Path(__file__).resolve().parents[2]


def export(digest, output, icon_ref="main"):
    if not re.fullmatch(r"sha256:[0-9a-f]{64}", digest):
        raise ValueError("Supply the published multiarchitecture manifest digest.")
    if not re.fullmatch(r"[a-zA-Z0-9_./-]+", icon_ref):
        raise ValueError("Invalid icon source ref.")
    version = (ROOT / "contrib/umbrel/VERSION").read_text().strip()
    source = "ghcr.io/uqda/core:" + version
    compose = (ROOT / "uqda-network/docker-compose.yml").read_text()
    compose = re.sub(re.escape(source) + r"(?:@sha256:[0-9a-f]{64})?", source + "@" + digest, compose)
    with zipfile.ZipFile(output, "w", compression=zipfile.ZIP_DEFLATED) as archive:
        archive.write(ROOT / "umbrel-app-store.yml", "umbrel-app-store.yml")
        archive.writestr("uqda-network/docker-compose.yml", compose)
        for path in sorted((ROOT / "uqda-network").rglob("*")):
            if path.is_file() and path.name != "docker-compose.yml":
                data = path.read_bytes()
                if path.name == "umbrel-app.yml":
                    data = data.replace(b"/Uqda/Core/main/contrib/", ("/Uqda/Core/" + icon_ref + "/contrib/").encode())
                archive.writestr(str(path.relative_to(ROOT)), data)
        archive.writestr("README.md", "# Uqda Community App Store\n\nUpload these files to the root of a public GitHub repository. Add that repository URL to Umbrel's Community App Stores.\n\nCore 26.0.4 with the Umbrel dashboard. Login with the app password shown by Umbrel. Add a trusted peer and your shared group password. Configuration and identity are saved in app data.\n\nSee https://github.com/Uqda/Core/blob/" + icon_ref + "/docs/umbrel.md for setup, permissions, backups and validation limits.\n")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--digest", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--icon-ref", default="main")
    options = parser.parse_args()
    export(options.digest, options.output, options.icon_ref)
