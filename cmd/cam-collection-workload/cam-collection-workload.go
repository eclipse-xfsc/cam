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
	"net"
	"os"

	"clouditor.io/clouditor/logging/formatter"
	clouditor_service "clouditor.io/clouditor/service"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/internal/config"
	"github.com/eclipse-xfsc/cam/service"
	"github.com/eclipse-xfsc/cam/service/collection/workload"
)

var (
	log       *logrus.Entry
	oAuthCred clientcredentials.Config
)

const (
	DefaultGrpcPort = 50054
	// APIJWKSURLFlag specifies the JWKS URL that is used to validate the incoming authentication tokens.
	APIJWKSURLFlag         = "api-jwks-url"
	OAuth2EndpointFlag     = "oauth2-token-endpoint"
	OAuth2ClientIDFlag     = "oauth2-client-id"
	OAuth2ClientSecretFlag = "oauth2-client-secret"
	OAuth2ScopesFlag       = "oauth2-scopes"
)

func init() {
	log = logrus.WithField("component", "collection-workload")
	log.Logger.Formatter = formatter.CapitalizeFormatter{Formatter: &logrus.TextFormatter{ForceColors: true}}

	cobra.OnInitialize(config.InitConfig)
}

func newCollectionWorkloadCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "cam-collection-workload",
		Short: "cam-collection-workload launches the CAM Collection Workload Module",
		Long:  "The CAM Collection Workload collects the configuration information of deployed resources.",
		RunE:  doCmd,
	}

	config.AddFlagString(cmd, APIJWKSURLFlag, "", "Specifies the JWKS URL that is used to validate the incoming authentication tokens. Setting this to empty will disable authentication (not recommended for production)")
	config.AddFlagString(cmd, OAuth2EndpointFlag, "", "Specifies the OAuth2 token URL that is used by the service to retrieve a token to authenticate with other services")
	config.AddFlagString(cmd, OAuth2ClientIDFlag, "", "Specifies the OAuth2 client ID that is used by the service to retrieve a token to authenticate with other services")
	config.AddFlagString(cmd, OAuth2ClientSecretFlag, "", "Specifies the OAuth2 client secret that is used by the service to retrieve a token to authenticate with other services")
	config.AddFlagStringSlice(cmd, OAuth2ScopesFlag, []string{}, "Specifies the OAuth2 scopes that are used by the service to retrieve a token to authenticate with other services")

	return cmd
}

func doCmd(_ *cobra.Command, _ []string) (err error) {
	var grpcOpts []grpc.ServerOption

	log.Info("Start Workload Configuration...")

	// create a new socket for gRPC communication
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", DefaultGrpcPort))
	if err != nil {
		log.Errorf("could not listen: %v", err)
	}

	// Get Oauth2 token URL from environment variable
	oAuthCred.TokenURL = viper.GetString(OAuth2EndpointFlag)

	// Get Oauth2 client id from environment variable
	oAuthCred.ClientID = viper.GetString(OAuth2ClientIDFlag)

	// Get Oauth2 client secret from environment variable
	oAuthCred.ClientSecret = viper.GetString(OAuth2ClientSecretFlag)

	// Get Oauth2 scopes from environment variable
	oAuthCred.Scopes = viper.GetStringSlice(OAuth2ScopesFlag)

	jwks := viper.GetString(APIJWKSURLFlag)
	if jwks != "" {
		log.Infof("Configuring API with JWKS URL %s to validate tokens", jwks)
		authConfig := clouditor_service.ConfigureAuth(clouditor_service.WithJWKSURL(viper.GetString(APIJWKSURLFlag)))
		defer authConfig.Jwks.EndBackground()

		grpcOpts = append(grpcOpts, grpc_middleware.WithUnaryServerChain(
			grpc_auth.UnaryServerInterceptor(authConfig.AuthFunc),
		), grpc_middleware.WithStreamServerChain(
			grpc_auth.StreamServerInterceptor(authConfig.AuthFunc),
		))
	}

	var opts []service.ServiceOption[workload.Server]
	if oAuthCred.TokenURL != "" {
		log.Infof("Configuring service with OAuth 2.0 using %s and client ID %s (scopes: %v)",
			oAuthCred.TokenURL, oAuthCred.ClientID, oAuthCred.Scopes)
		opts = append(opts, workload.WithOAuth2Authorizer(&oAuthCred))
	}

	// Create gRPC Server (srv) and register workload configuration service (svc) on it
	srv := grpc.NewServer(grpcOpts...)
	svc := workload.NewServer(opts...)
	collection.RegisterCollectionServer(srv, svc)

	// Enable reflection, primary for testing in early stages
	reflection.Register(srv)

	// Start server (blocks until process is killed or stopped)
	log.Infof("Starting gRPC server for Workload Configuration CM on port: %d", DefaultGrpcPort)
	if err = srv.Serve(lis); err != nil {
		log.Fatalf("Workload Configuration CM: failed to serve: %v", err)
	}

	return nil
}

func main() {
	var cmd = newCollectionWorkloadCommand()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
