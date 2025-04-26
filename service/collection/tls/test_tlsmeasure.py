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

from tls import *
import heartbleed
from camlog import *
import logging
setuplogger(logging.INFO)


hosts = """
gaia-x.eu
www.aisec.fraunhofer.de
""".split()

for host in hosts:
    tmp = host.split(':')
    host = tmp[0]
    if len(tmp) > 1:
        port = int(tmp[1])
    else:
        port = 443
    raw, ret = measure_certtrust(host, port)
    raw, ret = measure_version(host, port)
    raw, ret = measure_ciphersuites(host, port)
    ret = heartbleed.measure_heartbleed(host, port)
    info(f"Heart: {ret}")
