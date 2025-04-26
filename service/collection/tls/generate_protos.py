#!/usr/bin/env python3
from grpc_tools import protoc
import os
SDIR = os.path.dirname(os.path.abspath(__file__))
REPOROOT = os.path.abspath(SDIR+'/../../../')

protoc.main(f"""42
-I{REPOROOT}
-I{REPOROOT}/third_party
-I{REPOROOT}/third_party/clouditor
--python_out={SDIR}
--grpc_python_out={SDIR}
api/common/evidence.proto
api/collection/collection.proto
api/assessment/metric.proto
api/evaluation/evaluation.proto
google/api/annotations.proto
google/api/http.proto
tagger/tagger.proto
""".split())
