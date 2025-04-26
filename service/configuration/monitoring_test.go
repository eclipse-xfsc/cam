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
	"time"

	"clouditor.io/clouditor/api/orchestrator"
	"clouditor.io/clouditor/persistence"
	service_orchestrator "clouditor.io/clouditor/service/orchestrator"
	"github.com/go-co-op/gocron"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/eclipse-xfsc/cam/api/configuration"
	"github.com/eclipse-xfsc/cam/internal/testutil"
)

func Test_server_StartMonitoring(t *testing.T) {
	storage := testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
		err := populateStorage(t, s)
		assert.NoError(t, err)
	})

	type fields struct {
		OrchestratorService orchestrator.OrchestratorServer
		storage             persistence.Storage
		monitoring          map[string]*MonitorScheduler
	}
	type args struct {
		ctx context.Context
		req *configuration.StartMonitoringRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.ValueAssertionFunc
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Good",
			fields: fields{
				OrchestratorService: service_orchestrator.NewService(service_orchestrator.WithStorage(storage)),
				storage:             storage,
				monitoring: map[string]*MonitorScheduler{
					"0000": {
						scheduler:         mockStartedScheduler(t),
						monitoredControls: []string{"C1"},
					},
				},
			},
			args: args{
				req: &configuration.StartMonitoringRequest{ControlIds: []string{"Req-1"}, ServiceId: "myService"},
			},
			wantRes: func(tt assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				resp, ok := i1.(*configuration.StartMonitoringResponse)
				if !ok {
					return false
				}

				return assert.Equal(t, "myService", resp.Status.ServiceId)
			},
			wantErr: assert.NoError,
		},
		{
			name: "Invalid request",
			fields: fields{
				OrchestratorService: service_orchestrator.NewService(service_orchestrator.WithStorage(storage)),
				storage:             storage,
				monitoring:          make(map[string]*MonitorScheduler),
			},
			args: args{
				req: nil,
			},
			wantRes: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, codes.InvalidArgument, status.Code(err))
				return assert.ErrorContains(t, err, configuration.ErrRequestEmpty.Error())
			},
		},
		{
			name: "Service is monitored already",
			fields: fields{
				OrchestratorService: service_orchestrator.NewService(service_orchestrator.WithStorage(storage)),
				storage:             storage,
				monitoring: map[string]*MonitorScheduler{
					"0000": {
						scheduler:         mockStartedScheduler(t),
						monitoredControls: []string{"C1"},
					},
				},
			},
			args: args{
				req: &configuration.StartMonitoringRequest{ServiceId: "0000", ControlIds: []string{"C1"}},
			},
			wantRes: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, codes.AlreadyExists, status.Code(err))
				return assert.ErrorContains(t, err, "monitored already")
			},
		},
		{
			name: "DB error",
			fields: fields{
				OrchestratorService: service_orchestrator.NewService(service_orchestrator.WithStorage(storage)),
				storage:             &testutil.StorageWithError{ListErr: gorm.ErrInvalidData},
				monitoring:          make(map[string]*MonitorScheduler),
			},
			args: args{
				req: &configuration.StartMonitoringRequest{ServiceId: "0000", ControlIds: []string{"C1"}},
			},
			wantRes: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, codes.Internal, status.Code(err))
				return assert.ErrorContains(t, err, gorm.ErrInvalidData.Error())
			},
		},
		{
			name: "No controls in DB",
			fields: fields{
				OrchestratorService: service_orchestrator.NewService(service_orchestrator.WithStorage(storage)),
				storage:             testutil.NewInMemoryStorage(t),
				monitoring:          make(map[string]*MonitorScheduler),
			},
			args: args{
				req: &configuration.StartMonitoringRequest{ServiceId: "0000", ControlIds: []string{"C1"}},
			},
			wantRes: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, codes.NotFound, status.Code(err))
				return assert.ErrorContains(t, err, "no controls found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				OrchestratorServer: tt.fields.OrchestratorService,
				storage:            tt.fields.storage,
				monitoring:         tt.fields.monitoring,
			}
			gotRes, err := srv.StartMonitoring(tt.args.ctx, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("StartMonitoring(%v, %v)", tt.args.ctx, tt.args.req)) {
				return
			}

			if tt.wantRes != nil {
				tt.wantRes(t, gotRes, tt.args)
			}
		})
	}
}

