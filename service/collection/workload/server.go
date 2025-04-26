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

// Package workload contains service specific code for the Workload Configuration Collection Module.
package workload

import (
	"context"
	"fmt"
	"os"
	"sync"

	clapi "clouditor.io/clouditor/api"
	clapidiscovery "clouditor.io/clouditor/api/discovery"
	"clouditor.io/clouditor/service/discovery/aws"
	"clouditor.io/clouditor/service/discovery/k8s"
	"clouditor.io/clouditor/voc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/client-go/kubernetes"

	"github.com/eclipse-xfsc/cam/api"
	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/collection/workload"
	"github.com/eclipse-xfsc/cam/api/common"
	"github.com/eclipse-xfsc/cam/api/evaluation"
	"github.com/eclipse-xfsc/cam/internal/config"
	"github.com/eclipse-xfsc/cam/internal/protobuf"
	"github.com/eclipse-xfsc/cam/service"
	. "github.com/eclipse-xfsc/cam/service/collection"
	awsstrct "github.com/eclipse-xfsc/cam/service/collection/workload/aws/strct"
	"github.com/eclipse-xfsc/cam/service/collection/workload/openstack"
	openstackstrct "github.com/eclipse-xfsc/cam/service/collection/workload/openstack/strct"
)

var (
	log         = logrus.WithField("service", "collection-workload")
	ComponentID = config.DefaultCollectionWorkloadID
)

// Server is an implementation of the Workload Configuration service.
type Server struct {
	collection.UnimplementedCollectionServer

	// Stream to Evaluation Manager component
	stream *clapi.StreamsOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence]

	grpcOpts []grpc.DialOption

	authorizer clapi.Authorizer

	// Provider Configuration with the Cloud serviceID as key
	providerConfigs map[string]providerConfiguration
	// Mutex for provider configs
	providerConfigsMutex sync.Mutex
}

