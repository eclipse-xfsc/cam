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
	"errors"

	"clouditor.io/clouditor/api"
	"clouditor.io/clouditor/persistence"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/configuration"
)

// AddCollectionModule adds a collection module to the server
func (srv *Server) AddCollectionModule(_ context.Context, req *configuration.AddCollectionModuleRequest) (
	res *collection.CollectionModule, err error) {

	// Just save it, I don't care
	err = srv.storage.Create(req.Module)
	if err != nil {
		err = status.Errorf(codes.Internal, "DB error: %v", err)
		return
	}

	res = req.Module

	return
}

// ListCollectionModules returns all collection modules
func (srv *Server) ListCollectionModules(_ context.Context, _ *configuration.ListCollectionModulesRequest) (
	res *configuration.ListCollectionModulesResponse, err error) {
	res = new(configuration.ListCollectionModulesResponse)
	err = srv.storage.List(&res.Modules, "", true, 0, -1)
	if err != nil {
		err = status.Errorf(codes.Internal, "DB error: %v", err)
	}
	return
}

// RemoveCollectionModule removes the collection module with the ID specified in the request
func (srv *Server) RemoveCollectionModule(_ context.Context, req *configuration.RemoveCollectionModuleRequest) (
	res *emptypb.Empty, err error) {
	err = srv.storage.Delete(&collection.CollectionModule{}, "id = ?", req.ModuleId)
	// Catch error when no CM is found
	if errors.Is(err, persistence.ErrRecordNotFound) {
		err = status.Errorf(codes.NotFound, "Could not delete CM: %v", err.Error())
		return
	}
	// Catch other DB errors
	if err != nil {
		err = status.Errorf(codes.Internal, "DB error: %v", err)
		return
	}
	return
}

// GetServiceConfiguration return the service configuration for the corresponding serviceID. It does so by comparing the
// field ConfigMessageTypeUrl against the type_url field of a service configuration in the database. If there is no
// config, return empty config (no error)
func (srv *Server) GetServiceConfiguration(serviceID string, typeURL string) (
	config *collection.ServiceConfiguration) {
	config = new(collection.ServiceConfiguration)

	// Retrieve matching service configuration for the collection module
	err := srv.storage.Get(&config, "service_id = ? AND type_url = ?", serviceID, typeURL)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return
	} else if err != nil {
		log.Errorf("database error: %v", err)
	}

	return
}

// startCollectionModule triggers the collection of the given collection module
func (srv *Server) startCollectionModule(cm *collection.CollectionModule, serviceID string) {
	log.Infof("Triggering collection of evidences using module `%s`", cm.Name)

	// Create connection to the collection module. This will make use of our authorizer if it is configured in
	// api.DefaultGrpcDialOptions, because Server is implementing the UsesAuthorizer interface
	conn, err := grpc.Dial(cm.Address, api.DefaultGrpcDialOptions(cm.Address, srv)...)
	if err != nil {
		log.Errorf("Could not dial to `%s`: %v", cm.Name, err)
		return
	}

	// TODO(oxisto): Re-use collection client or use streaming, instead of creating a new client for each request
	collectionClient := collection.NewCollectionClient(conn)
	_, err = collectionClient.StartCollecting(context.TODO(), &collection.StartCollectingRequest{
		ServiceId:     serviceID,
		EvalManager:   srv.evalManagerAddress,
		Configuration: srv.GetServiceConfiguration(serviceID, cm.ConfigMessageTypeUrl),
	})
	if err != nil {
		log.Errorf("Could not start collection module `%s`: %v", cm.Name, err)
	}
}
