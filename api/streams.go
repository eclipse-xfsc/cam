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

package api

import (
	"context"
	"fmt"

	"github.com/eclipse-xfsc/cam/api/evaluation"
	"google.golang.org/grpc"
)

// TODO(oxisto): Once https://github.com/golang/go/issues/46477 is implemented, we can have type
// alias here to expose the clouditor/api generic types using a type alias, so we do not need to
// import api and clouditor_api in our services.

// InitEvalStream initializes a stream to the Evaluation Manager
func InitEvalStream(hostport string, additionalOpts ...grpc.DialOption) (stream evaluation.Evaluation_SendEvidencesClient, err error) {
	// Establish connection to Evaluation Manager gRPC service
	conn, err := grpc.Dial(hostport, additionalOpts...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to evaluation manager service: %w", err)
	}

	// Create new streaming clientw
	client := evaluation.NewEvaluationClient(conn)
	stream, err = client.SendEvidences(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not set up stream to evaluation manager service: %w", err)
	}

	return
}
