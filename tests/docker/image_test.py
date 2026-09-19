#!/usr/bin/env python3
"""Build and verify the image without privileged TUN access or registry pushes."""
import json
from pathlib import Path
import subprocess
import uuid

ROOT = Path(__file__).resolve().parents[2]


def run(*args):
    return subprocess.check_output(args, cwd=ROOT, stderr=subprocess.STDOUT, timeout=300).decode().strip()


def main():
    image = 'uqda-validation:' + uuid.uuid4().hex
    volume = 'uqda-validation-' + uuid.uuid4().hex
    run('docker', 'build', '-t', image, '.')
    try:
        expected = run('sh', 'contrib/semver/version.sh', '--display')
        assert run('docker', 'run', '--rm', '--entrypoint', 'uqda', image, '--version') == expected
        assert run('docker', 'run', '--rm', '--entrypoint', 'uqdactl', image, 'version') == expected
        run('docker', 'volume', 'create', volume)
        try:
            args = ('docker', 'run', '--rm', '-v', volume + ':/etc/uqda')
            run(*args, image, '-checkconf')
            first = run(*args, '--entrypoint', 'cat', image, '/etc/uqda/uqda.conf')
            key = run(*args, image, '-publickey')
            run(*args, image, '-checkconf')
            assert run(*args, '--entrypoint', 'cat', image, '/etc/uqda/uqda.conf') == first
            assert run(*args, image, '-publickey') == key
            run(*args, '--entrypoint', 'sh', image, '-c',
                'mv /etc/uqda/uqda.conf /etc/uqda/config.conf')
            run(*args, image, '-checkconf')
            assert run(*args, '--entrypoint', 'cat', image, '/etc/uqda/uqda.conf') == first
            # An invalid legacy file must fail without generating a replacement.
            run(*args, '--entrypoint', 'sh', image, '-c',
                "rm /etc/uqda/uqda.conf; printf '{}' > /etc/uqda/config.conf")
            result = subprocess.run((*args, image, '-checkconf'), cwd=ROOT, capture_output=True, timeout=30)
            assert result.returncode != 0, 'invalid migration accepted'
            run(*args, '--entrypoint', 'sh', image, '-c', 'test ! -e /etc/uqda/uqda.conf')
        finally:
            run('docker', 'volume', 'rm', volume)
        print('PASS image build, CLI identity, config path, persistence and migration')
    finally:
        run('docker', 'image', 'rm', image)


if __name__ == '__main__':
    main()
