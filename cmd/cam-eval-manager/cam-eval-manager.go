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

	"golang.org/x/oauth2/clientcredentials"

	"github.com/eclipse-xfsc/cam/api/common"
	"github.com/eclipse-xfsc/cam/api/evaluation"
	"github.com/eclipse-xfsc/cam/internal/config"
	"github.com/eclipse-xfsc/cam/service"
	serviceEvaluation "github.com/eclipse-xfsc/cam/service/evaluation"

	"clouditor.io/clouditor/logging/formatter"
	"clouditor.io/clouditor/persistence"
	"clouditor.io/clouditor/persistence/gorm"
	clouditor_service "clouditor.io/clouditor/service"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	log   *logrus.Entry
	db    persistence.Storage
	types = []any{
		evaluation.EvaluationResult{},
		evaluation.Compliance{},
		common.Evidence{},
	}
	oAuthCred clientcredentials.Config
)

const (
	ConfigurationServiceAddressFlag = "configuration-service-address"
	DBUserFlag                      = "db-user"
	DBPasswordFlag                  = "db-password"
	DBHostFlag                      = "db-host"
	DBPortFlag                      = "db-port"
	DBNameFlag                      = "db-name"
	DBSSLModeFlag                   = "db-sslmode"
	DBInMemoryFlag                  = "db-in-memory"
	APIgRPCPortFlag                 = "api-grpc-port"
	OAuth2EndpointFlag              = "oauth2-token-endpoint"
	OAuth2ClientIDFlag              = "oauth2-client-id"
	OAuth2ClientSecretFlag          = "oauth2-client-secret"
	OAuth2ScopesFlag                = "oauth2-scopes"

	// APIJWKSURLFlag specifies the JWKS URL that is used to validate the incoming authentication tokens.
	APIJWKSURLFlag = "api-jwks-url"

	DefaultConfigurationServiceAddress        = "localhost:50100"
	DefaultAPIgRPCPort                 uint16 = 50101

	DefaultDBUser            = "postgres"
	DefaultDBPassword        = "postgres"
	DefaultDBHost            = "localhost"
	DefaultDBPort     uint16 = 5432
	DefaultDBName            = "postgres"
	DefaultSSLMode           = "require"
	DefaultInMemory          = false
)

func init() {
	log = logrus.WithField("component", "eval-manager")
	// TODO(oxisto): Set log level via cobra flag
	logrus.SetLevel(logrus.DebugLevel)
	log.Logger.Formatter = formatter.CapitalizeFormatter{Formatter: &logrus.TextFormatter{ForceColors: true}}

	cobra.OnInitialize(config.InitConfig)
}

func newEvalManagerCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "cam-eval-manager",
		Short: "cam-eval-manager launches the CAM Evaluation Manager",
		Long:  "The CAM Evaluation Manager takes care of evaluating evidence gathered by the collection modules.",
		RunE:  doCmd,
	}

	config.AddFlagString(cmd, ConfigurationServiceAddressFlag, DefaultConfigurationServiceAddress, "Specifies the address of the configuration service (cam-req-manager)")
	config.AddFlagString(cmd, DBUserFlag, DefaultDBUser, "Specifies the username of the database")
	config.AddFlagString(cmd, DBPasswordFlag, DefaultDBPassword, "Specifies the password of the database")
	config.AddFlagString(cmd, DBHostFlag, DefaultDBHost, "Specifies the hostname of the database")
	config.AddFlagUint16(cmd, DBPortFlag, DefaultDBPort, "Specifies the port of the database")
	config.AddFlagString(cmd, DBSSLModeFlag, DefaultSSLMode, "Specifies the sslmode of the database")
	config.AddFlagString(cmd, DBNameFlag, DefaultDBName, "Specifies the name of the database")
	config.AddFlagBool(cmd, DBInMemoryFlag, DefaultInMemory, "Specifies whether to use an in-memory database")
	config.AddFlagUint16(cmd, APIgRPCPortFlag, DefaultAPIgRPCPort, "Specifies the port used by the gRPC API")
	config.AddFlagString(cmd, APIJWKSURLFlag, "", "Specifies the JWKS URL that is used to validate the incoming authentication tokens. Setting this to empty will disable authentication (not recommended for production)")
	config.AddFlagString(cmd, OAuth2EndpointFlag, "", "Specifies the OAuth2 token URL that is used by the service to retrieve a token to authenticate with other services")
	config.AddFlagString(cmd, OAuth2ClientIDFlag, "", "Specifies the OAuth2 client ID that is used by the service to retrieve a token to authenticate with other services")
	config.AddFlagString(cmd, OAuth2ClientSecretFlag, "", "Specifies the OAuth2 client secret that is used by the service to retrieve a token to authenticate with other services")
	config.AddFlagStringSlice(cmd, OAuth2ScopesFlag, []string{}, "Specifies the OAuth2 scopes that are used by the service to retrieve a token to authenticate with other services")

	return cmd
}

