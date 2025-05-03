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
	"io"
	"time"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/eclipse-xfsc/cam/api/common"
	"github.com/eclipse-xfsc/cam/api/evaluation"
	"github.com/eclipse-xfsc/cam/service"

	"clouditor.io/clouditor/api"
	cl_api_assessment "clouditor.io/clouditor/api/assessment"
	cl_api_evidence "clouditor.io/clouditor/api/evidence"
	"clouditor.io/clouditor/api/orchestrator"
	"clouditor.io/clouditor/persistence"
	"clouditor.io/clouditor/policies"
	cl_service "clouditor.io/clouditor/service"
	cl_service_assessment "clouditor.io/clouditor/service/assessment"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// DefaultListComplianceDays indicates the default value for `days` in the ListCompliance endpoint
	DefaultListComplianceDays = int64(30)
	// DefaultListEvidencesDays indicates the default value for `days` in the ListEvidences endpoint
	DefaultListEvidencesDays = int64(30)
)

var (
	log = logrus.WithField("component", "evaluation")

	DefaultRequirementsManagerAddress = grpcTarget{target: "127.0.0.1:50100"}
)

type grpcTarget struct {
	target string
	opts   []grpc.DialOption
}

type Server struct {
	evaluation.UnimplementedEvaluationServer

	// Clouditor's assessment service
	*cl_service_assessment.Service

	// reqManagerAddress is the gRPC address used for the connection to the requirements manager
	reqManagerAddress grpcTarget

	requirementsSource policies.RequirementsSource

	// storage is our persistence backend
	storage persistence.Storage

	authorizer api.Authorizer
}

// WithRequirementsManagerAddress is a Server option for setting the address of the Requirements Manager
func WithRequirementsManagerAddress(address string, opts ...grpc.DialOption) service.ServiceOption[Server] {
	return func(srv *Server) {
		srv.reqManagerAddress = grpcTarget{
			target: address,
			opts:   opts,
		}
	}
}

// WithStorage is an option to set the storage. If not set, NewServer will use inmemory storage.
func WithStorage(storage persistence.Storage) service.ServiceOption[Server] {
	return func(srv *Server) {
		srv.storage = storage
	}
}

// WithOAuth2Authorizer is an option to use an OAuth 2.0 authorizer
func WithOAuth2Authorizer(config *clientcredentials.Config) service.ServiceOption[Server] {
	return func(srv *Server) {
		srv.SetAuthorizer(api.NewOAuthAuthorizerFromClientCredentials(config))
	}
}

// SetAuthorizer implements UsesAuthorizer
func (srv *Server) SetAuthorizer(auth api.Authorizer) {
	srv.authorizer = auth
}

// Authorizer implements UsesAuthorizer
func (srv *Server) Authorizer() api.Authorizer {
	return srv.authorizer
}

// NewServer creates a new evaluation Server/service
func NewServer(opts ...service.ServiceOption[Server]) (srv *Server) {
	srv = &Server{
		reqManagerAddress: DefaultRequirementsManagerAddress,
	}

	// Apply any options
	for _, o := range opts {
		o(srv)
	}

	// Check if storage is set
	if srv.storage == nil {
		log.Errorf("Storage not initialized. Evaluation Server will probably not function correctly. " +
			"Use `WithStorage` option to set up a storage.")
	}

	// Build a new Clouditor assessment service. It takes care of the actual assessment of our policies.
	// It also fetches the needed metrics and their implementations automatically from the Requirements Manager,
	// since the Requirements Manager implements the Clouditor Orchestrator interface.
	srv.Service = cl_service_assessment.NewService(
		cl_service_assessment.WithOrchestratorAddress(srv.reqManagerAddress.target, srv.reqManagerAddress.opts...),
		cl_service_assessment.WithRegoPackageName("xfsc.metrics"),
		cl_service_assessment.WithoutEvidenceStore(),
		cl_service_assessment.WithAuthorizer(srv.authorizer),
	)

	srv.Service.RegisterAssessmentResultHook(srv.createEvaluationResult)

	srv.requirementsSource = CachedRequirementsSource(srv.Service)

	return
}

