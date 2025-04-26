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

"""This is the TLS collector module.

It implements the Collection grpc service, and delivers results to Evaluation grpc service.

"""

import collections
import json
import logging
import os
import uuid
import threading
import traceback
import datetime
from concurrent import futures
from queue import Queue
from time import sleep
from typing import Iterable

import grpc
from google.protobuf import empty_pb2, struct_pb2, timestamp_pb2
from grpc_reflection.v1alpha import reflection
from oauthlib.oauth2 import BackendApplicationClient
from requests import request
from requests_oauthlib import OAuth2Session

import generate_protos
import heartbleed
import tls
from api.collection import collection_pb2
from api.collection.collection_pb2 import *  # messages
from api.collection.collection_pb2_grpc import *  # services
from api.common.evidence_pb2 import *
from api.evaluation.evaluation_pb2 import *
from api.evaluation.evaluation_pb2_grpc import *
from camlog import *

setuplogger(logging.INFO)
SDIR = os.path.dirname(os.path.abspath(__file__))


class EvidenceUploader(threading.Thread):
    evidences = Queue()

    def __init__(self, eval_manager):
        super().__init__()
        self.eval_manager = eval_manager
        self.start()

    def run(self):
        debug(
            f"starting evidence uploader for eval manager {self.eval_manager}, {self}")
        channel = grpc.intercept_channel(
            grpc.insecure_channel(self.eval_manager), AuthInterceptor())
        # tok = grpc.access_token_call_credentials(token)
        # chan = grpc.experimental.insecure_channel_credentials()
        # ccc = grpc.composite_channel_credentials(chan, tok)
        # channel=grpc.secure_channel(self.eval_manager, ccc)
        stub = EvaluationStub(channel)
        # iter() is wrapping the queue to be iterable
        stub.SendEvidences(iter(self.evidences.get, Evidence()))


executor = futures.ThreadPoolExecutor()
evalmanagers = dict[str, EvidenceUploader]()


