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

// Package authsec contains service specific code for the Authentication Security Collection Module
package authsec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	clapi "clouditor.io/clouditor/api"
	"clouditor.io/clouditor/voc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/eclipse-xfsc/cam/api"
	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/common"
	"github.com/eclipse-xfsc/cam/api/evaluation"
	"github.com/eclipse-xfsc/cam/internal/config"
	"github.com/eclipse-xfsc/cam/internal/protobuf"
	"github.com/eclipse-xfsc/cam/service"
)

var (
	log         = logrus.WithField("service", "collection-authsec")
	ComponentID = config.DefaultCollectionAuthSecID
)

type Server struct {
	collection.UnimplementedCollectionServer
	streams    *clapi.StreamsOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence]
	grpcOpts   []grpc.DialOption
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

func NewServer(opts ...service.ServiceOption[Server]) collection.CollectionServer {
	s := &Server{
		streams: clapi.NewStreamsOf(clapi.WithLogger[evaluation.Evaluation_SendEvidencesClient, *common.Evidence](log)),
	}

	// Apply any options
	for _, o := range opts {
		o(s)
	}

	return s
}

func getAndValidateMetadata(config *collection.AuthenticationSecurityConfig) (*map[string]interface{}, *common.Error) {
	// Parse issuer identifier
	issuerID := config.Issuer
	if issuerID == "" {
		return nil, &common.Error{
			Code:        common.Error_ERROR_INVALID_CONFIGURATION,
			Description: "No issuer identifier given",
		}
	}
	issuer, err := url.Parse(issuerID)
	if err != nil {
		return nil, &common.Error{
			Code:        common.Error_ERROR_INVALID_CONFIGURATION,
			Description: "Error parsing issuer identifier: " + err.Error(),
		}
	}

	// Optionally parse metadataURL if given explicitly
	var metadataURL *url.URL = nil
	if metadataDocument := config.MetadataDocument; metadataDocument != "" {
		metadataURL, err = url.Parse(metadataDocument)
		if err != nil {
			return nil, &common.Error{
				Code:        common.Error_ERROR_INVALID_CONFIGURATION,
				Description: "Error parsing metadata URL: " + err.Error(),
			}
		}
	}

	// Fetch the metadata document
	metadata, errStruct := getMetadata(issuer, metadataURL)
	if errStruct != nil {
		return nil, errStruct
	}
	logrus.Traceln("Found Metadata document for " + (*metadata)["issuer"].(string))

	// Sanity check it
	metadata, err = checkMetadataRFC8414(metadata)
	if err != nil {
		if err != nil {
			return nil, &common.Error{
				Code:        common.Error_ERROR_PROTOCOL_VIOLATION,
				Description: "Error validating Metadata document (RFC 8414): " + err.Error(),
			}
		}
	}
	metadata, err = checkMetadataOIDCDiscovery(metadata)
	if err != nil {
		return nil, &common.Error{
			Code:        common.Error_ERROR_PROTOCOL_VIOLATION,
			Description: "Error validating Metadata document (OIDC Discovery): " + err.Error(),
		}
	}

	return metadata, nil
}