func doCmd(_ *cobra.Command, _ []string) (err error) {
	var grpcOpts []grpc.ServerOption

	// Check for in-memory database
	if viper.GetBool(DBInMemoryFlag) {
		db, err = gorm.NewStorage(
			gorm.WithInMemory(),
			gorm.WithAdditionalAutoMigration(types...),
			gorm.WithMaxOpenConns(1),
		)
	} else {
		log.Infof("Connecting to storage %s@%s:%d/%s (ssl: %s)",
			viper.GetString(DBUserFlag),
			viper.GetString(DBHostFlag),
			uint16(viper.GetUint(DBPortFlag)),
			viper.GetString(DBNameFlag),
			viper.GetString(DBSSLModeFlag),
		)

		db, err = gorm.NewStorage(
			gorm.WithPostgres(
				viper.GetString(DBHostFlag),
				uint16(viper.GetUint(DBPortFlag)),
				viper.GetString(DBUserFlag),
				viper.GetString(DBPasswordFlag),
				viper.GetString(DBNameFlag),
				viper.GetString(DBSSLModeFlag),
			),
			gorm.WithAdditionalAutoMigration(types...),
		)
	}
	if err != nil {
		return fmt.Errorf("could not create storage: %w", err)
	}

	port := viper.GetInt32(APIgRPCPortFlag)

	// Specify port for client requests
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
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
		authConfig := clouditor_service.ConfigureAuth(clouditor_service.WithJWKSURL(jwks))
		defer authConfig.Jwks.EndBackground()

		grpcOpts = append(grpcOpts, grpc_middleware.WithUnaryServerChain(
			grpc_auth.UnaryServerInterceptor(authConfig.AuthFunc),
		), grpc_middleware.WithStreamServerChain(
			grpc_auth.StreamServerInterceptor(authConfig.AuthFunc),
		))
	}

	// Create gRPC Server (srv) and register the evaluation service (svc) on it
	var opts = []service.ServiceOption[serviceEvaluation.Server]{
		serviceEvaluation.WithStorage(db),
		serviceEvaluation.WithRequirementsManagerAddress(viper.GetString(ConfigurationServiceAddressFlag)),
	}
	if oAuthCred.TokenURL != "" {
		log.Infof("Configuring service with OAuth 2.0 using %s and client ID %s (scopes: %v)",
			oAuthCred.TokenURL, oAuthCred.ClientID, oAuthCred.Scopes)
		opts = append(opts, serviceEvaluation.WithOAuth2Authorizer(&oAuthCred))
	}

	srv := grpc.NewServer(grpcOpts...)
	svc := serviceEvaluation.NewServer(opts...)
	evaluation.RegisterEvaluationServer(srv, svc)

	// Enable reflection, primary for testing in early stages
	reflection.Register(srv)

	// Start to serve (blocks until process is killed or stopped)
	log.Infof("Starting gRPC server for Evaluation Manager on port: %d", port)
	if err = srv.Serve(lis); err != nil {
		log.Errorf("Failed to serve: %v", err)
		return err
	}

	return nil
}

func main() {
	var cmd = newEvalManagerCommand()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
