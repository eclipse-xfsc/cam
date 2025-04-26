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
	"time"

	"clouditor.io/clouditor/api"
	"clouditor.io/clouditor/api/assessment"
	"clouditor.io/clouditor/api/orchestrator"
	"github.com/go-co-op/gocron"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/configuration"
	"github.com/eclipse-xfsc/cam/api/evaluation"
)

const (
	// Tags for scheduler jobs
	triggerCollectionModuleTag = "TriggerCM"
	calculateComplianceTag     = "CalculateCompliance"

	windowSize = 2 * time.Second
	threshold  = 100
)

// StartMonitoring enables the monitoring of the specified control IDs for the given service.
func (srv *Server) StartMonitoring(ctx context.Context, req *configuration.StartMonitoringRequest) (
	res *configuration.StartMonitoringResponse, err error) {
	res = new(configuration.StartMonitoringResponse)
	var (
		collectionModules []*collection.CollectionModule
		filteredModules   []*collection.CollectionModule
		controls          []*orchestrator.Requirement
		metricIDs         []string
	)

	// Validate request
	if err = req.Validate(); err != nil {
		err = status.Error(codes.InvalidArgument, err.Error())
		return
	}
	// Verify that monitoring of this service and controls hasn't started already
	if m := srv.monitoring[req.ServiceId]; m != nil && m.scheduler != nil && m.scheduler.IsRunning() {
		err = status.Errorf(codes.AlreadyExists, "Service %s is being monitored already. Stop it first, "+
			"if, e.g., you want to monitor a new set of controls for this service ", req.ServiceId)
		return
	}

	// Retrieve all controls via the respective IDs
	err = srv.storage.List(&controls, "id", true, 0, -1, "ID IN ?", req.ControlIds)
	if err != nil {
		err = status.Errorf(codes.Internal, "database error: %v", err)
		return
	}
	if len(controls) == 0 {
		err = status.Errorf(codes.NotFound, "no controls found")
		return
	}

	// Populate metric IDs
	for _, control := range controls {
		for _, m := range control.Metrics {
			metricIDs = append(metricIDs, m.Id)
		}
	}

	// List all collection modules and select those that match at least one of
	// the metrics. Due to how our persistence layer currently works, we cannot
	// directly query this from the associated table.
	err = srv.storage.List(&collectionModules, "id", true, 0, -1)
	if err != nil {
		err = status.Errorf(codes.Internal, "database error: %v", err)
		return
	}
	for _, cm := range collectionModules {
		if slices.IndexFunc(cm.Metrics, func(m *assessment.Metric) bool {
			return slices.Contains(metricIDs, m.Id)
		}) != -1 {
			filteredModules = append(filteredModules, cm)
		}
	}

	// Get minimum Metric interval (or Default if no interval is specified in any metric)
	// TODO(lebogg): This is only a workaround due to our inconsistent specification regarding metrics that shouldn't be bound to CMs. and thus, interval as well
	interval := srv.calculateInterval(metricIDs)

	// Start collection modules and run them every $DefaultInterval minutes
	scheduler := gocron.NewScheduler(time.UTC)
	for _, cm := range filteredModules {
		var responsibleMetricIds []string
		for _, metric := range cm.Metrics {
			responsibleMetricIds = append(responsibleMetricIds, metric.Id)
		}

		// We need to add one job for each the collection module. We also tag the job with the metric IDs this
		// collection module feels "responsible" for ...
		_, err = scheduler.Tag(triggerCollectionModuleTag).Tag(responsibleMetricIds...).Every(interval).Seconds().Do(srv.startCollectionModule, cm, req.ServiceId)
		if err != nil {
			log.Errorf("Could not start scheduler for `startCollectionModule` for %s: %v",
				cm.Name, err)
		}
	}

	//// ... and one, which will trigger the compliance calculation in the evaluation manager.
	//_, err = scheduler.Tag(calculateComplianceTag).Every(10).Seconds().WaitForSchedule().Do(func() {
	//	srv.triggerComplianceCalculation(req.ServiceId, req.ControlIds)
	//})
	//if err != nil {
	//	log.Errorf("Could not start scheduler for `triggerComplianceCalculation` for %s: %v",
	//		req.ServiceId, err)
	//}

	srv.monitoring[req.ServiceId] = new(MonitorScheduler)
	srv.monitoring[req.ServiceId].scheduler = scheduler
	srv.monitoring[req.ServiceId].monitoredControls = req.ControlIds

	jobs := scheduler.Jobs()
	log.Debugf("Scheduling %d jobs for execution for service %s", len(jobs), req.ServiceId)

	// Start all collection module jobs
	scheduler.StartAsync()

	log.Debugf("Started monitoring service %s controls: %v", req.ServiceId, req.ControlIds)

	// Retrieve the current status after the start
	res.Status, err = srv.GetMonitoringStatus(ctx, &configuration.GetMonitoringStatusRequest{ServiceId: req.ServiceId})
	if err != nil {
		err = status.Errorf(codes.Internal, "could not retrieve monitoring status after start: %v", err)
		return
	}

	return
}

