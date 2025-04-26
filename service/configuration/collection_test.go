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
	"fmt"
	"testing"

	"clouditor.io/clouditor/api"
	"clouditor.io/clouditor/api/assessment"
	"clouditor.io/clouditor/api/orchestrator"
	"clouditor.io/clouditor/persistence"
	service_orchestrator "clouditor.io/clouditor/service/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/configuration"
	"github.com/eclipse-xfsc/cam/internal/protobuf"
	"github.com/eclipse-xfsc/cam/internal/testutil"
	"github.com/eclipse-xfsc/cam/internal/testutil/testproto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

// TODO(lebogg): Add test for DB failure
func Test_server_ListCollectionModules(t *testing.T) {
	type fields struct {
		storage persistence.Storage
	}
	type args struct {
		in0 context.Context
		in1 *configuration.ListCollectionModulesRequest
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantModuleIDs []string
		wantErr       assert.ErrorAssertionFunc
	}{
		{
			name: "Correct execution (2 CMs)",
			fields: fields{
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
					_ = s.Save(&collection.CollectionModule{Id: "CM1"})
					_ = s.Save(&collection.CollectionModule{Id: "CM2"})
				}),
			},
			wantModuleIDs: []string{"CM1", "CM2"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "Correct execution (no CM)",
			fields: fields{
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {}),
			},
			wantModuleIDs: []string(nil),
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name: "DB error",
			fields: fields{
				storage: &testutil.StorageWithError{ListErr: gorm.ErrInvalidData},
			},
			wantModuleIDs: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) (isCorrect bool) {
				isCorrect = true
				if !assert.Equal(t, codes.Internal, status.Code(err), "Wrong status code") {
					isCorrect = false
				}
				if !assert.ErrorContains(t, err, gorm.ErrInvalidData.Error()) {
					isCorrect = false
				}
				return
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage: tt.fields.storage,
			}
			gotRes, err := srv.ListCollectionModules(tt.args.in0, tt.args.in1)
			if !tt.wantErr(t, err, fmt.Sprintf("ListCollectionModules(%v, %v)", tt.args.in0, tt.args.in1)) {
				return
			}
			// Assert that IDs of CMs are equal
			var moduleIDs []string
			for _, module := range gotRes.Modules {
				moduleIDs = append(moduleIDs, module.Id)
			}
			assert.Equalf(t, tt.wantModuleIDs, moduleIDs, "ListCollectionModules(%v, %v)", tt.args.in0, tt.args.in1)
		})
	}
}

func Test_server_RemoveCollectionModule(t *testing.T) {
	type fields struct {
		storage persistence.Storage
	}
	type args struct {
		in0 context.Context
		req *configuration.RemoveCollectionModuleRequest
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantNumberOfCMs int
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name: "Correct execution",
			fields: fields{storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
				_ = s.Save(&collection.CollectionModule{Id: "CM1"})
				_ = s.Save(&collection.CollectionModule{Id: "CM2"})
			})},
			args:            args{req: &configuration.RemoveCollectionModuleRequest{ModuleId: "CM1"}},
			wantNumberOfCMs: 1,
			wantErr:         assert.NoError,
		},
		{
			name: "CM not found",
			fields: fields{storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
				_ = s.Save(&collection.CollectionModule{Id: "CM1"})
			})},
			args:            args{req: &configuration.RemoveCollectionModuleRequest{ModuleId: "CM2"}},
			wantNumberOfCMs: 1,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, codes.NotFound, status.Code(err))
				return assert.ErrorContains(t, err, persistence.ErrRecordNotFound.Error())
			},
		},
		{
			name: "DB error",
			fields: fields{
				storage: &testutil.StorageWithError{DeleteErr: gorm.ErrInvalidData},
			},
			args:            args{req: &configuration.RemoveCollectionModuleRequest{ModuleId: "CM1"}},
			wantNumberOfCMs: 0,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, codes.Internal, status.Code(err))
				return assert.ErrorContains(t, err, gorm.ErrInvalidData.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage: tt.fields.storage,
			}
			_, err := srv.RemoveCollectionModule(tt.args.in0, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("RemoveCollectionModule(%v, %v)", tt.args.in0, tt.args.req)) {
				return
			}
			gotNumberOfCMs, err := tt.fields.storage.Count(&collection.CollectionModule{})
			if err != nil {
				return
			}
			assert.Equalf(t, tt.wantNumberOfCMs, int(gotNumberOfCMs), "RemoveCollectionModule(%v, %v)", tt.args.in0, tt.args.req)
		})
	}
}

