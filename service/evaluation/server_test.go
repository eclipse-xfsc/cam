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

package evaluation

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"clouditor.io/clouditor/api/assessment"
	cl_api_assessment "clouditor.io/clouditor/api/assessment"
	"clouditor.io/clouditor/api/orchestrator"
	"clouditor.io/clouditor/persistence"
	"clouditor.io/clouditor/policies"
	cl_service_assessment "clouditor.io/clouditor/service/assessment"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/eclipse-xfsc/cam/api/common"
	"github.com/eclipse-xfsc/cam/api/evaluation"
	"github.com/eclipse-xfsc/cam/internal/testutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type staticRequirementsSource struct {
	requirements []*orchestrator.Requirement
}

func (s *staticRequirementsSource) Requirements() (requirements []*orchestrator.Requirement, err error) {
	return s.requirements, nil
}

func StaticRequirementsSource(requirements []*orchestrator.Requirement) policies.RequirementsSource {
	return &staticRequirementsSource{
		requirements: requirements,
	}
}

var TestRequirementsSource = StaticRequirementsSource([]*orchestrator.Requirement{
	{
		Id: "Control1",
		Metrics: []*assessment.Metric{
			{Id: "Metric1"},
			{Id: "Metric2"},
		},
	},
})

func Test_Server_GetEvaluation(t *testing.T) {
	type fields struct {
		Service            *cl_service_assessment.Service
		requirementsSource policies.RequirementsSource
		storage            persistence.Storage
	}
	type args struct {
		in0     context.Context
		request *evaluation.GetEvaluationRequest
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantEval *evaluation.EvaluationResult
		wantErr  bool
	}{
		{
			name: "latest evaluation result",
			fields: fields{
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
					_ = s.Save(&evaluation.EvaluationResult{
						Id:        "00000000-0000-0000-0000-000000000000",
						MetricId:  "Metric1",
						ServiceId: "MyService",
						Time:      timestamppb.New(time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC)),
					})
					_ = s.Save(&evaluation.EvaluationResult{
						Id:        "11111111-1111-1111-1111-111111111111",
						MetricId:  "Metric1",
						ServiceId: "MyService",
						Time:      timestamppb.New(time.Date(2022, 2, 1, 1, 1, 1, 1, time.UTC)),
					})
				}),
			},
			args: args{request: &evaluation.GetEvaluationRequest{
				ServiceId: "MyService",
				MetricId:  "Metric1",
			}},
			wantEval: &evaluation.EvaluationResult{
				Id:        "11111111-1111-1111-1111-111111111111",
				MetricId:  "Metric1",
				ServiceId: "MyService",
				Time:      timestamppb.New(time.Date(2022, 2, 1, 1, 1, 1, 1, time.UTC)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Service:            tt.fields.Service,
				requirementsSource: tt.fields.requirementsSource,
				storage:            tt.fields.storage,
			}
			gotEval, err := s.GetEvaluation(tt.args.in0, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.GetEvaluation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotEval, tt.wantEval) {
				t.Errorf("Server.GetEvaluation() = %v, want %v", gotEval, tt.wantEval)
			}
		})
	}
}