// GetMonitoringStatus returns the set of controls which are currently monitored for service with specified ID. If
// service is not set up already, returns `Not Found` error. If it is set up but monitoring is currently stopped, return
// empty list of controls
func (srv *Server) GetMonitoringStatus(_ context.Context, req *configuration.GetMonitoringStatusRequest) (
	res *configuration.MonitoringStatus, err error) {
	// If monitoring does not exist (that is the struct is nil), return not found error
	m := srv.monitoring[req.ServiceId]
	if m == nil {
		err = status.Errorf(codes.NotFound, "Monitoring for service %s has not been started yet. "+
			"Start it via the `StartMonitoring` endpoint", req.ServiceId)
		return
	}

	// Get job for triggering collection modules
	jobs, err := m.scheduler.FindJobsByTag(triggerCollectionModuleTag)
	if err != nil {
		log.Errorf("Couldn't find job '%s' for scheduler: %v", triggerCollectionModuleTag, err)
		return
	}
	// We only have one job for each service and tag
	job := jobs[0]

	// Init response w/ serviceID and time of the last run
	res = &configuration.MonitoringStatus{
		ServiceId: req.ServiceId,
		// LastRun is empty if there was no run yet
		LastRun: timestamppb.New(job.LastRun()),
	}

	// If monitoring has been started but is not currently running, an empty list of controls will be returned
	if !m.scheduler.IsRunning() {
		return
	}

	// Add set of currently monitored controls and next scheduled run
	res.ControlIds = m.monitoredControls
	res.NextRun = timestamppb.New(job.NextRun())
	return
}

// StopMonitoring stops monitoring of the service with specified service ID. Returns error (not found), when there is no
// monitoring currently.
func (srv *Server) StopMonitoring(_ context.Context, req *configuration.StopMonitoringRequest) (
	res *configuration.StopMonitoringResponse, err error) {
	m := srv.monitoring[req.ServiceId]
	// Verify that the service is monitored currently
	if m == nil {
		err = status.Errorf(codes.NotFound, "Monitoring of service %s has not been started yet.", req.ServiceId)
		return
	}
	if !m.scheduler.IsRunning() {
		err = status.Errorf(codes.NotFound, "Monitoring of service %s has been stopped already", req.ServiceId)
		return
	}
	// Stop scheduler
	srv.monitoring[req.ServiceId].scheduler.Stop()

	res = &configuration.StopMonitoringResponse{}
	return
}