// checkConfiguration performs some basic input validation on the configuration data sent with the request
// Note that this is merely a static best effort and errors may arise later on.
// If that is the case, the errors are reported to the evaluation manager instead of the caller.
func checkConfiguration(req *collection.StartCollectingRequest) (config *collection.AuthenticationSecurityConfig, err error) {
	if req.Configuration == nil {
		err = errors.New("configuration is missing")
		return
	}
	if req.Configuration.ServiceId == "" {
		err = errors.New("ServiceID in Configuration is missing")
		return
	}
	if req.Configuration.RawConfiguration == nil {
		err = errors.New("RawConfiguration in Configuration is missing")
		return
	}
	if !req.Configuration.RawConfiguration.MessageIs(&collection.AuthenticationSecurityConfig{}) {
		err = errors.New("RawConfiguration is not an Authentication Security Config")
		return
	}

	config = new(collection.AuthenticationSecurityConfig)
	err = req.Configuration.RawConfiguration.UnmarshalTo(config)

	if err != nil {
		err = errors.New("RawConfiguration is not an Authentication Security Config")
		return
	}

	// We try to gather different evidences based on which parameters are set in
	// the service configuration
	if config.ApiEndpoint != "" {
		// Required Config: issuer (URL)
		if config.Issuer == "" {
			err = errors.New("missing required config parameter: issuer")
			return
		}
		if _, err = url.Parse(config.Issuer); err != nil {
			err = errors.New("error parsing issuer: " + err.Error())
			return
		}
		// Optional Config: metadata_document (URL)
		if metadata := config.MetadataDocument; metadata != "" {
			if _, err = url.Parse(metadata); err != nil {
				err = errors.New("error parsing metadata_document: " + err.Error())
				return
			}
		}
		if _, err = url.Parse(config.ApiEndpoint); err != nil {
			err = errors.New("error parsing API endpoint: " + err.Error())
			return
		}
	} else {
		// Required Config: issuer (URL)
		if config.Issuer == "" {
			err = errors.New("missing required config parameter: issuer")
			return
		}
		if _, err = url.Parse(config.Issuer); err != nil {
			err = errors.New("error parsing issuer: " + err.Error())
			return
		}
		// Optional Config: metadata_document (URL)
		if metadata := config.MetadataDocument; metadata != "" {
			if _, err = url.Parse(metadata); err != nil {
				err = errors.New("error parsing metadata_document: " + err.Error())
				return
			}
		}
	}
	return
}

// handleCollectionRequest runs the actual Collection Request, after the initial sanity checks.
// If any problems arise, an error is reported to the evaluation manager
// TODO(oxisto): EnqueueEvidences should be used instead or adapted
func handleCollectionRequest(serviceId string,
	evidenceStream *clapi.StreamChannelOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence],
	config *collection.AuthenticationSecurityConfig) {
	var err error
	var evidence *common.Evidence

	// In any case, we are collecting evidence about the OAuth 2.0 endpoint
	evidence, err = collectOAuth2Evidence(serviceId, config)
	if err != nil {
		err = fmt.Errorf("internal error while collecting OAuth2.0 evidence: %w", err)
		log.Error(err)
		return
	}
	enqueueEvidence(evidenceStream, evidence)

	// Optionally, we are also gathering evidence about the API endpoint and whether it is protected
	if config.ApiEndpoint != "" {
		evidence, err = collectAPIAccessEvidence(serviceId, config)
		if err != nil {
			err = fmt.Errorf("internal error while collecting OAuth2.0 evidence: %w", err)
			log.Error(err)
			return
		}
		enqueueEvidence(evidenceStream, evidence)
	}
}

// TODO(oxisto): Migrate this to the already existing collection.EnqueueEvidence, once we migrate the evidence value to
// the ontology.
func enqueueEvidence(evidenceStream *clapi.StreamChannelOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence], evidence *common.Evidence) {
	if evidence.Error != nil {
		log.Warnf("Reporting an error: %s", evidence.Error.Description)
	} else {
		log.Tracef("Sending '%v' to Evaluation Manager", evidence.Value)
	}

	evidenceStream.Send(evidence)

	log.Infof("Sent evidence {id: %s, target_resource: %s } to evaluation manager", evidence.Id, evidence.TargetResource)
}

func (s *Server) StartCollecting(_ context.Context, req *collection.StartCollectingRequest) (*collection.StartCollectingResponse, error) {
	log.Infof("Received StartCollecting Request for Service ID '%v'", req.ServiceId)
	// Parse and check configuration data
	// If problems are detected at this stage, they are returned to the caller
	// instead of being forwarded to the evaluation manager
	config, err := checkConfiguration(req)
	if err != nil {
		return nil, grpcstatus.Errorf(codes.InvalidArgument, "%v", err)
	}

	// Get stream for the Evaluation Manager
	component := "Evaluation Manager"
	evidenceStream, err := s.streams.GetStream(req.EvalManager, component, api.InitEvalStream,
		clapi.DefaultGrpcDialOptions(req.EvalManager, s, s.grpcOpts...)...)
	if err != nil {
		return nil, grpcstatus.Errorf(codes.FailedPrecondition, "Could not connect to Eval Manager: %v", err)
	}

	// Generate new Request ID
	requestId := uuid.NewString()

	// Handle the actual request in a separate goroutine
	// StartCollecting will return and later collection problems are reported to the evaluation manager

	go handleCollectionRequest(req.ServiceId, evidenceStream, config)
	return &collection.StartCollectingResponse{Id: requestId}, nil
}