// providerConfiguration contains the configs for
// * Kubernetes
// * Openstack
type providerConfiguration struct {
	kubernetes *kubernetes.Clientset
	openstack  *openstack.AuthOptions
	aws        *aws.Client
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

// NewServer creates a new workload server with default values.
func NewServer(opts ...service.ServiceOption[Server]) collection.CollectionServer {
	s := &Server{
		stream:          clapi.NewStreamsOf(clapi.WithLogger[evaluation.Evaluation_SendEvidencesClient, *common.Evidence](log)),
		grpcOpts:        []grpc.DialOption{},
		providerConfigs: make(map[string]providerConfiguration),
	}

	// Apply any options
	for _, o := range opts {
		o(s)
	}

	return s
}

// StartCollecting starts collecting configurations from Kubernetes and OpenStack, creates evidences and sends the
// evidences to the Evaluation Manager.
func (srv *Server) StartCollecting(_ context.Context, req *collection.StartCollectingRequest) (
	resp *collection.StartCollectingResponse, err error) {
	var conf collection.WorkloadSecurityConfig

	log.Infof("Received StartCollecting Request for Service ID '%v'", req.ServiceId)

	// Validate StartCollectingRequest
	if err = req.Validate(); err != nil {
		err = status.Error(codes.InvalidArgument, err.Error())
		log.Debug(err)
		return
	}

	// Extract protobuf rawConfiguration to Config type
	err = req.Configuration.RawConfiguration.UnmarshalTo(&conf)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, collection.ErrInvalidWorkloadConfigurationRawConfiguration.Error())
		log.Debug(err)
		return
	}

	// Validate the StartCollectingRequest.RawConfiguration
	// TODO (anatheka): after merging MR #74 (...Validate Method) we must add the CollectionModuleValidator call here
	// err := collection.Validate[Config](req, validator)
	// Convert and store provider configuration
	err = srv.addProviderConfig(req, &conf)
	if err != nil {
		err = fmt.Errorf("could not add provider configuration: %w", err)
		log.Error(err)
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	// Get stream for the Evaluation Manager
	stream, err := srv.stream.GetStream(req.EvalManager, TargetComponent, api.InitEvalStream,
		clapi.DefaultGrpcDialOptions(req.EvalManager, srv, srv.grpcOpts...)...)
	if err != nil {
		err = fmt.Errorf("could not get stream to Evaluation Manager: %w", err)
		log.Debug(err)
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	// Get workload configurations
	results, err := srv.getWorkloadConfigurations(req)
	if err != nil {
		err = fmt.Errorf("could not retrieve configurations: %w", err)
		log.Error(err)
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	// Create CAM evidence and send to stream channel
	err = EnqueueEvidences(ComponentID, req, results, stream, log)
	if err != nil {
		err = fmt.Errorf("could not enqueue CAM evidence in stream channel: %v", err)
		log.Error(err)
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	// Generate a new ID
	resp = &collection.StartCollectingResponse{
		Id: uuid.NewString(),
	}

	return
}

// StopCollecting stops collecting configurations
func (*Server) StopCollecting(_ context.Context, _ *collection.StopCollectingRequest) (*emptypb.Empty, error) {
	// For now we do not implement the StopCollection
	return nil, status.Error(codes.Unimplemented, "StopCollection not implemented")
}

// getWorkloadConfigurations configures discoverers based on the given ServiceID and retrieves resources
func (srv *Server) getWorkloadConfigurations(req *collection.StartCollectingRequest) ([]voc.IsCloudResource, error) {
	var (
		discoverer []clapidiscovery.Discoverer
		results    []voc.IsCloudResource
		err        error
	)

	// Set discoverer for existing provider configurations
	discoverer = srv.setDiscoverer(req.ServiceId)
	if discoverer == nil {
		err = fmt.Errorf("no discoverer available")
		log.Error(err)
		return nil, err
	}

	// Retrieve resources
	for _, v := range discoverer {
		list, err := v.List()
		if err != nil {
			err = fmt.Errorf("could not retrieve resources from %s: %v", v.Name(), err)
			log.Error(err)
		}
		results = append(results, list...)
	}

	return results, nil
}

// addProviderConfig stores the given provider configuration
func (srv *Server) addProviderConfig(req *collection.StartCollectingRequest, conf *collection.WorkloadSecurityConfig) (err error) {
	// If serviceID is already available, return
	if _, ok := srv.providerConfigs[req.ServiceId]; ok {
		return
	}

	if conf == nil {
		return collection.ErrMissingServiceConfiguration
	}

	// Set ServiceConfiguration for Kubernetes
	if k := conf.Kubernetes; !isEmpty(k) {
		err = srv.kubeConfig(k, req.ServiceId)
		if err != nil {
			return fmt.Errorf("%s: %w", collection.ErrInvalidKubernetesServiceConfiguration, err)
		}
	}

	// Set ServiceConfiguration for OpenStack
	if os := conf.Openstack; !isEmpty(os) {
		err := srv.openstackConfig(os, req.ServiceId)
		if err != nil {
			return fmt.Errorf("%s: %w", collection.ErrInvalidOpenstackServiceConfiguration, err)
		}
	}

	// Set ServiceConfiguration for AWS
	if aws := conf.Aws; !isEmpty(aws) {
		err = srv.awsConfig(aws, req.ServiceId)
		if err != nil {
			return fmt.Errorf("%s: %w", collection.ErrInvalidAWSServiceConfiguration, err)
		}
	}

	return nil
}

// kubeConfig stores the Kubernetes configuration
func (srv *Server) kubeConfig(value *structpb.Value, serviceId string) error {
	// Get byte array from protobuf value
	v, err := protobuf.ToByteArray(value)
	if err != nil {
		return collection.ErrConversionProtobufToByteArray
	}

	// Get Kubernetes clientset
	clientset, err := workload.AuthFromKubeConfig(v)
	if err != nil {
		return fmt.Errorf("%s: %w", collection.ErrKubernetesClientset, err)
	}

	// Store Kubernetes clientset for the specific serviceID
	configValue := providerConfiguration{
		kubernetes: clientset,
	}

	srv.providerConfigsMutex.Lock()
	srv.providerConfigs[serviceId] = configValue
	srv.providerConfigsMutex.Unlock()

	return nil
}

// openstackConfig stores the OpenStack configuration
func (srv *Server) openstackConfig(value *structpb.Value, serviceId string) error {

	// Get AuthOpts from protobuf value
	authOpts, err := openstackstrct.ToAuthOptions(value)
	if err != nil {
		return collection.ErrConversionProtobufToAuthOptions
	}

	// Store Openstack AuthOpts to the  specific serviceID
	configValue := providerConfiguration{
		openstack: authOpts,
	}
	srv.providerConfigsMutex.Lock()
	srv.providerConfigs[serviceId] = configValue
	srv.providerConfigsMutex.Unlock()

	return nil

}

// awsConfig stores the AWS configuration
func (srv *Server) awsConfig(value *structpb.Value, serviceId string) error {
	// Get config from protobuf value
	strct, err := awsstrct.ToConfig(value)
	if err != nil {
		return collection.ErrConversionProtobufToAuthOptions
	}
	// This is a little hack but should work. aws.NewClient() gets configured
	// from the environment variables, so we just set it here.
	os.Setenv("AWS_DEFAULT_REGION", strct.Region)
	os.Setenv("AWS_ACCESS_KEY_ID", strct.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", strct.SecretAccessKey)

	client, err := aws.NewClient()
	if err != nil {
		return fmt.Errorf("could not create AWS client: %w", err)
	}

	// Store AWS config to the specific serviceID
	configValue := providerConfiguration{
		aws: client,
	}
	srv.providerConfigsMutex.Lock()
	srv.providerConfigs[serviceId] = configValue
	srv.providerConfigsMutex.Unlock()

	return nil

}

// setDiscoverer sets discoverer for serviceID
func (srv *Server) setDiscoverer(serviceID string) (discoverer []clapidiscovery.Discoverer) {
	// Add Kubernetes discoverer for compute and storage
	if srv.providerConfigs[serviceID].kubernetes != nil {
		discoverer = append(discoverer, k8s.NewKubernetesComputeDiscovery(srv.providerConfigs[serviceID].kubernetes), k8s.NewKubernetesNetworkDiscovery(srv.providerConfigs[serviceID].kubernetes))
	}

	// Add Openstack discoverer for compute and storage
	if srv.providerConfigs[serviceID].openstack != nil {
		discoverer = append(discoverer, openstack.NewStorageDiscovery(openstack.WithAuthOpts(srv.providerConfigs[serviceID].openstack)), openstack.NewComputeDiscovery(openstack.WithAuthOpts(srv.providerConfigs[serviceID].openstack)))
	}

	// Add Openstack discoverer for compute and storage
	if srv.providerConfigs[serviceID].aws != nil {
		discoverer = append(discoverer, aws.NewAwsStorageDiscovery(srv.providerConfigs[serviceID].aws))
	}

	return
}

func isEmpty(value *structpb.Value) bool {
	// If value is nil, its definitly empty
	if value == nil {
		return true
	}

	// Check, if it is an empty map, which means that its not configured. A little bit of a nasty workaround
	strct := value.GetStructValue()
	if strct != nil && len(strct.Fields) == 0 {
		return true
	}

	return false
}

// Not used now
//func validator(req *collection.StartCollectingRequest) (conf *collection.ServiceConfiguration_WorkloadSecurityConfig, err error) {
//	if req.Configuration.CollectionModule != collection.ServiceConfiguration_WORKLOAD_CONFIGURATION {
//		err = collection.ErrInvalidCollectionModule
//		return
//	}
//
//	// TODO(anatheka): Use addProvider logic here?
//	//conf, err = strct.ToStruct[collection.WorkloadConfig](req.Configuration.GetRawConfiguration())
//	//if err != nil {
//	//	err = fmt.Errorf("invalid configuration: raw configuration is not of type 'Config' ('kubernetes' or 'openstack'): %s", err)
//	//	return
//	//}
//
//	return
//}
