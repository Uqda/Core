#!/usr/bin/env python3
"""Pinned, independent daemon and real IPv6 dataplane interoperability gate.

Uses separate OS processes and actual TCP/TLS links; no TUN device is required.
The adapter only injects/observes IPv6 frames through each implementation's
unmodified ipv6rwc package. It does not implement routing or forward peer traffic.
"""
import base64
import hashlib
import ipaddress
import json
import os
from pathlib import Path
import queue
import socket
import struct
import subprocess
import tempfile
import threading
import time
import zipfile

ROOT = Path(__file__).resolve().parents[2]
PIN = json.loads((Path(__file__).with_name('upstream.json')).read_text())
ENV = {**os.environ, 'GOWORK': 'off', 'GOFLAGS': ''}


def run(args, cwd=ROOT, timeout=240):
    return subprocess.check_output(list(map(str, args)), cwd=cwd, env=ENV,
                                   stderr=subprocess.STDOUT, timeout=timeout).decode().strip()


def require(condition, message):
    if not condition:
        raise AssertionError(message)


def port():
    with socket.socket() as sock:
        sock.bind(('127.0.0.1', 0))
        return sock.getsockname()[1]


def admin(endpoint, name):
    with socket.create_connection(('127.0.0.1', endpoint), timeout=2) as sock:
        sock.settimeout(2)
        sock.sendall(json.dumps({'request': name}).encode() + b'\n')
        reply = json.loads(sock.makefile('rb').read(1 << 20))
        require(reply['status'] == 'success', str(reply))
        return reply['response']


def eventually(check, timeout=30):
    deadline = time.monotonic() + timeout
    last = None
    while time.monotonic() < deadline:
        try:
            if check():
                return
        except (OSError, ValueError, KeyError) as exc:
            last = exc
        time.sleep(.2)
    raise AssertionError(f'condition timed out: {last}')


class Node:
    def __init__(self, directory, name, binary, config, adapter=False):
        self.name, self.binary, self.adapter = name, binary, adapter
        self.conf = directory / (name + '.conf')
        self.conf.write_text(json.dumps(config))
        self.conf.chmod(0o600)
        self.admin = int(config['AdminListen'].rsplit(':', 1)[1])
        self.logpath = directory / (name + '.log')
        self.proc = None
        try:
            self.start()
        except Exception:
            self.stop()
            print(self.logpath.read_text(errors='replace'), flush=True)
            raise

    def start(self):
        self.events = queue.Queue()
        self.log = self.logpath.open('ab')
        args = [str(self.binary), str(self.conf)] if self.adapter else [str(self.binary), '-useconffile', str(self.conf)]
        self.proc = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE if self.adapter else self.log,
                                     stderr=self.log, env=ENV)
        if self.adapter:
            def consume(stream, events):
                for line in stream:
                    try:
                        events.put(json.loads(line))
                    except ValueError:
                        events.put({'event': 'invalid', 'line': repr(line)})
            self.reader = threading.Thread(target=consume, args=(self.proc.stdout, self.events), daemon=True)
            self.reader.start()
            self.identity = self.event('ready')
        else:
            eventually(lambda: admin(self.admin, 'getSelf'))
            self.identity = admin(self.admin, 'getSelf')
        self.key = self.identity['key']
        self.address = self.identity['address']
        require(ipaddress.ip_address(self.address) in ipaddress.ip_network('200::/7'), 'invalid address')
        data = json.loads(self.conf.read_text())
        require(self.key == data['PrivateKey'][64:], 'identity changed during startup')

    def send(self, value):
        require(self.proc.poll() is None, f'{self.name} exited')
        self.proc.stdin.write(json.dumps(value).encode() + b'\n')
        self.proc.stdin.flush()

    def event(self, kind, timeout=10):
        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            try:
                event = self.events.get(timeout=max(.01, deadline-time.monotonic()))
            except queue.Empty:
                break
            require(event['event'] != 'invalid', str(event))
            if event['event'] == kind:
                return event
        raise AssertionError(f'{self.name}: no {kind} event; {self.logpath.read_text()}')

    def peers(self):
        if self.adapter:
            self.send({'Command': 'peers'})
            return set(self.event('peers')['keys'])
        return {p['key'] for p in admin(self.admin, 'getPeers')['peers'] if p['up']}

    def stop(self):
        if self.proc:
            self.proc.terminate()
            try:
                self.proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.proc.kill()
                self.proc.wait(timeout=5)
            self.proc.stdin.close()
            if self.adapter:
                self.reader.join(timeout=2)
                self.proc.stdout.close()
            self.log.close()
            self.proc = None