func Test_Server_calculateComplianceInternal(t *testing.T) {
	type fields struct {
		Service            *cl_service_assessment.Service
		requirementsSource policies.RequirementsSource
		storage            persistence.Storage
	}
	type args struct {
		serviceID string
		controlID string
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantCompliance assert.ValueAssertionFunc
		wantErr        bool
	}{
		{
			name: "No results",
			fields: fields{
				requirementsSource: TestRequirementsSource,
				storage:            testutil.NewInMemoryStorage(t),
			},
			args: args{
				serviceID: "MyService",
				controlID: "Control1",
			},
			wantCompliance: func(tt assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				c, ok := i1.(*evaluation.Compliance)
				if !ok {
					return false
				}

				if !assert.Equal(t, 0, len(c.Evaluations)) {
					return false
				}
				if !assert.True(t, c.Status) {
					return false
				}
				if !assert.Equal(t, "Control1", c.ControlId) {
					return false
				}

				return true
			},
		},
		{
			name: "Latest results compliant",
			fields: fields{
				requirementsSource: TestRequirementsSource,
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
					// This is an old result, which is not relevant for the current state
					_ = s.Save(&evaluation.EvaluationResult{
						Id:        uuid.NewString(),
						ServiceId: "MyService",
						MetricId:  "Metric1",
						Status:    false,
						Time:      timestamppb.New(time.Now().Add(-10 * time.Minute)),
					})
					_ = s.Save(&evaluation.EvaluationResult{
						Id:        uuid.NewString(),
						ServiceId: "MyService",
						MetricId:  "Metric1",
						Status:    true,
						Time:      timestamppb.Now(),
					})
					_ = s.Save(&evaluation.EvaluationResult{
						Id:        uuid.NewString(),
						ServiceId: "MyService",
						MetricId:  "Metric2",
						Status:    true,
						Time:      timestamppb.Now(),
					})
				}),
			},
			args: args{
				serviceID: "MyService",
				controlID: "Control1",
			},
			wantCompliance: func(tt assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				c, ok := i1.(*evaluation.Compliance)
				if !ok {
					return false
				}

				if !assert.Equal(t, 2, len(c.Evaluations)) {
					return false
				}
				if !assert.True(t, c.Status) {
					return false
				}
				if !assert.Equal(t, "Control1", c.ControlId) {
					return false
				}

				return true
			},
		},
		{
			name: "Latest results non-compliant",
			fields: fields{
				requirementsSource: TestRequirementsSource,
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
					_ = s.Save(&evaluation.EvaluationResult{
						Id:        uuid.NewString(),
						ServiceId: "MyService",
						MetricId:  "Metric1",
						Status:    false,
						Time:      timestamppb.New(time.Now().Add(-10 * time.Minute)),
					})
				}),
			},
			args: args{
				serviceID: "MyService",
				controlID: "Control1",
			},
			wantCompliance: func(tt assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				c, ok := i1.(*evaluation.Compliance)
				if !ok {
					return false
				}

				if !assert.Equal(t, 1, len(c.Evaluations)) {
					return false
				}
				if !assert.False(t, c.Status) {
					return false
				}
				if !assert.Equal(t, "Control1", c.ControlId) {
					return false
				}

				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Service:            tt.fields.Service,
				requirementsSource: tt.fields.requirementsSource,
				storage:            tt.fields.storage,
			}

			gotCompliance, err := s.calculateComplianceInternal(tt.args.serviceID, tt.args.controlID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.calculateCompliance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantCompliance != nil && !tt.wantCompliance(t, gotCompliance) {
				t.Errorf("Server.calculateCompliance() = %v, want %v", gotCompliance, tt.wantCompliance)
			}
		})
	}
}

func Test_Server_getEvidence(t *testing.T) {
	// Evidence with error to check if json type in gorm works
	e1 := &common.Evidence{Id: "e1", Error: &common.Error{
		Code:        common.Error_ERROR_CONNECTION_FAILURE,
		Description: "No connection",
	}}
	type fields struct {
		UnimplementedEvaluationServer evaluation.UnimplementedEvaluationServer
		Service                       *cl_service_assessment.Service
		reqManagerAddress             grpcTarget
		requirementsSource            policies.RequirementsSource
		storage                       persistence.Storage
	}
	type args struct {
		evidenceId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantE   *common.Evidence
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "correct",
			fields: fields{
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
					err := s.Save(e1)
					assert.NoError(t, err)
				}),
			},
			args:    args{evidenceId: "e1"},
			wantE:   e1,
			wantErr: assert.NoError,
		},
		{
			name: "evidence not found",
			fields: fields{
				storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
					err := s.Save(&common.Evidence{Id: "e1"})
					assert.NoError(t, err)
				}),
			},
			args:  args{evidenceId: "e2"},
			wantE: &common.Evidence{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, persistence.ErrRecordNotFound)
			},
		},
		{
			name: "storage error",
			fields: fields{
				storage: &testutil.StorageWithError{GetErr: gorm.ErrInvalidData},
			},
			args:  args{evidenceId: "e1"},
			wantE: &common.Evidence{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, gorm.ErrInvalidData)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				UnimplementedEvaluationServer: tt.fields.UnimplementedEvaluationServer,
				Service:                       tt.fields.Service,
				reqManagerAddress:             tt.fields.reqManagerAddress,
				requirementsSource:            tt.fields.requirementsSource,
				storage:                       tt.fields.storage,
			}
			gotE, err := srv.getEvidence(tt.args.evidenceId)
			if !tt.wantErr(t, err, fmt.Sprintf("getEvidence(%v)", tt.args.evidenceId)) {
				return
			}
			assert.Equalf(t, tt.wantE, gotE, "getEvidence(%v)", tt.args.evidenceId)
		})
	}
}

