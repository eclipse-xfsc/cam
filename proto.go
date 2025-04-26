// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Contributors:
//	Fraunhofer AISEC

package cam

// Generate Configuration service
//go:generate protoc -I . -I ./third_party/ -I ./third_party/clouditor --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. --openapi_out=./api/configuration --grpc-gateway_out=paths=source_relative:. --grpc-gateway_opt logtostderr=true api/configuration/configuration.proto
// Generate Collection service
//go:generate protoc -I . -I ./third_party/ -I ./third_party/clouditor --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. --grpc-gateway_out=paths=source_relative:. --grpc-gateway_opt logtostderr=true api/collection/collection.proto
// Generate Evaluation service
//go:generate protoc -I . -I ./third_party/ -I ./third_party/clouditor --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. --openapi_out=./api/evaluation --grpc-gateway_out=paths=source_relative:. --grpc-gateway_opt logtostderr=true api/evaluation/evaluation.proto
// Generate evidence resource
//go:generate protoc -I . -I ./third_party/ -I ./third_party/clouditor --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. --grpc-gateway_out=paths=source_relative:. --grpc-gateway_opt logtostderr=true api/common/evidence.proto
// Run protobuf tagging
//go:generate protoc -I . -I third_party -I ./third_party/clouditor --gotag_out=paths=source_relative:. --gotag_opt=Mapi/evaluation/evaluation.proto=github.com/eclipse-xfsc/cam/api/evaluation api/evaluation/evaluation.proto
//go:generate protoc -I . -I third_party -I ./third_party/clouditor --gotag_out=paths=source_relative:. --gotag_opt=Mapi/common/evidence.proto=github.com/eclipse-xfsc/cam/api/evidence api/common/evidence.proto
//go:generate protoc -I . -I third_party -I ./third_party/clouditor --gotag_out=paths=source_relative:. --gotag_opt=Mapi/collection/collection.proto=github.com/eclipse-xfsc/cam/api/collection api/collection/collection.proto
