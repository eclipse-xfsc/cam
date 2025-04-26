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

package main

import (
	"fmt"
	"os"

	"github.com/eclipse-xfsc/cam"
	"github.com/eclipse-xfsc/cam/internal/config"

	"clouditor.io/clouditor/logging/formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var log *logrus.Entry

const (
	ConfigurationServiceAddressFlag = "configuration-service-address"
	EvaluationServiceAddressFlag    = "evaluation-service-address"

	OAuth2DashboardAuthority             = "oauth2-dashboard-authority"
	OAuth2DashboardClientID              = "oauth2-dashboard-client-id"
	OAuth2DashboardRedirectURI           = "oauth2-dashboard-redirect-uri"
	OAuth2DashboardPostLogoutRedirectURI = "oauth2-dashboard-post-logout-redirect-uri"

	DefaultConfigurationServiceAddress = "localhost:50100"
	DefaultEvaluationServiceAddress    = "localhost:50101"

	DefaultOAuth2DashboardAuthority             = "http://localhost:8000"
	DefaultOAuth2DashboardClientID              = "public"
	DefaultOAuth2DashboardRedirectURI           = "http://localhost:8080/#/loggedin"
	DefaultOAuth2DashboardPostLogoutRedirectURI = "http://localhost:8080/#/loggedout"
)

func init() {
	log = logrus.WithField("component", "grpc-gateway")
	log.Logger.Formatter = formatter.CapitalizeFormatter{Formatter: &logrus.TextFormatter{ForceColors: true}}

	cobra.OnInitialize(config.InitConfig)
}

func newGatewayCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "cam-api-gateway",
		Short: "cam-api-gateway launches the CAM API Gateway",
		Long:  "The CAM API Gateway serves as central gateway to expose the functionality of different CAM micro-service using a central REST API.",
		RunE:  doCmd,
	}

	config.AddFlagString(cmd, ConfigurationServiceAddressFlag, DefaultConfigurationServiceAddress, "Specifies the address of the configuration service (cam-req-manager)")
	config.AddFlagString(cmd, EvaluationServiceAddressFlag, DefaultEvaluationServiceAddress, "Specifies the address of the evaluation service (cam-eval-manager)")
	config.AddFlagString(cmd, OAuth2DashboardAuthority, DefaultOAuth2DashboardAuthority, "Specifies the OAuth 2.0 authority used for the dashboard")
	config.AddFlagString(cmd, OAuth2DashboardClientID, DefaultOAuth2DashboardClientID, "Specifies the OAuth 2.0 client ID used for the dashboard")
	config.AddFlagString(cmd, OAuth2DashboardRedirectURI, DefaultOAuth2DashboardRedirectURI, "Specifies the OAuth 2.0 redirect URI used for the dashboard")
	config.AddFlagString(cmd, OAuth2DashboardPostLogoutRedirectURI, DefaultOAuth2DashboardPostLogoutRedirectURI, "Specifies the OAuth 2.0 post logout redirect URI used for the dashboard")

	return cmd
}

func doCmd(cmd *cobra.Command, args []string) error {
	err := cam.RunGateway(
		viper.GetString(ConfigurationServiceAddressFlag),
		viper.GetString(EvaluationServiceAddressFlag),
		cam.OAuth2Config{
			Authority:             viper.GetString(OAuth2DashboardAuthority),
			ClientID:              viper.GetString(OAuth2DashboardClientID),
			RedirectURI:           viper.GetString(OAuth2DashboardRedirectURI),
			PostLogoutRedirectURI: viper.GetString(OAuth2DashboardPostLogoutRedirectURI),
		},
		8080, log)
	if err != nil {
		return fmt.Errorf("could not start gRPC-Gateway for REST: %w", err)
	}

	return nil
}

func main() {
	var cmd = newGatewayCommand()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
