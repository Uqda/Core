"""Portable store-export tests: no daemon, Docker, or Linux-only dependencies."""
import importlib.util
from pathlib import Path
import tempfile
import shutil
import unittest
import zipfile
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[2]
spec = importlib.util.spec_from_file_location("export_store", ROOT / "contrib/umbrel/export_store.py")
exporter = importlib.util.module_from_spec(spec)
spec.loader.exec_module(exporter)


class ExportTests(unittest.TestCase):
    def test_local_configuration_and_backups_are_never_packaged(self):
        with tempfile.TemporaryDirectory() as directory:
            fixture = Path(directory) / "source"
            for relative in ("umbrel-app-store.yml", "contrib/umbrel/VERSION",
                             "uqda-network/umbrel-app.yml", "uqda-network/docker-compose.yml",
                             "uqda-network/data/config/.gitkeep", "uqda-network/data/control/.gitkeep"):
                destination = fixture / relative
                destination.parent.mkdir(parents=True, exist_ok=True)
                shutil.copyfile(ROOT / relative, destination)
            for name in ("uqda.conf", "uqda.conf.previous", "uqda.conf.before-umbrel"):
                (fixture / "uqda-network/data/config" / name).write_text("private-key-and-group-secret")
            output = Path(directory) / "store.zip"
            with patch.object(exporter, "ROOT", fixture):
                exporter.export("sha256:" + "a" * 64, output)
            with zipfile.ZipFile(output) as archive:
                self.assertEqual(len(archive.namelist()), 6)
                for name in archive.namelist():
                    self.assertNotIn(b"private-key-and-group-secret", archive.read(name))

    def test_export_pins_both_images_and_updates_previously_pinned_icon(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "store.zip"
            digest, revision = "sha256:" + "a" * 64, "b" * 40
            exporter.export(digest, output, revision)
            with zipfile.ZipFile(output) as archive:
                self.assertEqual(set(archive.namelist()), {
                    "umbrel-app-store.yml", "uqda-network/docker-compose.yml",
                    "uqda-network/umbrel-app.yml", "README.md",
                    "uqda-network/data/config/.gitkeep", "uqda-network/data/control/.gitkeep"})
                self.assertTrue(all("\\" not in name for name in archive.namelist()))
                self.assertEqual(archive.read("uqda-network/docker-compose.yml").decode().count("@" + digest), 2)
                manifest = archive.read("uqda-network/umbrel-app.yml").decode()
                self.assertIn("/Uqda/Core/" + revision + "/contrib/umbrel/web/icon.svg", manifest)
                self.assertIn("/blob/" + revision + "/docs/umbrel.md", archive.read("README.md").decode())

    def test_invalid_inputs_do_not_create_an_archive(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "store.zip"
            for digest, ref in [("sha256:invalid", "main"), ("sha256:" + "a" * 64, "bad ref")]:
                with self.assertRaises(ValueError):
                    exporter.export(digest, output, ref)
                self.assertFalse(output.exists())

    def test_mismatched_wrapper_version_is_rejected_before_export(self):
        original = Path.read_text
        def mismatched(path, *args, **kwargs):
            value = original(path, *args, **kwargs)
            return value.replace("ghcr.io/uqda/core:", "ghcr.io/uqda/wrong:") if path.name == "docker-compose.yml" else value
        with tempfile.TemporaryDirectory() as directory, patch.object(Path, "read_text", mismatched):
            output = Path(directory) / "store.zip"
            with self.assertRaises(ValueError):
                exporter.export("sha256:" + "a" * 64, output)
            self.assertFalse(output.exists())


if __name__ == "__main__":
    unittest.main()