func (s *Server) StopCollecting(_ context.Context, _ *collection.StopCollectingRequest) (*emptypb.Empty, error) {
	return nil, grpcstatus.Error(codes.Unimplemented, "StopCollection not implemented")
}

// collectOAuth2Evidence collects evidence about an OAuth 2.0 authorization server, such as metadata, grant types, and
// such.
func collectOAuth2Evidence(serviceID string, config *collection.AuthenticationSecurityConfig) (evidence *common.Evidence, err error) {
	log.Infof("Collecing evidence for OAuth 2.0 authentication %s", config.Issuer)

	evidenceId := uuid.NewString()

	// Prepare evidence struct
	evidence = &common.Evidence{
		Id:             evidenceId,
		Name:           evidenceId,
		TargetService:  serviceID,
		TargetResource: config.Issuer,
		GatheredAt:     timestamppb.Now(),
		ToolId:         ComponentID,
	}

	metadata, errStruct := getAndValidateMetadata(config)
	if errStruct != nil {
		evidence.Error = errStruct
		return
	}

	rawEvidence, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling raw evidence: %w", err)
	}

	evidence.RawEvidence = string(rawEvidence)

	// Ignore Errors. Values are already checked
	grantTypes, _ := shouldBeStringArrayOrNil(metadata, "grant_types_supported")
	iDTokenSigningAlgValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "id_token_signing_alg_values_supported")
	userinfoSigningAlgValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "userinfo_signing_alg_values_supported")
	requestObjectSigningAlgValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "request_object_signing_alg_values_supported")
	tokenEndpointAuthSigningAlgValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "token_endpoint_auth_signing_alg_values_supported")
	revocationEndpointAuthSigningAlgValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "revocation_endpoint_auth_signing_alg_values_supported")
	introspectionEndpointAuthSigningAlgValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "introspection_endpoint_auth_signing_alg_values_supported")
	iDTokenEncryptionAlgValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "id_token_encryption_alg_values_supported")
	iDTokenEncryptionEncValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "id_token_encryption_enc_values_supported")
	userinfoEncryptionAlgValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "userinfo_encryption_alg_values_supported")
	userinfoEncryptionEncValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "userinfo_encryption_enc_values_supported")
	requestObjectEncryptionAlgValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "request_object_encryption_alg_values_supported")
	requestObjectEncryptionEncValuesSupported, _ := shouldBeStringArrayOrNil(metadata, "request_object_encryption_enc_values_supported")

	evidenceValue := Value{
		Resource: voc.Resource{
			// ID and Type has to be set. Otherwise, evaluation will fail due to evidence validation
			ID:   voc.ResourceID(config.Issuer),
			Type: []string{"OAuthGrantTypes"},
		},
		OAuthGrantTypes: &OAuthGrantTypes{
			GrantTypes:                                         grantTypes,
			IDTokenSigningAlgValuesSupported:                   iDTokenSigningAlgValuesSupported,
			UserinfoSigningAlgValuesSupported:                  userinfoSigningAlgValuesSupported,
			RequestObjectSigningAlgValuesSupported:             requestObjectSigningAlgValuesSupported,
			TokenEndpointAuthSigningAlgValuesSupported:         tokenEndpointAuthSigningAlgValuesSupported,
			RevocationEndpointAuthSigningAlgValuesSupported:    revocationEndpointAuthSigningAlgValuesSupported,
			IntrospectionEndpointAuthSigningAlgValuesSupported: introspectionEndpointAuthSigningAlgValuesSupported,
			IDTokenEncryptionAlgValuesSupported:                iDTokenEncryptionAlgValuesSupported,
			IDTokenEncryptionEncValuesSupported:                iDTokenEncryptionEncValuesSupported,
			UserinfoEncryptionAlgValuesSupported:               userinfoEncryptionAlgValuesSupported,
			UserinfoEncryptionEncValuesSupported:               userinfoEncryptionEncValuesSupported,
			RequestObjectEncryptionAlgValuesSupported:          requestObjectEncryptionAlgValuesSupported,
			RequestObjectEncryptionEncValuesSupported:          requestObjectEncryptionEncValuesSupported,
		},
	}

	evidence.Value, err = protobuf.ToValue(evidenceValue)
	if err != nil {
		return nil, fmt.Errorf("error while converting evidence to protobuf value: %w", err)
	}

	return
}

