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

package collection_test

import (
	"fmt"
	"testing"

	"clouditor.io/clouditor/api/assessment"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/internal/testutil"
	"github.com/eclipse-xfsc/cam/internal/testutil/testproto"
)

var (
	MockCollectionModuleID = uuid.NewString()
	MockMetricID           = "SomeMetricID"
)

func Test_StartCollectingRequest_validate(t *testing.T) {
	type args struct {
		req *collection.StartCollectingRequest
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantResp assert.ValueAssertionFunc
	}{
		{
			name: "empty service id",
			args: args{
				req: &collection.StartCollectingRequest{
					ServiceId: "",
				},
			},
			wantResp: func(tt assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				err, _ := i1.(error)

				return assert.ErrorContains(t, err, collection.ErrMissingServiceID.Error())
			},
			wantErr: true,
		},
		{
			name: "wrong service id",
			args: args{
				req: &collection.StartCollectingRequest{
					ServiceId: "wrongServiceID",
				},
			},
			wantResp: func(tt assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				err, _ := i1.(error)

				return assert.ErrorContains(t, err, collection.ErrInvalidServiceID.Error())
			},
			wantErr: true,
		},
		{
			name: "evalManager missing",
			args: args{
				req: &collection.StartCollectingRequest{
					ServiceId:   "00000000-0000-0000-0000-000000000000",
					EvalManager: "",
				},
			},
			wantResp: func(tt assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				err, _ := i1.(error)

				return assert.Equal(t, err, collection.ErrMissingEvalManager)
			},
			wantErr: true,
		},
		{
			name: "configuration serviceId missing",
			args: args{
				req: &collection.StartCollectingRequest{
					ServiceId:   "00000000-0000-0000-0000-000000000000",
					EvalManager: "some-evalManager",
					Configuration: &collection.ServiceConfiguration{
						ServiceId:        "",
						RawConfiguration: testproto.NewAny(t, &collection.WorkloadSecurityConfig{}),
					},
				},
			},
			wantResp: func(tt assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				err, _ := i1.(error)
				return assert.ErrorContains(t, err, collection.ErrMissingServiceConfigurationServiceID.Error())
			},
			wantErr: true,
		},
		{
			name: "raw configuration missing",
			args: args{
				req: &collection.StartCollectingRequest{
					ServiceId:   "00000000-0000-0000-0000-000000000000",
					EvalManager: "some-evalManager",
					Configuration: &collection.ServiceConfiguration{
						ServiceId: "00000000-0000-0000-0000-000000000000",
					},
				},
			},
			wantResp: func(tt assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				err, _ := i1.(error)

				return assert.ErrorContains(t, err, collection.ErrMissingServiceConfiguration.Error())
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if err = tt.args.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantResp != nil {
				tt.wantResp(t, err)
			}
		})
	}
}

func TestStartCollectingRequest_KubeConfig(t *testing.T) {
	type fields struct {
		ServiceId     string
		MetricId      string
		EvalManager   string
		Configuration *collection.ServiceConfiguration
	}
	tests := []struct {
		name       string
		fields     fields
		wantConfig []byte
		wantErr    assert.ErrorAssertionFunc
	}{
		{
			name: "Empty configuration",
			fields: fields{
				ServiceId:     "00000000-0000-0000-0000-000000000000",
				Configuration: &collection.ServiceConfiguration{},
			},
			wantConfig: nil,
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				return assert.Contains(t, err.Error(), "configuration not available:")
			},
		},
		{
			name: "Wrong configuration type",
			fields: fields{
				ServiceId: "00000000-0000-0000-0000-000000000000",
				Configuration: &collection.ServiceConfiguration{
					RawConfiguration: testproto.NewAny(t, &collection.AuthenticationSecurityConfig{}),
				},
			},
			wantConfig: nil,
			wantErr: func(tt assert.TestingT, err error, i2 ...interface{}) bool {
				return assert.ErrorIs(t, err, collection.ErrInvalidWorkloadConfigurationRawConfiguration)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &collection.StartCollectingRequest{
				ServiceId:     tt.fields.ServiceId,
				EvalManager:   tt.fields.EvalManager,
				Configuration: tt.fields.Configuration,
			}
			got, err := req.KubeConfig()
			assert.Equal(t, got, tt.wantConfig)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_PersistCollectionModule(t *testing.T) {
	var (
		err     error
		cm2     *collection.CollectionModule
		metric2 *assessment.Metric
	)

	storage := testutil.NewInMemoryStorage(t)

	metric := &assessment.Metric{
		Id:          MockMetricID,
		Name:        "SomeMetricName",
		Description: "SomeMetricDescription",
		Category:    "SomeMetricCategory",
	}

	cm := &collection.CollectionModule{
		Id:          MockCollectionModuleID,
		Name:        "Test Collection Module",
		Description: "Some Description",
		Metrics:     []*assessment.Metric{metric},
		Address:     "",
	}

	err = storage.Create(&cm)
	assert.NoError(t, err)

	// Assert that CM was stored successfully (metric included)
	err = storage.Get(&cm2, "id = ?", MockCollectionModuleID)
	assert.NoError(t, err)
	fmt.Println(cm2)
	assert.True(t, proto.Equal(cm, cm2))

	// Assert that metric was stored separately as well (= metric association was upserted)
	err = storage.Get(&metric2, "id = ?", MockMetricID)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(metric, metric2))
}