class CollectorTls(CollectionServicer):
    monkeys = dict[int, futures.Future]()

    # '[::]:50051'):
    def __init__(self, bindto='0.0.0.0:50051', max_workers=100):
        info(f"TLS collector starting at {bindto}")
        # self.__dict__.update(locals())
        self.bindto = bindto
        self.max_workers = max_workers
        self.server = grpc.server(
            futures.ThreadPoolExecutor(max_workers=self.max_workers))
        add_CollectionServicer_to_server(self, self.server)
        reflection.enable_server_reflection(
            (collection_pb2.DESCRIPTOR.services_by_name['Collection'].full_name, reflection.SERVICE_NAME), self.server)
        self.server.add_insecure_port(self.bindto)
        self.server.start()

    def StartCollecting(self, req, context):
        info('StartCollectingRequest:' + str(req).replace('\n', ' '))
        jobid = uuid.uuid4()
        if not req.eval_manager in evalmanagers:
            debug(
                f"creating EvidenceUploader for eval_manager {req.eval_manager}")
            evalmanagers[req.eval_manager] = EvidenceUploader(req.eval_manager)

        config = CommunicationSecurityConfig()
        # unpack raw configuration
        req.configuration.raw_configuration.Unpack(config)

        # if you are the API wizard, you can correct me here:
        port = 443
        tmp = config.endpoint.split(':')
        hostname = tmp[0]
        if len(tmp) > 1:
            port = int(tmp[1])

        self.monkeys[jobid] = executor.submit(
            self.measure, jobid=jobid, service_id=req.service_id, hostname=hostname, port=port, eval_manager=req.eval_manager)
        return StartCollectingResponse(id=str(jobid))

    def StopCollecting(self, request, context):
        if not self.monkeys[request.id].cancel():
            warn(f"Job {request.id} is already running or completed")
        return empty_pb2.Empty()

    def StartCollectingStream(self, request_iterator, context):
        for screq in request_iterator:
            self.StartCollecting(screq, context=None)
        return empty_pb2.Empty()

    def stop(self, grace=None):
        # for jobid,futu in self.monkeys.items(): futu.cancel()
        executor.shutdown(cancel_futures=True)
        # StopIteration of Evidence queues, terminate evidenceuplader threads
        for evalmanager, v in evalmanagers.items():
            debug(f"stopping {v}")
            # empty evidence means end of the queue
            v.evidences.put(Evidence())
            v.join()
        return self.server.stop(grace)

    @property
    def client(self):
        return CollectionStub(grpc.insecure_channel(self.bindto))

    def measure(self, jobid, service_id, hostname, port: int, eval_manager: str):
        try:
            info(f"Starting measuring {hostname}:{port}")
            rawv, versions = tls.measure_version(hostname, port)
            info(f"{hostname}:{port} supports {versions}")
            rawt, certtrust = tls.measure_certtrust(hostname, port)
            info(f"{hostname}:{port} cert trusted: {certtrust}")
            rawc, ciphersuites = tls.measure_ciphersuites(hostname, port)
            info(f"{hostname}:{port} ciphersuites: {ciphersuites}")
            heartvuln = heartbleed.measure_heartbleed(hostname, port)
            info(f"{hostname}:{port} heartbleed: {heartvuln}")
            commonvulns = ["heartbleed"] if heartvuln else []
            ts = timestamp_pb2.Timestamp()
            ts.GetCurrentTime()
            s = struct_pb2.Struct()
            s.update({
                # id and type seems necessary because of our ontology. we need to properly model this struct after the ontology (TransportEncryption).
                "id": f"{hostname}:{port}",
                "type": ["TlsEndpoint"],
                "transportEncryption": {
                    "tlsVersion": versions,
                    "tlsCipherSuite": ciphersuites,
                    "tlsCertTrust": certtrust,
                    "tlsCommonVulns": commonvulns}
            })
            v = struct_pb2.Value(struct_value=s)
            raw = json.dumps(
                {"TlsVersion": rawv,    "TlsCipherSuite": rawc,        "TlsCertTrust": rawt})
            e = Evidence(id=str(jobid), target_service=service_id, tool_id="commsec", target_resource=f"{hostname}:{port}",
                         value=v, raw_evidence=raw, gathered_at=ts)

            debug(f"Sending evidence from job {jobid} to {eval_manager}. {e}")
            evalmanagers[eval_manager].evidences.put(e)
        except Exception as exc:
            warn(f"{host}:{port}: {exc}")
            debug(traceback.format_exc())
            evalmanagers[eval_manager].evidences.put(Evidence(
                id=str(jobid), target_service=service_id, tool_id="commsec", error=exc.args[0]))


class _ClientCallDetails(
        collections.namedtuple(
            '_ClientCallDetails',
            ('method', 'timeout', 'metadata', 'credentials')),
        grpc.ClientCallDetails):
    pass


class AuthInterceptor(grpc.StreamUnaryClientInterceptor):
    """this is needed because of a bug that does not allows us to
    transmit the credentials within the cluster without TLS with the regular API (see
    https://github.com/grpc/grpc/issues/29643) Hopefully, one day the min_security_level will be exposed on the call
    credentials API in python
    """

    client_id = os.environ['CAM_OAUTH2_CLIENT_ID']
    client_secret = os.environ['CAM_OAUTH2_CLIENT_SECRET']
    token_url = os.environ['CAM_OAUTH2_TOKEN_ENDPOINT']
    scope = os.environ['CAM_OAUTH2_SCOPES']
    client = BackendApplicationClient(client_id=client_id)
    oauth = OAuth2Session(client=client)

    token = None

    def intercept_stream_unary(self, continuation, client_call_details, request):
        metadata = []
        if client_call_details.metadata is not None:
            metadata = list(client_call_details.metadata)

        if self.token is None:
            info("Need to fetch a new token")
            self.token = self.oauth.fetch_token(token_url=self.token_url, client_id=self.client_id,
                                                client_secret=self.client_secret, scope=[self.scope])

        # TODO(oxisto): It seems that this is only called once at the beginning of the stream establishment. Do we need
        # to reconnect this at connection failure?
        metadata.append((
            'authorization',
            'bearer ' + self.token["access_token"],
        ))

        new_details = _ClientCallDetails(
            client_call_details.method, client_call_details.timeout, metadata,
            client_call_details.credentials)

        response = continuation(new_details, next(iter((request,))))
        return response


if __name__ == '__main__':
    collector = CollectorTls()
    collector.server.wait_for_termination()