// SendEvidences stores and evaluates evidences sent by a SendEvidencesClient via a stream
func (srv *Server) SendEvidences(stream evaluation.Evaluation_SendEvidencesServer) (err error) {
	var (
		// Evidence in CAM format
		evidence *common.Evidence
		// Evidence in Clouditor format
		clouditorEvidence *cl_api_evidence.Evidence
		// Evidence including error is stored but not evaluated
		isEvidenceWithError = false
	)

	// Loop through stream for receiving evidences
	for {
		evidence, err = stream.Recv()
		// If error represents end of file, log it and return
		if errors.Is(err, io.EOF) {
			log.Errorf("Received final input: %v", err)
			return nil
		}
		// If another (connection) error occurs, log it and return
		if err != nil {
			err = fmt.Errorf("cannot receive stream request: %v", err)
			log.Error(err)
			// Transform error into gRPC error and return it
			err = status.Error(codes.Unknown, err.Error())
			return
		}

		log.Infof("Received evidence %s from collection module %s", evidence.Id, evidence.ToolId)

		// Validate evidence
		// TODO(lebogg): Directly return since it is likely that the following evidences will also be invalid
		if err = evidence.Validate(); err != nil {
			if errors.Is(err, common.ErrEvidenceWithError) {
				log.Info("Error contains error and, thus, is not evaluated.")
				isEvidenceWithError = true
			} else {
				log.Errorf("Evidence is not valid: %v", err)
				continue
			}
		}

		// Use `AssessEvidence` of Clouditor's assessment service to evaluate the evidence
		// First we transform a CAM evidence into a Clouditor evidence
		clouditorEvidence = &cl_api_evidence.Evidence{
			Id:        evidence.Id,
			Timestamp: evidence.GatheredAt,
			ServiceId: evidence.TargetService,
			ToolId:    evidence.ToolId,
			Raw:       evidence.RawEvidence,
			Resource:  evidence.Value,
		}

		err = srv.storage.Create(evidence)
		if err != nil {
			err = fmt.Errorf("couldn't store evidence: %w", err)
			log.Error(err)
			continue
		}
		log.Tracef("Stored evidence: %v", evidence)

		// Evidence that includes error won't be evaluated
		if isEvidenceWithError {
			isEvidenceWithError = false
			continue
		}

		// Use transformed evidence to evaluate evidence ("to assess" in Clouditor terminology).
		// We discard response since it is used in the streaming case `AssessEvidences`
		_, err = srv.Service.AssessEvidence(
			context.TODO(),
			&cl_api_assessment.AssessEvidenceRequest{Evidence: clouditorEvidence})
		// Log error and continue since we still want to evaluate new incoming evidences
		if err != nil {
			err = fmt.Errorf("couldn't evaluate evidence: %w", err)
			log.Error(err)
			continue
		}
		log.Infof("Assessed new evidence")
	}
}

// GetEvidence returns the evidence given by evidence_id.
func (srv *Server) GetEvidence(_ context.Context, req *evaluation.GetEvidenceRequest) (e *common.Evidence, err error) {
	e, err = srv.getEvidence(req.EvidenceId)
	// Check if error occurred and, if so, wrap it into an error in gRPC format
	if errors.Is(err, persistence.ErrRecordNotFound) {
		err = status.Errorf(codes.NotFound, "%v: %v", EvidenceNotFoundErrorMsg, err)
		return
	}
	// Any other error is a DB error
	if err != nil {
		err = status.Errorf(codes.Internal, "%s: %v", DatabaseErrorMsg, err)
		return
	}
	return
}

