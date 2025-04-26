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
	"testing"

	"clouditor.io/clouditor/api/orchestrator"
	"clouditor.io/clouditor/persistence/gorm"
	"github.com/cucumber/godog"
	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/configuration"
	"github.com/eclipse-xfsc/cam/internal/testutil"
)

type configurationFeature struct {
	*Server
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			NoColors: false,
			Paths:    []string{"../../features/configuration_interface.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func (*configurationFeature) theConfigurationInterface() error {
	return nil
}

func (c *configurationFeature) theFollowingServicesExist(services *godog.Table) error {
	// Add the services to the storage
	for _, row := range services.Rows {
		err := c.storage.Create(&orchestrator.CloudService{
			Id:          row.Cells[0].Value,
			Name:        row.Cells[1].Value,
			Description: row.Cells[2].Value,
		})
		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}
	}

	return nil
}

func (c *configurationFeature) theUserAccesses(in context.Context, endpoint string) (ctx context.Context, err error) {
	return c.theUserAccessesWithSetTo(in, endpoint, "", "")
}

func (c *configurationFeature) theUserAccessesWithSetTo(in context.Context, endpoint string, key string, value string) (ctx context.Context, err error) {
	switch endpoint {
	case "/v1/configuration/cloud_services":
		res, err := c.ListCloudServices(in, &orchestrator.ListCloudServicesRequest{})
		return testutil.ToContext(in, res), err
	case "/v1/configuration/cloud_services/{service_id}/configurations":
		if key != "service_id" {
			return nil, errors.New("wrong key")
		}
		res, err := c.ListCloudServiceConfigurations(in, &configuration.ListCloudServiceConfigurationsRequest{ServiceId: value})
		return testutil.ToContext(in, res), err
	}

	return in, nil
}

func (*configurationFeature) itShouldPresentAListOfMonitorableServices(in context.Context) error {
	var res = testutil.FromContext[orchestrator.ListCloudServicesResponse](in)

	if res == nil {
		return errors.New("empty response")
	}

	if len(res.Services) == 0 {
		return errors.New("no services listed in response")
	}

	return nil
}

func (*configurationFeature) heCanConfigureTheCloudService(in context.Context, cloudServiceID string) error {
	var res = testutil.FromContext[configuration.ListCloudServiceConfigurationsResponse](in)

	if res == nil {
		return errors.New("empty response")
	}

	// TODO(oxisto): Use a better assert
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	storage, _ := gorm.NewStorage(gorm.WithInMemory(),
		gorm.WithAdditionalAutoMigration(
			collection.ServiceConfiguration{},
		),
	)

	feature := &configurationFeature{
		Server: NewServer(WithStorage(storage)),
	}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// maybe for something else later
		return ctx, nil
	})

	ctx.Step(`^the following services exist:$`, feature.theFollowingServicesExist)
	ctx.Step(`^the configuration interface$`, feature.theConfigurationInterface)
	ctx.Step(`^the user accesses "([^"]*)" with "([^"]*)" set to "([^"]*)"$`, feature.theUserAccessesWithSetTo)
	ctx.Step(`^the user accesses "([^"]*)"$`, feature.theUserAccesses)
	ctx.Step(`^it should present a list of monitorable services$`, feature.itShouldPresentAListOfMonitorableServices)
	ctx.Step(`^he can configure the cloud service "([^"]*)"$`, feature.heCanConfigureTheCloudService)
}
