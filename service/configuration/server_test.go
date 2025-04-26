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
	"testing"

	"clouditor.io/clouditor/api/assessment"
	"clouditor.io/clouditor/persistence"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/eclipse-xfsc/cam/internal/testutil"
	"github.com/eclipse-xfsc/cam/service"
)

func TestNewServer(t *testing.T) {
	type args struct {
		opts []service.ServiceOption[Server]
	}
	tests := []struct {
		name string
		args args
		want assert.ComparisonAssertionFunc
	}{
		{
			name: "Without options",
			args: args{
				[]service.ServiceOption[Server]{},
			},
			want: func(t assert.TestingT, i interface{}, i2 interface{}, i3 ...interface{}) bool {
				srv, ok := i.(*Server)
				assert.True(t, ok)
				assert.Equal(t, "", srv.evalManagerAddress)
				assert.True(t, ok)
				assert.NotNil(t, srv.OrchestratorServer)
				assert.NotNil(t, srv.monitoring)
				assert.Equal(t, DefaultInterval, srv.interval)
				return true
			},
		},
		{
			name: "Set evaluation address",
			args: args{
				[]service.ServiceOption[Server]{
					WithEvalManagerAddress("someAddress:9090"),
				},
			},
			want: func(t assert.TestingT, i interface{}, i2 interface{}, i3 ...interface{}) bool {
				srv, oK := i.(*Server)
				assert.True(t, oK)
				return assert.Equal(t, "someAddress:9090", srv.evalManagerAddress)
			},
		},
		{
			name: "Set interval",
			args: args{
				[]service.ServiceOption[Server]{
					WithInterval(10),
				},
			},
			want: func(t assert.TestingT, i interface{}, i2 interface{}, i3 ...interface{}) bool {
				srv, oK := i.(*Server)
				assert.True(t, oK)
				return assert.Equal(t, 10, srv.interval)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServer(tt.args.opts...)
			tt.want(t, got, nil)
		})
	}
}

func Test_server_getInterval(t *testing.T) {
	type fields struct {
		interval int
		storage  persistence.Storage
	}
	type args struct {
		metricIDs []string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantInterval int
	}{
		{
			name: "Get interval with metrics",
			fields: fields{
				interval: DefaultInterval,
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
					_ = s.Save(&assessment.Metric{Id: "M1", Interval: 60})
					_ = s.Save(&assessment.Metric{Id: "M2", Interval: 50})
					_ = s.Save(&assessment.Metric{Id: "M3", Interval: 0}) // invalid interval should be ignored
					_ = s.Save(&assessment.Metric{Id: "M4", Interval: 40})

				}),
			},
			args:         args{metricIDs: []string{"M1", "M2", "M3"}},
			wantInterval: 50,
		},
		{
			name: "Get interval with metrics without intervals",
			fields: fields{
				interval: DefaultInterval,
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
					_ = s.Save(&assessment.Metric{Id: "M1"})
					_ = s.Save(&assessment.Metric{Id: "M2"})

				}),
			},
			args:         args{metricIDs: []string{"M1", "M2"}},
			wantInterval: DefaultInterval,
		},
		{
			name: "Get interval when no metrics exist",
			fields: fields{
				interval: DefaultInterval,
				storage:  testutil.NewInMemoryStorage(t),
			},
			args:         args{metricIDs: []string{"M1", "M2"}},
			wantInterval: DefaultInterval,
		},
		{
			name: "Get interval when DB error occurs",
			fields: fields{
				interval: DefaultInterval,
				storage:  &testutil.StorageWithError{GetErr: gorm.ErrInvalidData},
			},
			args:         args{metricIDs: []string{"M1", "M2"}},
			wantInterval: DefaultInterval,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				interval: tt.fields.interval,
				storage:  tt.fields.storage,
			}
			assert.Equalf(t, tt.wantInterval, srv.calculateInterval(tt.args.metricIDs), "calculateInterval(%v)", tt.args.metricIDs)
		})
	}
}