// ListEvidences returns evidences. The amount of evidences per request and the scope are determined by the query parameters.
func (srv *Server) ListEvidences(_ context.Context, req *evaluation.ListEvidencesRequest) (
	res *evaluation.ListEvidencesResponse, err error) {
	res = new(evaluation.ListEvidencesResponse)
	if req.ServiceId == "" {
		err = status.Errorf(codes.InvalidArgument, "Service ID is missing")
		return
	}
	// TODO(lebogg): Its in the RMs DB! Would need to add RM client here
	// Verify that the Cloud Service exists
	//err = storage.VerifyExistence(srv.storage, &orchestrator.CloudService{}, "id", req.ServiceId)
	//if err != nil {
	//	return
	//}

	days := req.Days
	if days == 0 {
		days = DefaultListEvidencesDays
	}

	now := time.Now()
	before := now.Add(-time.Hour * 24 * time.Duration(days))

	res.Evidences, res.NextPageToken, err = cl_service.PaginateStorage[*common.Evidence](req, srv.storage,
		cl_service.DefaultPaginationOpts,
		"target_service = ? AND gathered_at BETWEEN ? AND ?", req.ServiceId, before, now)
	if err != nil {
		err = status.Errorf(codes.Internal, "%s: %v", DatabaseErrorMsg, err)
		return
	}
	// Return the list of evidences (can be of length zero)
	return
}