func TestServer_GetServiceConfiguration(t *testing.T) {
	storage := testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
		err := populateStorage(t, s)
		if err != nil {
			t.Error(err)
		}
	})

	type fields struct {
		OrchestratorServer orchestrator.OrchestratorServer
		evalManagerAddress string
		interval           int
		storage            persistence.Storage
		authorizer         api.Authorizer
		monitoring         map[string]*MonitorScheduler
	}
	type args struct {
		serviceID string
		typeURL   string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantConfig *collection.ServiceConfiguration
	}{
		{
			name: "Found config",
			fields: fields{
				OrchestratorServer: service_orchestrator.NewService(service_orchestrator.WithStorage(storage)),
				storage:            storage,
				monitoring:         make(map[string]*MonitorScheduler),
			},
			args: args{
				serviceID: "myService",
				typeURL:   protobuf.TypeURL(&collection.AuthenticationSecurityConfig{}),
			},
			wantConfig: &collection.ServiceConfiguration{
				ServiceId: "myService",
				TypeUrl:   protobuf.TypeURL(&collection.AuthenticationSecurityConfig{}),
				RawConfiguration: testproto.NewAny(t, &collection.AuthenticationSecurityConfig{
					Issuer: "myissuer",
				}),
			},
		},
		{
			name: "Empty config",
			fields: fields{
				OrchestratorServer: service_orchestrator.NewService(service_orchestrator.WithStorage(storage)),
				storage:            storage,
				monitoring:         make(map[string]*MonitorScheduler),
			},
			args: args{
				serviceID: "myService",
				typeURL:   protobuf.TypeURL(&collection.WorkloadSecurityConfig{}),
			},
			wantConfig: &collection.ServiceConfiguration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				OrchestratorServer: tt.fields.OrchestratorServer,
				evalManagerAddress: tt.fields.evalManagerAddress,
				interval:           tt.fields.interval,
				storage:            tt.fields.storage,
				authorizer:         tt.fields.authorizer,
				monitoring:         tt.fields.monitoring,
			}
			if gotConfig := srv.GetServiceConfiguration(tt.args.serviceID, tt.args.typeURL); !proto.Equal(gotConfig, tt.wantConfig) {
				t.Errorf("Server.GetServiceConfiguration() = %v, want %v", gotConfig, tt.wantConfig)
			}
		})
	}
}

func populateStorage(t *testing.T, storage persistence.Storage) (err error) {
	// Populate some static metric data
	err = storage.Create(&orchestrator.Requirement{Id: "Req-1", Metrics: []*assessment.Metric{
		{Id: "Metric-1", Interval: 60},
		{Id: "Metric-2"},
	}})
	if err != nil {
		return err
	}

	// Populate DB with some collection modules
	err = storage.Create(&collection.CollectionModule{
		Id: "Module-1", Metrics: []*assessment.Metric{
			{Id: "Metric-1"},
		},
		ConfigMessageTypeUrl: protobuf.TypeURL(&collection.WorkloadSecurityConfig{}),
	})
	if err != nil {
		return err
	}

	err = storage.Create(&collection.CollectionModule{
		Id: "Module-2", Metrics: []*assessment.Metric{
			{Id: "Metric-2"},
		},
		ConfigMessageTypeUrl: protobuf.TypeURL(&collection.AuthenticationSecurityConfig{}),
	})
	if err != nil {
		return err
	}

	// Populate DB with some service configurations
	err = storage.Save(&collection.ServiceConfiguration{
		ServiceId: "myService",
		TypeUrl:   protobuf.TypeURL(&collection.AuthenticationSecurityConfig{}),
		RawConfiguration: testproto.NewAny(t, &collection.AuthenticationSecurityConfig{
			Issuer: "myissuer",
		}),
	})
	if err != nil {
		return err
	}

	return
}
