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

from main import *
from test_main import *
import fileinput

if __name__ == '__main__':
    setuplogger(logging.INFO)
    collector = CollectorTls()
    evaluationmock = EvaluationMock()

    def line2req(x): return StartCollectingRequest(
        service_id=f"{x.strip()}:443", eval_manager='localhost:50100')
    collector.client.StartCollectingStream(
        map(line2req, fileinput.FileInput()))

    # TODO: wait on all jobs completed instead of sleeping arbitrary number of seconds
    sleep(100)
    collector.stop(grace=2)
    evaluationmock.stop(grace=2)
