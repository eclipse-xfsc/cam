#!/usr/bin/env python3
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Contributors:
#	Fraunhofer AISEC

import re
from subprocess import run
import sys
import socket
import ssl
from logging import debug, info, warn, error
from api.common.evidence_pb2 import Error

import yaml
tlsversionmap = yaml.safe_load("""
tls1   : 1.0
tls1_1 : 1.1
tls1_2 : 1.2
tls1_3 : 1.3
""")


def tlsmeasure(hostname, port=443):
    return tlsmeasure_withopenssl(hostname, port)


def tlsmeasure_withpython(hostname, port=443):
    try:
        context = ssl.create_default_context()
        #con2 = ssl.SSLContext(ssl.PROTOCOL_TLSv1_1)
        with socket.create_connection((hostname, port)) as sock:
            with context.wrap_socket(sock, server_hostname=hostname) as ssock:
                return ssock.cipher()
    except (ssl.SSLError, ssl.SSLEOFError) as e:
        warn(f"Fail {e}. {hostname}:{port}")
    except Exception as e:
        warn(f"FAIL generic {e}. {hostname}:{port}")


def _openssl_error_handler(comproc):
    if comproc.returncode:
        #info(f"{hostname}:{port}:{proto}: {comproc.stderr.decode()}")
        if "Connection refused" in comproc.stderr.decode() or "Connection timed out" in comproc.stderr.decode():
            raise Exception(Error(code=Error.ERROR_CONNECTION_FAILURE))


def measure_version(hostname, port=443):
    supported_protocols = []
    for proto in ['tls1', 'tls1_1', 'tls1_2', 'tls1_3']:
        comproc = run(
            f"openssl s_client -connect {hostname}:{port} -{proto} < /dev/null", shell=True, capture_output=True)
        _openssl_error_handler(comproc)
        raw = comproc.stdout.decode()
        if not comproc.returncode:
            supported_protocols.append(tlsversionmap[proto])
    info(f"{hostname}:{port}: {supported_protocols}")
    return raw, supported_protocols


def measure_ciphersuites(hostname, port=443):
    supported_ciphersuites = []
    raw = ''
    protocolmap = {'SSLv3': 'ssl3', 'TLSv1': 'tls1',
                   'TLSv1.2': 'tls1_2', 'TLSv1.3': 'tls1_3'}
    for opensslcipher in getciphers():
        opensslcipher = opensslcipher[:2]
        protoswitch = protocolmap[opensslcipher[1]]
        if protoswitch == 'tls1_3':
            cipherswitch = 'ciphersuites'
        else:
            cipherswitch = 'cipher'
        cmd = f"openssl s_client -connect {hostname}:{port} -{protoswitch} -{cipherswitch} {opensslcipher[0]} < /dev/null"
        debug(cmd)
        comproc = run(cmd, shell=True, capture_output=True)
        _openssl_error_handler(comproc)
        opensslstdout = comproc.stdout.decode()
        opensslstderr = comproc.stderr.decode()
        raw += opensslstdout
        if comproc.returncode == 0:
            debug(f"{hostname}:{port} {opensslcipher[1]} {opensslcipher[0]}")
        if comproc.returncode != 0:
            if opensslstderr.startswith('s_client: Unknown option: -ssl3'):
                debug('SSL3 not supported by this openssl')
        debug(str(comproc.returncode) +
              str(re.search(r'Cipher is[^\n]*\n', opensslstdout)))
        if comproc.returncode == 0:
            supported_ciphersuites.append(opensslcipher)
    ret = [c[0] for c in supported_ciphersuites]
    info(f"{hostname}:{port}: {' '.join(ret)}")
    return raw, ret


def measure_certtrust(hostname, port=443):
    cmd = f"openssl s_client -connect {hostname}:{port} < /dev/null"
    comproc = run(cmd, shell=True, capture_output=True)
    _openssl_error_handler(comproc)
    raw = comproc.stdout.decode()
    if comproc.returncode == 0:
        certverification = re.search(r'Verification.*?:(.*?)\n', raw)
        status = certverification.group(1).strip()
        info(f"{hostname}:{port} Cert verification: {status}")
        if status == 'OK':
            return raw, True
        else:
            return raw, False
    else:
        return raw, None


def getciphers():
    opensslciphers = run(f"openssl ciphers -v", shell=True,
                         capture_output=True, check=True)
    return map(str.split, opensslciphers.stdout.decode().strip().split('\n'))


if __name__ == '__main__':
    from camlog import *
    import logging
    setuplogger(logging.DEBUG)
    import sys
    for target in sys.argv[1:]:
        if not ':' in target:
            target += ':443'
        host, port = target.split(':')
        raw, ret = measure_certtrust(host, port)
        # info(ret)
        raw, ret = measure_version(host, port)
        info(f"Error: {err}")
        raw, ret = measure_ciphersuites(host, port)
        # info(ret)