func Test_server_GetMonitoringStatus(t *testing.T) {
	type fields struct {
		storage    persistence.Storage
		monitoring map[string]*MonitorScheduler
	}
	type args struct {
		ctx context.Context
		req *configuration.GetMonitoringStatusRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.ValueAssertionFunc
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Good",
			fields: fields{
				storage: nil,
				monitoring: map[string]*MonitorScheduler{
					"0000": {
						scheduler:         mockStartedScheduler(t),
						monitoredControls: []string{"C1"},
					},
				},
			},
			args: args{
				ctx: nil,
				req: &configuration.GetMonitoringStatusRequest{ServiceId: "0000"},
			},
			wantRes: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				res, ok := i.(*configuration.MonitoringStatus)
				assert.True(t, ok)
				monitorScheduler, ok := i2[0].(*MonitorScheduler)
				assert.True(t, ok)
				assert.Equal(t, monitorScheduler.scheduler.Jobs()[0].LastRun(), res.LastRun.AsTime())
				assert.Equal(t, monitorScheduler.scheduler.Jobs()[0].NextRun(), res.NextRun.AsTime())
				return assert.Equal(t, res.ServiceId, "0000")
			},
			wantErr: assert.NoError,
		},
		{
			name: "Monitoring never started",
			fields: fields{
				storage:    nil,
				monitoring: nil,
			},
			args: args{
				ctx: nil,
				req: &configuration.GetMonitoringStatusRequest{ServiceId: "0000"},
			},
			wantRes: assert.Nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, codes.NotFound, status.Code(err))
				return assert.ErrorContains(t, err, "has not been started yet")
			},
		},
		{
			name: "Monitoring currently not running",
			fields: fields{
				storage: nil,
				monitoring: map[string]*MonitorScheduler{
					"0000": {
						scheduler:         mockStoppedScheduler(),
						monitoredControls: []string{"C1"},
					},
				},
			},
			args: args{
				ctx: nil,
				req: &configuration.GetMonitoringStatusRequest{ServiceId: "0000"},
			},
			wantRes: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				res, ok := i.(*configuration.MonitoringStatus)
				assert.True(t, ok)
				monitorScheduler, ok := i2[0].(*MonitorScheduler)
				assert.True(t, ok)
				assert.Equal(t, monitorScheduler.scheduler.Jobs()[0].LastRun(), res.LastRun.AsTime())
				assert.Nil(t, res.NextRun)
				assert.Nil(t, res.ControlIds)
				return assert.Equal(t, res.ServiceId, "0000")
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage:    tt.fields.storage,
				monitoring: tt.fields.monitoring,
			}
			gotRes, err := srv.GetMonitoringStatus(tt.args.ctx, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("GetMonitoringStatus(%v, %v)", tt.args.ctx, tt.args.req)) {
				return
			}
			tt.wantRes(t, gotRes, tt.fields.monitoring[tt.args.req.ServiceId])
		})
	}
}

func Test_server_StopMonitoring(t *testing.T) {
	type fields struct {
		storage    persistence.Storage
		monitoring map[string]*MonitorScheduler
	}
	type args struct {
		in0 context.Context
		req *configuration.StopMonitoringRequest
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		want           *configuration.StopMonitoringResponse
		wantMonitoring assert.ComparisonAssertionFunc
		wantErr        assert.ErrorAssertionFunc
	}{
		{
			name: "good",
			fields: fields{
				storage: nil,
				monitoring: map[string]*MonitorScheduler{
					"0000": {
						scheduler:         mockStartedScheduler(t),
						monitoredControls: []string{"C1"},
					},
				},
			},
			args: args{
				in0: nil,
				req: &configuration.StopMonitoringRequest{ServiceId: "0000"},
			},
			want: &configuration.StopMonitoringResponse{},
			wantMonitoring: func(t assert.TestingT, monitoringI interface{}, i2 interface{}, i3 ...interface{}) bool {
				monitoring, ok := monitoringI.(*MonitorScheduler)
				assert.True(t, ok)
				return assert.False(t, monitoring.scheduler.IsRunning())
			},
			wantErr: assert.NoError,
		},
		{
			name: "Monitoring never started",
			fields: fields{
				storage:    nil,
				monitoring: nil,
			},
			args: args{
				in0: nil,
				req: &configuration.StopMonitoringRequest{ServiceId: "0000"},
			},
			want:           nil,
			wantMonitoring: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, codes.NotFound, status.Code(err))
				return assert.ErrorContains(t, err, "has not been started yet")
			},
		},
		{
			name: "Monitoring currently stopped",
			fields: fields{
				storage: nil,
				monitoring: map[string]*MonitorScheduler{
					"0000": {
						scheduler:         mockStoppedScheduler(),
						monitoredControls: []string{"C1"},
					},
				},
			},
			args: args{
				in0: nil,
				req: &configuration.StopMonitoringRequest{ServiceId: "0000"},
			},
			want:           nil,
			wantMonitoring: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, codes.NotFound, status.Code(err))
				return assert.ErrorContains(t, err, "has been stopped already")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage:    tt.fields.storage,
				monitoring: tt.fields.monitoring,
			}
			got, err := srv.StopMonitoring(tt.args.in0, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("StopMonitoring(%v, %v)", tt.args.in0, tt.args.req)) {
				return
			}
			// Assert monitoring after successful call (otherwise, rec could be empty in test)
			if err == nil {
				tt.wantMonitoring(t, srv.monitoring[tt.args.req.ServiceId], nil)
			}
			assert.Equalf(t, tt.want, got, "StopMonitoring(%v, %v)", tt.args.in0, tt.args.req)
		})
	}
}

// mockStartedScheduler returns a mocked scheduler (w/ noop job) which has started already
func mockStartedScheduler(t *testing.T) (s *gocron.Scheduler) {
	s = gocron.NewScheduler(time.UTC)
	_, err := s.Tag(triggerCollectionModuleTag).Every(1).Hours().Do(func() {})
	assert.NoError(t, err)
	s.StartAsync()
	return
}

// mockStoppedScheduler returns a mocked scheduler (w/ noop job) which is stopped (== not running)
func mockStoppedScheduler() (s *gocron.Scheduler) {
	s = gocron.NewScheduler(time.UTC).Tag(triggerCollectionModuleTag)
	return
}
