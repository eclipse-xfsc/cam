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

package integrity

import (
	"context"
	"fmt"
	"time"

	ci "github.com/Fraunhofer-AISEC/cmc/cmcinterface"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	timeoutSec = 10
)

// Collect collects integrity information from prover
func Collect(serviceId string, nonce []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutSec*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, serviceId,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to service %v: %w", serviceId, err)
	}
	defer conn.Close()

	client := ci.NewCMCServiceClient(conn)

	request := ci.AttestationRequest{
		Nonce: nonce,
	}
	response, err := client.Attest(ctx, &request)
	if err != nil {
		return nil, fmt.Errorf("gRPC Attest call failed: %w", err)
	}
	if response.GetStatus() != ci.Status_OK {
		return nil, fmt.Errorf("gRPC Attest call returned status %v", response.GetStatus())
	}

	return response.AttestationReport, nil
}
