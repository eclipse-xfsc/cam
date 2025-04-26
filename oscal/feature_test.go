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

package oscal_test

import (
	"context"
	"errors"
	"testing"

	"clouditor.io/clouditor/api/assessment"
	"clouditor.io/clouditor/api/orchestrator"
	"github.com/cucumber/godog"
	"github.com/eclipse-xfsc/cam/oscal"
)

type metricsFeature struct {
	file         string
	metricsFile  string
	requirements []*orchestrator.Requirement
	metrics      []*assessment.Metric
}

func (o *metricsFeature) anOSCALModelOnFile(ctx context.Context) error {
	o.file = "../gxfs.json"
	o.metricsFile = "../metrics.json"

	return nil
}

func (o *metricsFeature) itIsLoaded() error {
	var err error

	// Load controls from OSCAL file
	o.requirements, err = oscal.LoadRequirements(o.file)
	if err != nil {
		return err
	}

	// Also load metrics
	o.metrics, err = oscal.LoadMetrics(o.metricsFile)
	if err != nil {
		return err
	}

	return nil
}

func (o *metricsFeature) itShouldContainRequirements() error {
	if len(o.requirements) == 0 {
		return errors.New("there are no requirements")
	}

	return nil
}

func (o *metricsFeature) theMetricShouldHaveScale(metricID string, scale int) error {
	for _, m := range o.metrics {
		if m.Id == metricID && m.Scale == assessment.Metric_Scale(scale) {
			return nil
		}
	}

	return errors.New("metric not found")
}

func (o *metricsFeature) theControlShouldHaveMetric(controlID, metricID string) error {
	for _, c := range o.requirements {
		if c.Id == controlID {
			for _, m := range c.Metrics {
				if m.Id == metricID {
					return nil
				}
			}
		}
	}

	return errors.New("metric not found")
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features/metrics.feature"},
			TestingT: t,
			Tags:     "Model",
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	feature := &metricsFeature{}

	ctx.Step(`^an OSCAL model on file$`, feature.anOSCALModelOnFile)
	ctx.Step(`^it is loaded$`, feature.itIsLoaded)
	ctx.Step(`^it should contain requirements$`, feature.itShouldContainRequirements)
	ctx.Step(`^the metric "([^"]*)" should have scale (\d+)$`, feature.theMetricShouldHaveScale)
	ctx.Step(`^the control "([^"]*)" should have metric "([^"]*)"$`, feature.theControlShouldHaveMetric)
}