func Test_Server_ListCompliance(t *testing.T) {
	now := time.Now()
	beforeDefaultThreshold := now.Add(-time.Hour * 24 * time.Duration(DefaultListComplianceDays+1))
	daysFromRequest := int64(10)
	between := now.Add(-time.Hour * 24 * time.Duration(daysFromRequest-1)) // -1 s.t. it will be in the range
	c1 := &evaluation.Compliance{Id: "1", ServiceId: "1", Status: true, ControlId: "1", Time: timestamppb.New(beforeDefaultThreshold)}
	c2 := &evaluation.Compliance{Id: "2", ServiceId: "1", Status: false, ControlId: "1", Time: timestamppb.New(now)}
	c3 := &evaluation.Compliance{Id: "3", ServiceId: "1", Status: false, ControlId: "1", Time: timestamppb.New(between)}
	type fields struct {
		storage persistence.Storage
	}
	type args struct {
		in0     context.Context
		request *evaluation.ListComplianceRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes assert.ComparisonAssertionFunc
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Good with default days",
			fields: fields{storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
				_ = s.Save(c1)
				_ = s.Save(c2)
			})},
			args: args{
				request: &evaluation.ListComplianceRequest{ServiceId: "1"},
			},
			wantRes: func(t assert.TestingT, i interface{}, i2 interface{}, i3 ...interface{}) bool {
				res, ok := i.(*evaluation.ListComplianceResponse)
				assert.True(t, ok)
				assert.Len(t, res.ComplianceResults, 1) // since c1's date is out of range
				assert.Equal(t, c2.Id, res.ComplianceResults[0].Id)
				assert.Equal(t, c2.ServiceId, res.ComplianceResults[0].ServiceId)
				assert.Equal(t, c2.ControlId, res.ComplianceResults[0].ControlId)
				assert.Equal(t, c2.Status, res.ComplianceResults[0].Status)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "Good with user-defined days",
			fields: fields{storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
				_ = s.Save(c1)
				_ = s.Save(c2)
				_ = s.Save(c3)
			})},
			args: args{
				request: &evaluation.ListComplianceRequest{ServiceId: "1", Days: daysFromRequest},
			},
			wantRes: func(t assert.TestingT, i interface{}, i2 interface{}, i3 ...interface{}) bool {
				res, ok := i.(*evaluation.ListComplianceResponse)
				assert.True(t, ok)
				assert.Len(t, res.ComplianceResults, 2) // since c1's date is out of range
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "No compliances found",
			fields: fields{storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
				_ = s.Save(c1)
				_ = s.Save(c2)
			})},
			args: args{
				request: &evaluation.ListComplianceRequest{ServiceId: "2"}, // no compliance for this service
			},
			wantRes: func(t assert.TestingT, i interface{}, i2 interface{}, i3 ...interface{}) bool {
				res, ok := i.(*evaluation.ListComplianceResponse)
				assert.True(t, ok)
				assert.Len(t, res.ComplianceResults, 0)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name:   "DB error",
			fields: fields{storage: &testutil.StorageWithError{ListErr: gorm.ErrInvalidData}},
			args: args{
				request: &evaluation.ListComplianceRequest{ServiceId: "2"}, // no compliance for this service
			},
			wantRes: func(t assert.TestingT, i interface{}, i2 interface{}, i3 ...interface{}) bool {
				res, ok := i.(*evaluation.ListComplianceResponse)
				assert.True(t, ok)
				assert.Len(t, res.ComplianceResults, 0)
				return true
			},
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
			gotRes, err := srv.ListCompliance(tt.args.in0, tt.args.request)
			if !tt.wantErr(t, err, fmt.Sprintf("ListCompliance(%v, %v)", tt.args.in0, tt.args.request)) {
				return
			}
			tt.wantRes(t, gotRes, nil)
		})
	}
}

