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


FROM python:3-slim
COPY service/collection/tls/requirements.txt /app/services/collection/tls/requirements.txt
WORKDIR /app/services/collection/tls
RUN pip install -r requirements.txt
COPY service/collection/tls /app/services/collection/tls
COPY api /app/api
COPY third_party /app/third_party

# The ingress is not sending the intermediate certificate so we need to bring it along
ADD third_party/rapidssl.crt /usr/local/share/ca-certificates/rapidssl.crt
RUN chmod 644 /usr/local/share/ca-certificates/rapidssl.crt && update-ca-certificates

EXPOSE 50051

#CMD pipenv run /app/services/collection/tls/main.py
CMD python /app/services/collection/tls/main.py