def config(binary, listen, peers):
    cfg = json.loads(run([binary, '-genconf', '-json']))
    cfg.update(Listen=listen, Peers=peers, InterfacePeers={}, MulticastInterfaces=[],
               IfName='none', AdminListen=f'tcp://127.0.0.1:{port()}')
    return cfg


def connected(a, b):
    eventually(lambda: b.key in a.peers() and a.key in b.peers())
    require(a.key != b.key and a.address != b.address, 'nodes must have distinct identities')


def transfer(src, dst, label):
    # IPv6 experimental next-header 253 carries a unique, verifiable payload.
    payload = (label + ':' + os.urandom(16).hex()).encode()
    frame = struct.pack('!IHBB16s16s', 6 << 28, len(payload), 253, 64,
                        ipaddress.ip_address(src.address).packed, ipaddress.ip_address(dst.address).packed) + payload
    # Direct peer state becomes visible before the distributed routing tree and
    # encrypted end-to-end session have necessarily converged. Keep probing the
    # same authenticated frame through that convergence window.
    deadline = time.monotonic() + 90
    while time.monotonic() < deadline:
        src.send({'Command': 'send', 'Packet': base64.b64encode(frame).decode()})
        try:
            event = dst.events.get(timeout=1)
        except queue.Empty:
            continue
        if event['event'] == 'packet' and base64.b64decode(event['packet']) == frame:
            print('PASS packet', label, flush=True)
            return
    raise AssertionError(f'packet not delivered: {label}')


def build(directory):
    module = json.loads(run(['go', 'mod', 'download', '-json', PIN['Path'] + '@' + PIN['Version']]))
    for field in ('Path', 'Version', 'Sum', 'GoModSum'):
        require(module[field] == PIN[field], f'upstream {field} mismatch')
    require(module['Origin']['Hash'] == PIN['Commit'], 'upstream revision mismatch')
    archive = Path(module['Zip'])
    require(hashlib.sha256(archive.read_bytes()).hexdigest() == PIN['ZipSHA256'], 'upstream archive mismatch')
    source = directory / 'upstream-source'
    source.mkdir()
    prefix = PIN['Path'] + '@' + PIN['Version'] + '/'
    with zipfile.ZipFile(archive) as z:
        for entry in z.infolist():
            require(entry.filename.startswith(prefix), 'unexpected archive path')
            relative = Path(entry.filename[len(prefix):])
            require(not relative.is_absolute() and '..' not in relative.parts, 'unsafe archive path')
            dest = source / relative
            if entry.is_dir():
                dest.mkdir(parents=True, exist_ok=True)
            else:
                dest.parent.mkdir(parents=True, exist_ok=True)
                dest.write_bytes(z.read(entry))
    snapshot = {str(f.relative_to(source)): hashlib.sha256(f.read_bytes()).hexdigest() for f in source.rglob('*') if f.is_file()}
    modules = {'uqda': ('github.com/Uqda/Core', ROOT, 'uqda'),
               'upstream': (PIN['Path'], source, 'yggdrasil')}
    binaries = {}
    template = Path(__file__).with_name('fixture').joinpath('packet.go.in').read_text()
    for name, (mod, cwd, command) in modules.items():
        flags = [] if name == 'uqda' else ['-ldflags', f'-X {mod}/src/version.buildName=yggdrasil -X {mod}/src/version.buildVersion={PIN["Version"][1:]}']
        binary = directory / name
        run(['go', 'build', '-mod=readonly', '-buildvcs=false', '-trimpath', *flags, '-o', binary, './cmd/' + command], cwd)
        adapter = directory / (name + '-packet')
        adapter_source = directory / (name + '-packet.go')
        adapter_source.write_text(template.replace('__MODULE__', mod))
        run(['go', 'build', '-mod=readonly', '-buildvcs=false', '-trimpath', *flags, '-o', adapter, adapter_source], cwd)
        for built in (binary, adapter):
            metadata = run(['go', 'version', '-m', built])
            require(mod in metadata, f'wrong source module: {built}')
            if name == 'upstream':
                require('github.com/Uqda/Core' not in metadata, 'upstream built from Uqda')
            print('BINARY', built.name, hashlib.sha256(built.read_bytes()).hexdigest(), mod, flush=True)
        version = run([binary, '--version'])
        require(('yggdrasil' in version.lower() and PIN['Version'][1:] in version) if name == 'upstream' else version.startswith('Uqda Core '), 'wrong binary identity')
        print('VERSION', name, version.replace('\n', ' / '), flush=True)
        binaries[name] = (binary, adapter)
    require(binaries['uqda'][0].read_bytes() != binaries['upstream'][0].read_bytes(), 'same daemon binary')
    after = {str(f.relative_to(source)): hashlib.sha256(f.read_bytes()).hexdigest() for f in source.rglob('*') if f.is_file()}
    require(snapshot == after, 'upstream source was modified during build')
    print('UPSTREAM', PIN['Version'], PIN['Commit'], PIN['Sum'], 'source unchanged', flush=True)
    return binaries


