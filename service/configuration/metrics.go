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
	"fmt"

	"clouditor.io/clouditor/api/assessment"
	"clouditor.io/clouditor/api/orchestrator"
)

// ListMetrics is a wrapper around Clouditor orchestrator
func (s *Server) ListMetrics(ctx context.Context, req *orchestrator.ListMetricsRequest) (res *orchestrator.ListMetricsResponse, err error) {
	return s.OrchestratorServer.ListMetrics(ctx, req)
}

// GetMetric is a wrapper around Clouditor orchestrator
func (s *Server) GetMetric(ctx context.Context, req *orchestrator.GetMetricRequest) (res *assessment.Metric, err error) {
	return s.OrchestratorServer.GetMetric(ctx, req)
}

// GetMetricConfiguration is a wrapper around Clouditor orchestrator
func (s *Server) GetMetricConfiguration(ctx context.Context, req *orchestrator.GetMetricConfigurationRequest) (res *assessment.MetricConfiguration, err error) {
	return s.OrchestratorServer.GetMetricConfiguration(ctx, req)
}

// UpdateMetricConfiguration is a wrapper around Clouditor orchestrator
func (s *Server) UpdateMetricConfiguration(ctx context.Context, req *orchestrator.UpdateMetricConfigurationRequest) (res *assessment.MetricConfiguration, err error) {
	defer s.retriggerCollection(req.ServiceId, req.MetricId)

	return s.OrchestratorServer.UpdateMetricConfiguration(ctx, req)
}

func (s *Server) retriggerCollection(serviceId, metricId string) (err error) {
	var control *orchestrator.Requirement

	// Look for the control of this metric
	control, err = s.controlFor(metricId)
	if err != nil {
		return fmt.Errorf("error while retrieving control: %w", err)
	}

	// No control, nothing to do
	if control == nil {
		return nil
	}

	// Quickly check if we are monitoring at all
	monitor, ok := s.monitoring[serviceId]
	if !ok {
		return nil
	}

	// Otherwise, lets check if a collection module is monitoring our metric
	err = monitor.scheduler.RunByTag(metricId)
	if err != nil {
		return fmt.Errorf("error while running jobs: %w", err)
	}

	return nil
}

func (srv *Server) controlFor(metricId string) (control *orchestrator.Requirement, err error) {
	var controls []*orchestrator.Requirement

	// Retrieve all controls
	err = srv.storage.List(&controls, "id", true, 0, -1)
	if err != nil {
		err = fmt.Errorf("database error: %w", err)
		return
	}
	if len(controls) == 0 {
		err = errors.New("no controls found")
		return
	}

	// Look for one with the metric
	for _, control := range controls {
		for _, m := range control.Metrics {
			if m.Id == metricId {
				return control, nil
			}
		}
	}

	return nil, errors.New("no controls found")
}
