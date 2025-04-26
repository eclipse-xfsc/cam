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

package testevaluation

import (
	"context"
	"log"
	"net"

	"clouditor.io/clouditor/api/orchestrator"
	"clouditor.io/clouditor/persistence/gorm"

	"github.com/eclipse-xfsc/cam/api/configuration"
	api_evaluation "github.com/eclipse-xfsc/cam/api/evaluation"
	service_configuration "github.com/eclipse-xfsc/cam/service/configuration"
	service_evaluation "github.com/eclipse-xfsc/cam/service/evaluation"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const DefaultBufferSize = 1024 * 1024

var (
	bufConnListener *bufconn.Listener
)

func BufConnDialer(context.Context, string) (net.Conn, error) {
	return bufConnListener.Dial()
}

// StartBufConnServerToEvaluation starts an gRPC listening on a bufconn listener. It exposes
// real functionality of the following service for testing purposes:
// * cam-eval-manager
// * cam-req-manager
func StartBufConnServerToEvaluation() (*grpc.Server, api_evaluation.EvaluationServer, configuration.ConfigurationServer) {
	bufConnListener = bufconn.Listen(DefaultBufferSize)

	server := grpc.NewServer()

	db, err := gorm.NewStorage(gorm.WithInMemory())
	if err != nil {
		log.Fatalf("Couldn't create storage for bufconn listener: %v", err)
	}
	evaluationServer := service_evaluation.NewServer(
		service_evaluation.WithStorage(db),
		service_evaluation.WithRequirementsManagerAddress("bufnet", grpc.WithContextDialer(BufConnDialer)))
	api_evaluation.RegisterEvaluationServer(server, evaluationServer)

	configServer := service_configuration.NewServer()
	configuration.RegisterConfigurationServer(server, configServer)
	orchestrator.RegisterOrchestratorServer(server, configServer)

	go func() {
		if err := server.Serve(bufConnListener); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	return server, evaluationServer, configServer
}
