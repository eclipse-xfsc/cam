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
	"reflect"
	"testing"

	"clouditor.io/clouditor/api/orchestrator"
	"clouditor.io/clouditor/persistence"
	orchestrator2 "clouditor.io/clouditor/service/orchestrator"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"

	"github.com/eclipse-xfsc/cam/api"
	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/configuration"
	"github.com/eclipse-xfsc/cam/internal/testutil"
	"github.com/eclipse-xfsc/cam/internal/testutil/testproto"
)

func Test_server_ListCloudServiceConfigurations(t *testing.T) {
	config1 := &collection.ServiceConfiguration{
		ServiceId:        orchestrator2.DefaultTargetCloudServiceId,
		RawConfiguration: testproto.NewAny(t, &collection.AuthenticationSecurityConfig{Issuer: "Im an Issuer"}),
	}
	config1.TypeUrl = config1.RawConfiguration.TypeUrl

	config2 := &collection.ServiceConfiguration{
		ServiceId: orchestrator2.DefaultTargetCloudServiceId,
		RawConfiguration: testproto.NewAny(t, &collection.WorkloadSecurityConfig{
			Openstack: &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "Im Openstack"}},
		}),
	}
	config2.TypeUrl = config2.RawConfiguration.TypeUrl

	type fields struct {
		storage persistence.Storage
	}
	type args struct {
		in0 context.Context
		req *configuration.ListCloudServiceConfigurationsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.ValueAssertionFunc
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Correct",
			fields: fields{
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
					assert.NoError(t, s.Create(&orchestrator.CloudService{Id: orchestrator2.DefaultTargetCloudServiceId}))
					assert.NoError(t, s.Create(config1))
					assert.NoError(t, s.Create(config2))
				}),
			},
			args: args{
				req: &configuration.ListCloudServiceConfigurationsRequest{ServiceId: orchestrator2.DefaultTargetCloudServiceId},
			},
			wantRes: func(t assert.TestingT, i interface{}, _ ...interface{}) bool {
				got := i.(*configuration.ListCloudServiceConfigurationsResponse)
				assert.Len(t, got.Configurations, 2)
				assert.True(t, containsConfig(got.Configurations, config1))
				assert.True(t, containsConfig(got.Configurations, config2))
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "Empty list of configuration",
			fields: fields{
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
					err := s.Create(&orchestrator.CloudService{Id: orchestrator2.DefaultTargetCloudServiceId})
					assert.NoError(t, err)
				}),
			},
			args: args{
				req: &configuration.ListCloudServiceConfigurationsRequest{ServiceId: orchestrator2.DefaultTargetCloudServiceId},
			},
			wantRes: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				want := &configuration.ListCloudServiceConfigurationsResponse{
					Configurations: []*collection.ServiceConfiguration{},
				}
				got := i.(*configuration.ListCloudServiceConfigurationsResponse)
				return assert.Equal(t, want, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "DB Error",
			fields: fields{
				storage: testutil.NewInMemoryStorageWithListError(t, gorm.ErrInvalidData.Error(), func(s persistence.Storage) {
					assert.NoError(t, s.Create(&orchestrator.CloudService{Id: orchestrator2.DefaultTargetCloudServiceId}))

				}),
			},
			args: args{
				req: &configuration.ListCloudServiceConfigurationsRequest{ServiceId: orchestrator2.DefaultTargetCloudServiceId},
			},
			wantRes: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				want := new(configuration.ListCloudServiceConfigurationsResponse)
				got := i.(*configuration.ListCloudServiceConfigurationsResponse)
				return assert.Equal(t, want, got)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.ErrorContains(t, err, api.DatabaseErrorMsg)
				return assert.ErrorContains(t, err, gorm.ErrInvalidData.Error())
			},
		},
		{
			name: "Service does not exist yet (Verification)",
			fields: fields{
				storage: testutil.NewInMemoryStorage(t),
			},
			args: args{
				req: &configuration.ListCloudServiceConfigurationsRequest{ServiceId: orchestrator2.DefaultTargetCloudServiceId},
			},
			wantRes: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				want := new(configuration.ListCloudServiceConfigurationsResponse)
				got := i.(*configuration.ListCloudServiceConfigurationsResponse)
				return assert.Equal(t, want, got)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "No Cloud Service found")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage: tt.fields.storage,
			}
			gotRes, err := srv.ListCloudServiceConfigurations(tt.args.in0, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("ListCloudServiceConfigurations(%v, %v)", tt.args.in0, tt.args.req)) {
				return
			}
			tt.wantRes(t, gotRes)
		})
	}
}

func containsConfig(configurations []*collection.ServiceConfiguration, wantConfig *collection.ServiceConfiguration) bool {
	for _, c := range configurations {
		if reflect.DeepEqual(c, wantConfig) {
			return true
		}
	}
	return false
}