// GetEvaluation returns the most recent evaluation result for given service and metric IDs
func (srv *Server) GetEvaluation(_ context.Context, request *evaluation.GetEvaluationRequest) (eval *evaluation.EvaluationResult, err error) {
	var list []*evaluation.EvaluationResult

	// We need to use List because Get does not have any ordering. We limit the
	// list to 1 (which is essentially what Get does anyway)
	err = srv.storage.List(&list, "time", false, 0, 1, "metric_id = ? AND service_id = ?", request.MetricId, request.ServiceId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	if len(list) == 0 {
		return nil, status.Error(codes.NotFound, "evaluation result not found")
	}

	eval = list[0]

	return
}

// GetCompliance returns the most recent compliance result for a particular control of a service
func (srv *Server) GetCompliance(_ context.Context, request *evaluation.GetComplianceRequest) (compliance *evaluation.Compliance, err error) {
	var list []*evaluation.Compliance

	// We need to use List because Get does not have any ordering. We limit the
	// list to 1 (which is essentially what Get does anyway)
	err = srv.storage.List(&list, "time", false, 0, 1, "service_id = ?", request.ServiceId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	if len(list) == 0 {
		return nil, status.Error(codes.NotFound, ComplianceResultNotFoundErrorMsg)
	}

	compliance = list[0]

	return
}

// ListCompliance returns all current compliance results for a particular service
func (srv *Server) ListCompliance(_ context.Context, req *evaluation.ListComplianceRequest) (
	res *evaluation.ListComplianceResponse, err error) {
	res = new(evaluation.ListComplianceResponse)
	days := req.Days
	if days == 0 {
		days = DefaultListComplianceDays
	}

	now := time.Now()
	before := now.Add(-time.Hour * 24 * time.Duration(days))

	res.ComplianceResults, res.NextPageToken, err = cl_service.PaginateStorage[*evaluation.Compliance](
		req, srv.storage, cl_service.DefaultPaginationOpts,
		"service_id = ? AND time BETWEEN ?  AND ?", req.ServiceId, before, now)
	if err != nil {
		err = status.Errorf(codes.Internal, "%s: %v", DatabaseErrorMsg, err)
		return
	}
	// If there are no compliance results yet, we return empty list (and not an error)
	return
}

// CalculateCompliance triggers the compliance calculation for a particular services and a set of controls. This will
// most likely be triggered by the Requirements Manager.
func (srv *Server) CalculateCompliance(_ context.Context, req *evaluation.CalculateComplianceRequest) (empty *emptypb.Empty, err error) {
	for _, controlID := range req.ControlIds {
		_, err = srv.calculateComplianceInternal(req.ServiceId, controlID)
		if err != nil {
			log.Errorf("Error while calculating compliance for service %s: %v. Compliance results will only be partially available.", req.ServiceId, err)
		}
	}

	empty = new(emptypb.Empty)

	return
}

// calculateCompliance by checking the status of each evaluation result in newestResults
func (srv *Server) calculateComplianceInternal(serviceID, controlID string) (compliance *evaluation.Compliance, err error) {
	log.Infof("Calculating compliance for service '%s' and control '%s'", serviceID, controlID)
	var requirements []*orchestrator.Requirement
	var control *orchestrator.Requirement

	// Start with compliant case. Non-compliant, if at least one result status is false
	compliance = &evaluation.Compliance{
		Id:          uuid.NewString(),
		ControlId:   controlID,
		Evaluations: nil,
		Status:      true,
		Time:        timestamppb.Now(),
		ServiceId:   serviceID,
	}

	// Fetch requirements (controls) from our source
	requirements, err = srv.requirementsSource.Requirements()
	if err != nil {
		return nil, fmt.Errorf("could not fetch requirements: %w", err)
	}

	for _, r := range requirements {
		if r.Id == controlID {
			control = r
			break
		}
	}

	if control == nil {
		return nil, fmt.Errorf("control not found")
	}

	// For each metric (of this control), get latest evaluation result and calculate the compliance
	// TODO(oxisto): We should actually get the latest evaluation result _per_ discovered resource
	for _, m := range control.Metrics {
		var result []*evaluation.EvaluationResult
		err = srv.storage.List(&result, "time", false, 0, 1, "metric_id = ? AND service_id = ?", m.Id, serviceID)
		if err != nil {
			return nil, fmt.Errorf("error while retrieving evaluation result from storage: %w", err)
		}

		// There is no result (currently) for this metric -> continue with next one
		if len(result) == 0 {
			continue
		}

		// If one result status is false -> non-compliant
		if !result[0].Status {
			compliance.Status = false
		}
		compliance.Evaluations = append(compliance.Evaluations, result[0])
	}

	// Store it
	err = srv.storage.Create(compliance)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	log.Debugf("Is compliant: %v", compliance.Status)
	return
}

// createEvaluationResult creates a XFSC evaluation result out of a Clouditor
// assessment result. In the future, we might merge the two concepts, but for
// now unfortunately, we have to do it like this.
func (srv *Server) createEvaluationResult(result *cl_api_assessment.AssessmentResult, err error) {
	var (
		evidence *common.Evidence
	)

	// Catch potential error
	if err != nil {
		log.Errorf("Can not create evaluation result: %s", err.Error())
		return
	}

	// Fetch evidence
	evidence, err = srv.getEvidence(result.EvidenceId)
	if err != nil {
		err = fmt.Errorf("could not fetch evidence %s for assessment result %s: %w", result.EvidenceId, result.Id, err)
		log.Error(err)
		return
	}

	eval := &evaluation.EvaluationResult{
		Id:         uuid.NewString(),
		MetricId:   result.MetricId,
		ServiceId:  evidence.TargetService,
		EvidenceId: evidence.Id,
		Status:     result.Compliant,
		Time:       timestamppb.Now(),
	}

	err = srv.storage.Create(&eval)
	log.Tracef("Stored evaluation result: %v", eval)
	if err != nil {
		log.Errorf("Could not save result into database: %v", err)
		return
	}
}

// getEvidence returns the Evidence for a given evidenceID. Errors are returned with wrapped error constants s.t. the
// caller can propagate them in a proper way.
func (srv *Server) getEvidence(evidenceId string) (e *common.Evidence, err error) {
	// Mak sure to init evidence e since API calls could fail otherwise (marshalling of responses in the gRPC GW)
	e = new(common.Evidence)
	err = srv.storage.Get(&e, "id = ?", evidenceId)
	return
}

type cachedRequirementsSource struct {
	src          policies.RequirementsSource
	requirements []*orchestrator.Requirement
}

func (c *cachedRequirementsSource) Requirements() (requirements []*orchestrator.Requirement, err error) {
	if c.requirements == nil {
		requirements, err = c.src.Requirements()
		if err != nil {
			c.requirements = requirements
		}

		return
	}

	return c.requirements, nil
}

func CachedRequirementsSource(src policies.RequirementsSource) policies.RequirementsSource {
	return &cachedRequirementsSource{
		src: src,
	}
}
