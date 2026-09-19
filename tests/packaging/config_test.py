"""Exercise the actual installer helper and daemon without system installation.

Run: python3 tests/packaging/config_test.py /absolute/path/to/uqda
"""
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

BINARY = str(Path(sys.argv.pop(1)).resolve())
HELPER = str(Path(__file__).resolve().parents[2] / 'contrib/packaging/install-config.sh')


class IdentityMigration(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.dest = self.root / 'uqda.conf'
        self.legacy = self.root / 'legacy.conf'
        self.config = subprocess.check_output([BINARY, '-genconf', '-json'])

    def install(self, *sources, success=True):
        result = subprocess.run(['sh', HELPER, str(self.dest), *map(str, sources)],
                                env={**os.environ, 'UQDA_BIN': BINARY}, capture_output=True)
        self.assertEqual(result.returncode == 0, success, result.stderr.decode())
        self.assertEqual(list(self.root.glob('*.tmp.*')), [])

    def test_migrate_and_reinstall_preserve_exact_identity(self):
        self.legacy.write_bytes(self.config)
        self.install(self.legacy)
        self.assertEqual(self.dest.read_bytes(), self.config)
        self.assertEqual(self.legacy.read_bytes(), self.config)
        self.assertEqual(self.dest.stat().st_mode & 0o777, 0o600)
        self.legacy.write_bytes(b'broken')
        self.install(self.legacy)
        self.assertEqual(self.dest.read_bytes(), self.config)

    def test_invalid_or_missing_identity_aborts(self):
        for content in (b'broken', b'{}', b'{"PrivateKey":"00"}'):
            self.legacy.write_bytes(content)
            self.install(self.legacy, success=False)
            self.assertFalse(self.dest.exists())
            self.assertEqual(self.legacy.read_bytes(), content)

    def test_invalid_existing_destination_is_not_overwritten(self):
        self.dest.write_bytes(b'{}')
        self.install(success=False)
        self.assertEqual(self.dest.read_bytes(), b'{}')

    def test_ambiguous_sources_abort(self):
        self.legacy.write_bytes(self.config)
        second = self.root / 'second.conf'
        second.write_bytes(self.config)
        self.install(self.legacy, second, success=False)
        self.assertFalse(self.dest.exists())

    def test_external_key_is_retained(self):
        self.legacy.write_bytes(self.config)
        key = self.root / 'identity.pem'
        self.legacy.chmod(0o600)
        key.write_bytes(subprocess.check_output([BINARY, '-useconffile', str(self.legacy), '-exportkey']))
        key.chmod(0o600)
        data = json.loads(self.config)
        del data['PrivateKey']
        data['PrivateKeyPath'] = str(key)
        content = json.dumps(data).encode()
        self.legacy.write_bytes(content)
        self.install(self.legacy)
        self.assertEqual(self.dest.read_bytes(), content)
        before = subprocess.check_output([BINARY, '-useconffile', str(self.legacy), '-publickey'])
        after = subprocess.check_output([BINARY, '-useconffile', str(self.dest), '-publickey'])
        self.assertEqual(before, after)

    def test_new_install_persists_generated_identity(self):
        self.install(self.legacy)
        original = self.dest.read_bytes()
        self.install()
        self.assertEqual(self.dest.read_bytes(), original)


if __name__ == '__main__':
    unittest.main()