func TestServer_ConfigureCloudService(t *testing.T) {
	type fields struct {
		storage        persistence.Storage
		configurations map[string]map[string]*collection.ServiceConfiguration
	}
	type args struct {
		ctx context.Context
		req *configuration.ConfigureCloudServiceRequest
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantRes     *configuration.ConfigureCloudServiceResponse
		wantStorage assert.ValueAssertionFunc
		wantErr     assert.ErrorAssertionFunc
	}{
		{
			name: "Correct",
			fields: fields{
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
					assert.NoError(t, s.Create(&orchestrator.CloudService{Id: orchestrator2.DefaultTargetCloudServiceId}))

				}),
				configurations: make(map[string]map[string]*collection.ServiceConfiguration),
			},
			args: args{
				req: &configuration.ConfigureCloudServiceRequest{
					ServiceId: orchestrator2.DefaultTargetCloudServiceId,
					Configurations: &configuration.Configurations{
						Configurations: []*collection.ServiceConfiguration{
							{
								ServiceId: orchestrator2.DefaultTargetCloudServiceId,
								RawConfiguration: testproto.NewAny(t,
									&collection.AuthenticationSecurityConfig{
										Issuer: "someURL",
									}),
							},
							{
								ServiceId: orchestrator2.DefaultTargetCloudServiceId,
								RawConfiguration: testproto.NewAny(t,
									&collection.CommunicationSecurityConfig{
										Endpoint: "host:443",
									}),
							},
						}},
				},
			},
			wantRes: &configuration.ConfigureCloudServiceResponse{},
			wantStorage: func(testingT assert.TestingT, i interface{}, i2 ...interface{}) bool {
				storage, ok := i.(persistence.Storage)
				assert.True(testingT, ok)

				var configs []*collection.ServiceConfiguration

				assert.NoError(testingT, storage.List(&configs, "", true, 0, -1,
					"service_id = ?", orchestrator2.DefaultTargetCloudServiceId))
				assert.Len(testingT, configs, 2)

				var config *collection.ServiceConfiguration
				assert.NoError(testingT, storage.Get(&config, "service_id = ? AND type_url = ?",
					orchestrator2.DefaultTargetCloudServiceId, testproto.NewAny(t, &collection.AuthenticationSecurityConfig{}).TypeUrl))

				authConfig := new(collection.AuthenticationSecurityConfig)
				assert.Nil(testingT, config.RawConfiguration.UnmarshalTo(authConfig))

				return assert.Equal(testingT, "someURL", authConfig.Issuer)
			},
			wantErr: assert.NoError,
		},
		{
			name:        "Missing service id",
			fields:      fields{storage: testutil.NewInMemoryStorage(t)},
			args:        args{req: &configuration.ConfigureCloudServiceRequest{ServiceId: ""}},
			wantRes:     &configuration.ConfigureCloudServiceResponse{},
			wantStorage: assert.NotNil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, api.ServiceIDIsMissingErrMsg)
			},
		},
		{
			name: "Service does not exist yet (Verification)",
			fields: fields{
				storage: testutil.NewInMemoryStorage(t),
			},
			args:        args{req: &configuration.ConfigureCloudServiceRequest{ServiceId: "0123"}},
			wantRes:     &configuration.ConfigureCloudServiceResponse{},
			wantStorage: assert.NotNil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "No Cloud Service found")
			},
		},
		{
			name: "DB error",
			fields: fields{
				storage: testutil.NewInMemoryStorageWithSaveError(t, gorm.ErrInvalidTransaction.Error(), func(s persistence.Storage) {
					assert.NoError(t, s.Create(&orchestrator.CloudService{Id: orchestrator2.DefaultTargetCloudServiceId}))
				}),
			},
			args: args{
				req: &configuration.ConfigureCloudServiceRequest{
					ServiceId: orchestrator2.DefaultTargetCloudServiceId,
					Configurations: &configuration.Configurations{
						Configurations: []*collection.ServiceConfiguration{
							{
								ServiceId: orchestrator2.DefaultTargetCloudServiceId,
								TypeUrl:   "no.such.type",
								RawConfiguration: testproto.NewAny(t,
									&collection.AuthenticationSecurityConfig{Issuer: "someURL"}),
							},
						}}}},
			wantRes:     &configuration.ConfigureCloudServiceResponse{},
			wantStorage: assert.NotNil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, gorm.ErrInvalidTransaction.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				storage: tt.fields.storage,
			}
			gotRes, err := s.ConfigureCloudService(tt.args.ctx, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("ConfigureCloudService(%v, %v)", tt.args.ctx, tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.wantRes, gotRes, "ConfigureCloudService(%v, %v)", tt.args.ctx, tt.args.req)
			tt.wantStorage(t, s.storage)
		})
	}
}