// collectAPIAccessEvidence collects evidence about an OAuth 2.0 protected API endpoint
func collectAPIAccessEvidence(serviceID string, config *collection.AuthenticationSecurityConfig) (evidence *common.Evidence, err error) {
	log.Infof("Collecing evidence for API access to endpoint %s", config.ApiEndpoint)

	evidenceId := uuid.NewString()

	// Prepare evidence struct
	evidence = &common.Evidence{
		Id:             evidenceId,
		Name:           evidenceId,
		TargetService:  serviceID,
		TargetResource: config.ApiEndpoint,
		GatheredAt:     timestamppb.Now(),
		ToolId:         ComponentID,
	}

	metadata, errStruct := getAndValidateMetadata(config)
	if errStruct != nil {
		evidence.Error = errStruct
		return
	}

	endpointString := config.ApiEndpoint
	clientId := config.ClientId
	clientSecret := config.ClientSecret
	scopes := strings.Fields(config.Scopes)
	endpoint, err := url.Parse(endpointString)
	if err != nil {
		evidence.Error = &common.Error{
			Code:        common.Error_ERROR_INVALID_CONFIGURATION,
			Description: "Error parsing endpoint URL: " + err.Error(),
		}
		// Clear the error, because we want to send an evidence with the error, but not fail
		err = nil
		return
	}

	// TODO(bellebaum): Add small comment
	unprotectedAccess, err := CheckAPIAccess(*endpoint, "GET", nil) // Only GET supported atm, so we needn't worry about request bodies
	if err != nil {
		evidence.Error = &common.Error{
			Code:        common.Error_ERROR_CONNECTION_FAILURE,
			Description: "Cannot connect to service: " + err.Error(),
		}
		// Clear the error, because we want to send an evidence with the error, but not fail
		err = nil
		return
	}

	// TODO(bellebaum): Add small comment
	accessToken, err := acquireAccessToken(metadata, clientId, clientSecret, scopes)
	if err != nil {
		evidence.Error = &common.Error{
			Code:        common.Error_ERROR_CONNECTION_FAILURE,
			Description: "Cannot acquire Token: " + err.Error(),
		}
		// Clear the error, because we want to send an evidence with the error, but not fail
		err = nil
		return
	}

	// TODO(bellebaum): Add small comment
	protectedAccess, err := CheckAPIAccess(*endpoint, "GET", accessToken) // Only GET supported atm, so we needn't worry about request bodies
	if err != nil {
		evidence.Error = &common.Error{
			Code:        common.Error_ERROR_CONNECTION_FAILURE,
			Description: "Cannot connect to service: " + err.Error(),
		}
		// Clear the error, because we want to send an evidence with the error, but not fail
		err = nil
		return
	}

	// Determine Status
	status := ""
	if unprotectedAccess {
		status = "unprotected"
	} else if protectedAccess {
		status = "protected"
	} else {
		status = "no_access"
	}

	// Prepare Evidence
	evidenceValue := Value{
		Resource: voc.Resource{
			// ID and Type has to be set. Otherwise, evaluation will fail due to evidence validation
			ID:   voc.ResourceID(config.ApiEndpoint),
			Type: []string{"APIOAuthProtected"},
		},
		APIOAuthProtected: &APIOAuthProtected{
			Url:    endpointString,
			Status: status,
		},
	}

	evidence.Value, err = protobuf.ToValue(evidenceValue)
	if err != nil {
		return nil, fmt.Errorf("error while converting evidence to protobuf value: %w", err)
	}

	return
}