def main():
    with tempfile.TemporaryDirectory(prefix='uqda-interop-') as temp:
        directory = Path(temp)
        binaries = build(directory)
        nodes = []
        try:
            for transport in ('tcp', 'tls'):
                for initiator in ('uqda', 'upstream'):
                    receiver = 'upstream' if initiator == 'uqda' else 'uqda'
                    uri = f'{transport}://127.0.0.1:{port()}'
                    a = Node(directory, f'{transport}-{receiver}', binaries[receiver][0], config(binaries[receiver][0], [uri], [])); nodes.append(a)
                    b = Node(directory, f'{transport}-{initiator}', binaries[initiator][0], config(binaries[initiator][0], [], [uri])); nodes.append(b)
                    connected(a, b)
                    # Cross-check addresses using each stock daemon's address command.
                    for node in (a, b):
                        derived = run([node.binary, '-useconffile', node.conf, '-address']).splitlines()[-1]
                        require(node.address == derived, 'daemon address derivation mismatch')
                    print('PASS daemons', transport, initiator, '->', receiver, flush=True)
                    uqda = a if receiver == 'uqda' else b
                    identity = (uqda.key, uqda.address)
                    uqda.stop(); uqda.start()
                    require(identity == (uqda.key, uqda.address), 'restart changed identity')
                    connected(a, b)
                    print('PASS daemon reconnect/identity', transport, initiator, flush=True)
                    a.stop(); b.stop()
                uri = f'{transport}://127.0.0.1:{port()}'
                # Only B listens: A and C cannot establish a direct peer link.
                b = Node(directory, transport+'-relay', binaries['upstream'][1], config(binaries['upstream'][0], [uri], []), True); nodes.append(b)
                a = Node(directory, transport+'-a', binaries['uqda'][1], config(binaries['uqda'][0], [], [uri]), True); nodes.append(a)
                c = Node(directory, transport+'-c', binaries['uqda'][1], config(binaries['uqda'][0], [], [uri]), True); nodes.append(c)
                connected(a, b); connected(c, b)
                require(a.peers() == {b.key} and c.peers() == {b.key} and b.peers() == {a.key, c.key}, 'wrong relay topology')
                for src, dst in ((a,b),(b,a),(a,c),(c,a)):
                    transfer(src, dst, transport+' '+src.name+' -> '+dst.name)
                identity = (b.key, b.address)
                b.stop(); b.start()
                require(identity == (b.key,b.address), 'relay restart changed identity')
                connected(a,b); connected(c,b)
                for src,dst in ((a,b),(b,a),(a,c),(c,a)):
                    transfer(src,dst,transport+' reconnect '+src.name+' -> '+dst.name)
                a.stop(); b.stop(); c.stop()
            print('PASS independent daemon, packet, reconnect, identity and multihop gates', flush=True)
        except Exception:
            for node in nodes:
                print('LOG',node.name,node.logpath.read_text(errors='replace'),flush=True)
            raise
        finally:
            for node in nodes:
                node.stop()


if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as exc:
        print(exc.output.decode(errors='replace'), flush=True)
        raise