func Test_Server_ListEvidences(t *testing.T) {
	e1 := &common.Evidence{Id: "e1", TargetService: "s1",
		GatheredAt: timestamppb.Now()}
	e2 := &common.Evidence{Id: "e2", TargetService: "s1",
		GatheredAt: timestamppb.Now()}
	type fields struct {
		storage persistence.Storage
	}
	type args struct {
		in0 context.Context
		req *evaluation.ListEvidencesRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes *evaluation.ListEvidencesResponse
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Correct",
			fields: fields{storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
				assert.NoError(t, s.Create(&orchestrator.CloudService{Id: "s1"}))
				assert.NoError(t, s.Create(e1))
				assert.NoError(t, s.Create(e2))
			})},
			args: args{
				req: &evaluation.ListEvidencesRequest{ServiceId: "s1"},
			},
			wantRes: &evaluation.ListEvidencesResponse{
				Evidences:     []*common.Evidence{e1, e2},
				NextPageToken: "",
			},
			wantErr: assert.NoError,
		},
		{
			name: "No service id given",
			fields: fields{storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
				assert.NoError(t, s.Create(&orchestrator.CloudService{Id: "s1"}))
			})},
			args: args{
				req: &evaluation.ListEvidencesRequest{ServiceId: ""},
			},
			wantRes: &evaluation.ListEvidencesResponse{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "Service ID is missing")
			},
		},
		{
			name: "No evidences in DB",
			fields: fields{storage: testutil.NewInMemoryStorage(t, func(s persistence.Storage) {
				assert.NoError(t, s.Create(&orchestrator.CloudService{Id: "s1"}))
			})},
			args: args{
				req: &evaluation.ListEvidencesRequest{ServiceId: "s1"},
			},
			wantRes: &evaluation.ListEvidencesResponse{
				Evidences:     []*common.Evidence{},
				NextPageToken: "",
			},
			wantErr: assert.NoError,
		},
		{
			name: "DB error",
			fields: fields{storage: testutil.NewInMemoryStorageWithListError(t, DatabaseErrorMsg, func(s persistence.Storage) {
				assert.NoError(t, s.Create(&orchestrator.CloudService{Id: "s1"}))
			})},
			args: args{
				req: &evaluation.ListEvidencesRequest{ServiceId: "s1"},
			},
			wantRes: &evaluation.ListEvidencesResponse{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, DatabaseErrorMsg)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage: tt.fields.storage,
			}
			gotRes, err := srv.ListEvidences(tt.args.in0, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("ListEvidences(%v, %v)", tt.args.in0, tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.wantRes, gotRes, "ListEvidences(%v, %v)", tt.args.in0, tt.args.req)
		})
	}
}

func TestServer_createEvaluationResult(t *testing.T) {
	e := &common.Evidence{
		Id:            "00000000-0000-0000-0000-000000000001",
		TargetService: testutil.DefaultServiceID,
	}

	type args struct {
		evidence *common.Evidence
		result   *cl_api_assessment.AssessmentResult
		err      error
	}
	r := &cl_api_assessment.AssessmentResult{
		Id:        "10000000-0000-0000-0000-000000000001",
		Timestamp: timestamppb.Now(),
		MetricId:  "SomeMetricID",
		MetricConfiguration: &cl_api_assessment.MetricConfiguration{
			Operator:    "==",
			TargetValue: structpb.NewStringValue("3"),
			IsDefault:   true,
		},
		Compliant:             true,
		EvidenceId:            "00000000-0000-0000-0000-000000000001",
		ResourceId:            uuid.NewString(),
		ResourceTypes:         []string{"Compute", "VM"},
		NonComplianceComments: "",
		ServiceId:             testutil.DefaultServiceID,
	}
	tests := []struct {
		name string
		args args
		want assert.ValueAssertionFunc
	}{
		{
			name: "Successfully created the evaluation result",
			args: args{
				evidence: e,
				result:   r,
				err:      nil,
			},
			want: func(t assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				s, ok := i1.(*Server)
				assert.True(t, ok)
				var got *evaluation.EvaluationResult
				got, err := s.GetEvaluation(context.Background(), &evaluation.GetEvaluationRequest{
					ServiceId: testutil.DefaultServiceID,
					MetricId:  "SomeMetricID",
				})
				assert.NoError(t, err)
				assert.Equal(t, "00000000-0000-0000-0000-000000000001", got.EvidenceId)
				assert.True(t, got.Status)
				return true
			},
		},
		{
			name: "Can not create the evaluation result due to error",
			args: args{
				evidence: e,
				result:   r,
				err:      errors.New("SomeError"),
			},
			want: func(t assert.TestingT, i1 interface{}, i2 ...interface{}) bool {
				s, ok := i1.(*Server)
				assert.True(t, ok)
				_, err := s.GetEvaluation(context.Background(), &evaluation.GetEvaluationRequest{
					ServiceId: testutil.DefaultServiceID,
					MetricId:  "SomeMetricID",
				})
				assert.Equal(t, codes.NotFound, status.Code(err))
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage: testutil.NewInMemoryStorage(t),
			}
			err := srv.storage.Create(*tt.args.evidence)
			assert.NoError(t, err)
			srv.createEvaluationResult(tt.args.result, tt.args.err)
			tt.want(t, srv)
		})
	}
}
