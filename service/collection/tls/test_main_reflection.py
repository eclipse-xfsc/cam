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

from grpc_reflection.v1alpha.proto_reflection_descriptor_database import ProtoReflectionDescriptorDatabase
import grpc
import generate_protos
from main import *
from camlog import *
setuplogger()


if __name__ == '__main__':
    collector = CollectorTls()

    channel = grpc.insecure_channel('localhost:50051')
    reflection_db = ProtoReflectionDescriptorDatabase(channel)
    from google.protobuf.descriptor_pool import DescriptorPool
    desc_pool = DescriptorPool(reflection_db)
    services = reflection_db.get_services()
    info(services)
    for s in services:
        service_desc = desc_pool.FindServiceByName(s)
        info(service_desc.methods_by_name.keys())
