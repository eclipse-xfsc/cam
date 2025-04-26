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

package configuration

import (
	"context"

	"clouditor.io/clouditor/api/orchestrator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/eclipse-xfsc/cam/api"
	"github.com/eclipse-xfsc/cam/api/configuration"
	"github.com/eclipse-xfsc/cam/internal/storage"
)

// RegisterCloudService is a wrapper around Clouditor orchestrator
func (s *Server) RegisterCloudService(ctx context.Context, req *orchestrator.RegisterCloudServiceRequest) (*orchestrator.CloudService, error) {
	return s.OrchestratorServer.RegisterCloudService(ctx, req)
}

// UpdateCloudService is a wrapper around Clouditor orchestrator
func (s *Server) UpdateCloudService(ctx context.Context, req *orchestrator.UpdateCloudServiceRequest) (*orchestrator.CloudService, error) {
	return s.OrchestratorServer.UpdateCloudService(ctx, req)
}

// GetCloudService is a wrapper around Clouditor orchestrator
func (s *Server) GetCloudService(ctx context.Context, req *orchestrator.GetCloudServiceRequest) (*orchestrator.CloudService, error) {
	return s.OrchestratorServer.GetCloudService(ctx, req)
}

// ListCloudServices is a wrapper around Clouditor orchestrator
func (s *Server) ListCloudServices(ctx context.Context, req *orchestrator.ListCloudServicesRequest) (*orchestrator.ListCloudServicesResponse, error) {
	return s.OrchestratorServer.ListCloudServices(ctx, req)
}

// RemoveCloudService is a wrapper around Clouditor orchestrator
func (s *Server) RemoveCloudService(ctx context.Context, req *orchestrator.RemoveCloudServiceRequest) (*emptypb.Empty, error) {
	return s.OrchestratorServer.RemoveCloudService(ctx, req)
}

// ConfigureCloudService configures the cloud service with the corresponding service configuration
func (s *Server) ConfigureCloudService(_ context.Context, req *configuration.ConfigureCloudServiceRequest) (res *configuration.ConfigureCloudServiceResponse, err error) {
	res = new(configuration.ConfigureCloudServiceResponse)
	// TODO(lebogg): Outsource to Validate fct and extend it
	if req.ServiceId == "" {
		err = status.Error(codes.InvalidArgument, api.ServiceIDIsMissingErrMsg)
		return
	}

	// Verify that service exists already
	err = storage.VerifyExistence(s.storage, "Cloud Service", &orchestrator.CloudService{}, "id", req.ServiceId)
	if err != nil {
		return
	}

	// Store the cloud service configuration
	for _, c := range req.Configurations.Configurations {
		if c.RawConfiguration == nil {
			continue
		}

		typeURL := c.RawConfiguration.TypeUrl

		// Set service ID and type module name, both act as a composite primary key
		c.ServiceId = req.ServiceId
		c.TypeUrl = c.RawConfiguration.TypeUrl

		err = s.storage.Save(c, "service_id = ?", req.ServiceId)
		if err != nil {
			err = status.Errorf(codes.Internal, "%s: could not store configuration for CM '%s': %v",
				api.DatabaseErrorMsg, typeURL, err)
		}
	}
	return
}

// ListCloudServiceConfigurations list all configurations for a service
func (srv *Server) ListCloudServiceConfigurations(_ context.Context, req *configuration.ListCloudServiceConfigurationsRequest) (
	res *configuration.ListCloudServiceConfigurationsResponse, err error) {
	res = new(configuration.ListCloudServiceConfigurationsResponse)
	// Verify that the Cloud Service exists
	err = storage.VerifyExistence(srv.storage, "Cloud Service", &orchestrator.CloudService{}, "id", req.ServiceId)
	if err != nil {
		return
	}

	// List all configurations
	err = srv.storage.List(&res.Configurations, "", true, 0, -1, "service_id = ?",
		req.ServiceId)
	if err != nil {
		err = status.Errorf(codes.Internal, "%s: %v", api.DatabaseErrorMsg, err)
		return
	}

	return
}
