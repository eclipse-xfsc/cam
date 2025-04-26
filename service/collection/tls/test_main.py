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

import generate_protos
from main import *
from camlog import *
setuplogger(logging.INFO)


class EvaluationMock(EvaluationServicer):
    def SendEvidences(self, request_iterator, context):
        info("starting evaluation mock")
        for e in request_iterator:
            esline = str(e).replace('\n', ' ')
            debug(f"Evidence arrived: {esline} {self.bindto}")
            info(f"Evidence arrived: {e.id} {e.target_service} {e.error} {self.bindto}")
        return empty_pb2.Empty()

    max_workers = 10

    def __init__(self, bindto='127.0.0.1:50100'):
        self.bindto = bindto
        self.server = grpc.server(
            futures.ThreadPoolExecutor(max_workers=self.max_workers))
        add_EvaluationServicer_to_server(self, self.server)
        self.server.add_insecure_port(self.bindto)
        self.server.start()

    def stop(self, grace=None):
        return self.server.stop(grace)


if __name__ == '__main__':
    evaluationmock = EvaluationMock()
    evaluationmock2 = EvaluationMock('127.0.0.2:50100')

    collector = CollectorTls()

    testrequests = [StartCollectingRequest(
        service_id='www.aisec.fraunhofer.de:443', eval_manager='127.0.0.1:50100')]
    testrequests += [StartCollectingRequest(
        service_id='gaia-x.eu:443', eval_manager='127.0.0.2:50100')]
    collector.client.StartCollecting(testrequests[0])
    collector.client.StartCollecting(testrequests[1])
    collector.client.StartCollectingStream(iter(testrequests))
    #collector.client.StartCollecting(StartCollectingRequest(service_id='www.google.com:444', eval_manager='127.0.0.1:50100')) #TCP timeout case
    sleep(24)
    print("TEST OVER")
    collector.stop(grace=3)
    evaluationmock.stop(grace=3)
    evaluationmock2.stop(grace=3)
    sleep(4)
