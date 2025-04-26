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
	"crypto/rand"
	"encoding/json"
	"fmt"

	clapi "clouditor.io/clouditor/api"
	"clouditor.io/clouditor/voc"
	"github.com/Fraunhofer-AISEC/cmc/attestationreport"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/eclipse-xfsc/cam/api"
	"github.com/eclipse-xfsc/cam/api/collection"
	apicollection "github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/common"
	"github.com/eclipse-xfsc/cam/api/evaluation"
	"github.com/eclipse-xfsc/cam/internal/config"
	"github.com/eclipse-xfsc/cam/service"
)

var (
	log         = logrus.WithField("service", "collection-integrity")
	metricId    = "SystemComponentsIntegrity"
	ComponentID = config.DefaultCollectionIntegrityID
)

type Server struct {
	apicollection.UnimplementedCollectionServer

	streams  *clapi.StreamsOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence]
	grpcOpts []grpc.DialOption

	authorizer clapi.Authorizer
}

// WithAdditionalGRPCOpts is an option to configure additional gRPC options.
func WithAdditionalGRPCOpts(opts ...grpc.DialOption) service.ServiceOption[Server] {
	return func(s *Server) {
		s.grpcOpts = opts
	}
}

// WithOAuth2Authorizer is an option to use an OAuth 2.0 authorizer
func WithOAuth2Authorizer(config *clientcredentials.Config) service.ServiceOption[Server] {
	return func(srv *Server) {
		srv.SetAuthorizer(clapi.NewOAuthAuthorizerFromClientCredentials(config))
	}
}

// SetAuthorizer implements UsesAuthorizer
func (srv *Server) SetAuthorizer(auth clapi.Authorizer) {
	srv.authorizer = auth
}

// Authorizer implements UsesAuthorizer
func (srv *Server) Authorizer() clapi.Authorizer {
	return srv.authorizer
}

func NewServer(opts ...service.ServiceOption[Server]) apicollection.CollectionServer {
	s := &Server{
		streams: clapi.NewStreamsOf(clapi.WithLogger[evaluation.Evaluation_SendEvidencesClient, *common.Evidence](log)),
	}

	// Apply any options
	for _, o := range opts {
		o(s)
	}

	return s
}

func (s *Server) StartCollecting(_ context.Context, req *apicollection.StartCollectingRequest) (
	res *apicollection.StartCollectingResponse, err error) {
	log.Infof("Received StartCollecting Request for Service ID %v", req.ServiceId)

	var rawConfig collection.RemoteIntegrityConfig
	err = req.Configuration.RawConfiguration.UnmarshalTo(&rawConfig)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, apicollection.ErrInvalidRemoteIntegrityRawConfiguration.Error())
	}

	capem := []byte(rawConfig.Certificate)

	// Collecting integrity information from external service requires nonce
	// to avoid replay attacks
	nonce := make([]byte, 8)
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate random bytes (%v)", err)
	}

	log.Tracef("Collecting integrity information from service: %v\n", req.ServiceId)
	ar, err := Collect(rawConfig.Target, nonce)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to collect information from service %v: %v",
			req.ServiceId, err)
	}
	log.Tracef("Collected integrity information from service: %v\n", req.ServiceId)

	result := attestationreport.Verify(string(ar), nonce, capem, nil)
	if !result.Success {
		log.Tracef("Verification of integrity information failed - Service is not trustworthy")
	}

	rawEvidence, err := json.Marshal(result)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal integrity result: %v", err)
	}

	value := Value{
		Resource: voc.Resource{
			// ID and Type has to be set. Otherwise, evaluation will fail due to evidence validation
			ID:   "JustNotEmpty",
			Type: []string{"SomeType"},
		},
		SystemComponentsIntegrity: SystemComponentsIntegrity{Status: result.Success},
	}
	evidenceValue, err := toStruct(value)
	if err != nil {
		err = fmt.Errorf("could not convert struct to structpb.Value: %w", err)
		log.Error(err)
	}

	log.Tracef("Sending evidences to eval manager %v", req.EvalManager)

	// Get stream for the Evaluation Manager
	component := "Evaluation Manager"
	stream, err := s.streams.GetStream(req.EvalManager, component, api.InitEvalStream,
		clapi.DefaultGrpcDialOptions(req.EvalManager, s, s.grpcOpts...)...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	requestId := uuid.NewString()

	evidence := &common.Evidence{
		Id:             requestId,
		Name:           requestId,
		TargetService:  req.ServiceId,
		TargetResource: result.PlainAttReport.DeviceDescription.Fqdn,
		ToolId:         ComponentID,
		GatheredAt:     timestamppb.Now(),
		Value:          evidenceValue,
		RawEvidence:    string(rawEvidence),
	}

	// Send evidence to stream via channel
	stream.Send(evidence)
	log.Infof("Sending evidence '%s' to evaluation manager stream", evidence.Id)

	res = &apicollection.StartCollectingResponse{
		Id: requestId,
	}
	return
}

func (s *Server) StopCollecting(_ context.Context, req *apicollection.StopCollectingRequest) (*emptypb.Empty, error) {
	log.Tracef("Received StopCollecting Request for Service ID %v", req.Id)

	return nil, status.Error(codes.Unimplemented, "StopCollection not implemented")
}