// calculateInterval calculates the interval as the smallest interval of all metric intervals
func (srv *Server) calculateInterval(metricIDs []string) (interval int) {
	var (
		err     error
		isFirst = true
	)

	for _, id := range metricIDs {
		var metric *assessment.Metric
		// Ensure metrics are retrieved only when having defined a valid interval
		err = srv.storage.Get(&metric, "id = ? AND NOT Interval <= 0", id)
		if err != nil {
			log.Tracef("No metric %s found with interval > 0 , will ignore it for interval calculation: %v", id, err)
			continue
		}
		// Set interval of first metric
		if isFirst {
			interval = int(metric.Interval)
			isFirst = false
			continue
		}
		// If metrics interval is smaller, use it as the new interval
		if int(metric.Interval) < interval {
			interval = int(metric.Interval)
		}
	}

	// If no metric had an interval (== zero value), use pre-defined interval in Server (Default one or set via Option)
	if interval == 0 {
		interval = srv.interval
		log.Infof("No interval given by any metric. Using pre-defined interval: %d", interval)
	}
	return
}

// triggerComplianceCalculation triggers the compliance calculation at the evaluation manager. TODO(oxisto): we should
// re-use connections instead of creating new ones
func (srv *Server) triggerComplianceCalculation(serviceID string, controlIDs []string) {
	log.Infof("Triggering compliance calculation for `%s` and controls: [%v]", serviceID, controlIDs)

	// Create connection to the evaluation manager
	conn, err := grpc.Dial(srv.evalManagerAddress, api.DefaultGrpcDialOptions(srv.evalManagerAddress, srv)...)
	if err != nil {
		log.Errorf("Could not dial to `%s`: %v", srv.evalManagerAddress, err)
		return
	}

	client := evaluation.NewEvaluationClient(conn)
	_, err = client.CalculateCompliance(context.TODO(), &evaluation.CalculateComplianceRequest{
		ServiceId:  serviceID,
		ControlIds: controlIDs,
	})
	if err != nil {
		log.Errorf("Could not trigger compliance calculation for `%s`: %v", serviceID, err)
	}
}

func (srv *Server) handleIncomingAssessmentResults(result *assessment.AssessmentResult, err error) {
	var key = fmt.Sprintf("%s-%s", result.ServiceId, result.MetricId)

	// We not want to trigger the compliance calculation for every result. We need to have some kind of algorithm here.
	// Basically, we want to have two things:
	//
	// - First, we accrue assessment results for a particular service and metric and only increment a counter. Once the
	// counter reaches a certain threshold, we actually trigger, otherwise we just increment the counter
	//
	// - Second, if we do not reach the desired counter within a time threshold, we also trigger, to not leave results hanging
	var w *complianceCalcWindow
	var ok bool
	w, ok = srv.ccw[key]
	if !ok {
		w = &complianceCalcWindow{start: time.Now()}
		srv.ccw[key] = w
	}

	if w.counter.Load() == 0 {
		// We are the first to arrive, so we start a go-routine that will finish the job in any case in delay time
		go func() {
			log.Debugf("Starting new compliance calculation window for service %s and metric %s", result.ServiceId, result.MetricId)
			time.Sleep(windowSize)

			srv.actuallyTrigger(result.ServiceId, result.MetricId)
			w.counter.Store(0)
		}()
	}

	w.counter.Add(1)

	if w.counter.Load() >= threshold {
		srv.actuallyTrigger(result.ServiceId, result.MetricId)
		w.counter.Store(0)
	}
}

func (srv *Server) actuallyTrigger(serviceID, metricID string) {
	var err error

	var service *orchestrator.CloudService
	var control *orchestrator.Requirement
	// Every time, we get an incoming assessment result, let's check if the control is monitoring and if yes, then
	// trigger the compliance calculation TODO(oxisto): We probably want to have some kind of intelligent algorithm here
	// that checks for X amount of results in a particular time or something like that and only then does a trigger,
	// because we often have a batch of assessment results coming in at more or less the same time.

	// Find control for metric
	control, err = srv.controlFor(metricID)
	if err != nil {
		log.Errorf("Could not find control for metric: %v", err)
		return
	}

	service, err = srv.GetCloudService(context.Background(), &orchestrator.GetCloudServiceRequest{ServiceId: serviceID})
	if err != nil {
		log.Errorf("Could not find cloud service: %v", err)
		return
	}

	if slices.Contains(service.Requirements.RequirementIds, control.Id) {
		srv.triggerComplianceCalculation(service.Id, []string{control.Id})
	}
}
