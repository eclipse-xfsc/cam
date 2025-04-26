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
	"context"
	"fmt"
	"net"
	"os"

	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/configuration"
	"github.com/eclipse-xfsc/cam/internal/config"
	"github.com/eclipse-xfsc/cam/internal/protobuf"
	"github.com/eclipse-xfsc/cam/service"
	service_configuration "github.com/eclipse-xfsc/cam/service/configuration"

	"clouditor.io/clouditor/api/assessment"
	"clouditor.io/clouditor/api/orchestrator"
	"clouditor.io/clouditor/logging/formatter"
	"clouditor.io/clouditor/persistence"
	"clouditor.io/clouditor/persistence/gorm"
	clouditor_service "clouditor.io/clouditor/service"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	log       *logrus.Entry
	db        persistence.Storage
	types     = []any{collection.CollectionModule{}, collection.ServiceConfiguration{}}
	oAuthCred clientcredentials.Config
)

const (
	EvaluationServiceAddressFlag = "evaluation-service-address"
	DBUserFlag                   = "db-user"
	DBPasswordFlag               = "db-password"
	DBHostFlag                   = "db-host"
	DBPortFlag                   = "db-port"
	DBNameFlag                   = "db-name"
	DBInMemoryFlag               = "db-in-memory"
	DBSSLModeFlag                = "db-sslmode"
	APIgRPCPortFlag              = "api-grpc-port"
	OAuth2EndpointFlag           = "oauth2-token-endpoint"
	OAuth2ClientIDFlag           = "oauth2-client-id"
	OAuth2ClientSecretFlag       = "oauth2-client-secret"
	OAuth2ScopesFlag             = "oauth2-scopes"

	// APIJWKSURLFlag specifies the JWKS URL that is used to validate the incoming authentication tokens.
	APIJWKSURLFlag = "api-jwks-url"

	// CollectionModuleAutoCreateFlag specifies whether collection modules should be auto-created at start-up. This is
	// useful, if deployed with Helm.
	CollectionModuleAutoCreateFlag = "collection-autocreate"

	// CollectionCommSecServiceHost, if CollectionModuleAutoCreate is specified, defines the host for a default
	// collection integrity module. In a Kubernetes cluster deployment with Helm, this will be auto-configured.
	CollectionCommSecServiceHostFlag = "collection-commsec-service-host"
	// CollectionCommSecServicePort, if CollectionModuleAutoCreate is specified, defines the port for a default
	// collection integrity module. In a Kubernetes cluster deployment with Helm, this will be auto-configured.
	CollectionCommSecServicePortFlag = "collection-commsec-service-port"
	// CollectionAuthsecServiceHost, if CollectionModuleAutoCreate is specified, defines the host for a default
	// collection authsec module. In a Kubernetes cluster deployment with Helm, this will be auto-configured.
	CollectionAuthSecServiceHostFlag = "collection-authsec-service-host"
	// CollectionAuthSecServicePort, if CollectionModuleAutoCreate is specified, defines the port for a default
	// collection authsec module. In a Kubernetes cluster deployment with Helm, this will be auto-configured.
	CollectionAuthSecServicePortFlag = "collection-authsec-service-port"
	// CollectionIntegrityServiceHost, if CollectionModuleAutoCreate is specified, defines the host for a default
	// collection integrity module. In a Kubernetes cluster deployment with Helm, this will be auto-configured.
	CollectionIntegrityServiceHostFlag = "collection-integrity-service-host"
	// CollectionIntegrityServicePort, if CollectionModuleAutoCreate is specified, defines the port for a default
	// collection integrity module. In a Kubernetes cluster deployment with Helm, this will be auto-configured.
	CollectionIntegrityServicePortFlag = "collection-integrity-service-port"
	// CollectionWorkloadServiceHost, if CollectionModuleAutoCreate is specified, defines the host for a default
	// collection integrity module. In a Kubernetes cluster deployment with Helm, this will be auto-configured.
	CollectionWorkloadServiceHostFlag = "collection-workload-service-host"
	// CollectionWorkloadServicePort, if CollectionModuleAutoCreate is specified, defines the port for a default
	// collection integrity module. In a Kubernetes cluster deployment with Helm, this will be auto-configured.
	CollectionWorkloadServicePortFlag = "collection-workload-service-port"

	DefaultCollectionModuleAutoCreate            = false
	DefaultCollectionCommSecServiceHost          = "localhost"
	DefaultCollectionCommSecServicePort   uint16 = 50051
	DefaultCollectionAuthSecServiceHost          = "localhost"
	DefaultCollectionAuthSecServicePort   uint16 = 50052
	DefaultCollectionIntegrityServiceHost        = "localhost"
	DefaultCollectionIntegrityServicePort uint16 = 50053
	DefaultCollectionWorkloadServiceHost         = "localhost"
	DefaultCollectionWorkloadServicePort  uint16 = 50054

	// DefaultEvaluationServiceAddress sets the default target address (evaluation) for the collection modules
	DefaultEvaluationServiceAddress = "localhost:50101"

	DefaultDBUser            = "postgres"
	DefaultDBPassword        = "postgres"
	DefaultDBHost            = "localhost"
	DefaultDBPort     uint16 = 5432
	DefaultDBName            = "postgres"
	DefaultSSLMode           = "require"
	DefaultInMemory          = false

	// DefaultAPIgRPCPort sets the default port for the requirements manager
	DefaultAPIgRPCPort uint16 = 50100
)

func init() {
	log = logrus.WithField("component", "req-manager")
	// TODO(oxisto): Set log level via cobra flag
	logrus.SetLevel(logrus.DebugLevel)
	log.Logger.Formatter = formatter.CapitalizeFormatter{Formatter: &logrus.TextFormatter{ForceColors: true}}

	cobra.OnInitialize(config.InitConfig)
}

func newReqManagerCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "cam-req-manager",
		Short: "cam-req-manager launches the CAM Requirements Manager",
		Long:  "The CAM Requirements Manager takes care of configuring everything.",
		RunE:  doCmd,
	}

	config.AddFlagString(cmd, EvaluationServiceAddressFlag, DefaultEvaluationServiceAddress, "Specifies the address of the evaluation service (cam-eval-manager)")
	config.AddFlagString(cmd, DBUserFlag, DefaultDBUser, "Specifies the username of the database")
	config.AddFlagString(cmd, DBPasswordFlag, DefaultDBPassword, "Specifies the password of the database")
	config.AddFlagString(cmd, DBHostFlag, DefaultDBHost, "Specifies the hostname of the database")
	config.AddFlagUint16(cmd, DBPortFlag, DefaultDBPort, "Specifies the port of the database")
	config.AddFlagString(cmd, DBNameFlag, DefaultDBName, "Specifies the name of the database")
	config.AddFlagString(cmd, DBSSLModeFlag, DefaultSSLMode, "Specifies the sslmode of the database")
	config.AddFlagBool(cmd, DBInMemoryFlag, DefaultInMemory, "Specifies whether to use an in-memory database")
	config.AddFlagUint16(cmd, APIgRPCPortFlag, DefaultAPIgRPCPort, "Specifies the port used by the gRPC API")
	config.AddFlagString(cmd, APIJWKSURLFlag, "", "Specifies the JWKS URL that is used to validate the incoming authentication tokens. Setting this to empty will disable authentication (not recommended for production)")
	config.AddFlagString(cmd, OAuth2EndpointFlag, "", "Specifies the OAuth2 token URL that is used by the service to retrieve a token to authenticate with other services")
	config.AddFlagString(cmd, OAuth2ClientIDFlag, "", "Specifies the OAuth2 client ID that is used by the service to retrieve a token to authenticate with other services")
	config.AddFlagString(cmd, OAuth2ClientSecretFlag, "", "Specifies the OAuth2 client secret that is used by the service to retrieve a token to authenticate with other services")
	config.AddFlagStringSlice(cmd, OAuth2ScopesFlag, []string{}, "Specifies the OAuth2 scopes that are used by the service to retrieve a token to authenticate with other services")

	config.AddFlagBool(cmd, CollectionModuleAutoCreateFlag, DefaultCollectionModuleAutoCreate, "Specifies whether collection modules should be auto-created")
	config.AddFlagString(cmd, CollectionCommSecServiceHostFlag, DefaultCollectionCommSecServiceHost, "Specifies the host for a default collection commsec module")
	config.AddFlagUint16(cmd, CollectionCommSecServicePortFlag, DefaultCollectionCommSecServicePort, "Specifies the port for a default collection commsec module")
	config.AddFlagString(cmd, CollectionAuthSecServiceHostFlag, DefaultCollectionAuthSecServiceHost, "Specifies the host for a default collection authsec module")
	config.AddFlagUint16(cmd, CollectionAuthSecServicePortFlag, DefaultCollectionAuthSecServicePort, "Specifies the port for a default collection authsec module")
	config.AddFlagString(cmd, CollectionIntegrityServiceHostFlag, DefaultCollectionIntegrityServiceHost, "Specifies the host for a default collection integrity module")
	config.AddFlagUint16(cmd, CollectionIntegrityServicePortFlag, DefaultCollectionIntegrityServicePort, "Specifies the port for a default collection integrity module")
	config.AddFlagString(cmd, CollectionWorkloadServiceHostFlag, DefaultCollectionWorkloadServiceHost, "Specifies the host for a default collection workload module")
	config.AddFlagUint16(cmd, CollectionWorkloadServicePortFlag, DefaultCollectionWorkloadServicePort, "Specifies the port for a default collection workload module")

	return cmd
}

func doCmd(_ *cobra.Command, _ []string) (err error) {
	var grpcOpts []grpc.ServerOption

	log.Println("This is GXFS's Requirements Manager")

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

	port := uint16(viper.GetUint(APIgRPCPortFlag))

	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("could not listen: %v", err)
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

	// Create a gRPC Server (srv) and register configuration service (svc) on it. Additionally register a Clouditor
	// orchestrator service on it.
	var opts = []service.ServiceOption[service_configuration.Server]{
		service_configuration.WithStorage(db),
		service_configuration.WithEvalManagerAddress(viper.GetString(EvaluationServiceAddressFlag)),
	}
	if oAuthCred.TokenURL != "" {
		log.Infof("Configuring service with OAuth 2.0 using %s and client ID %s (scopes: %v)",
			oAuthCred.TokenURL, oAuthCred.ClientID, oAuthCred.Scopes)
		opts = append(opts, service_configuration.WithOAuth2Authorizer(&oAuthCred))
	}

	srv := grpc.NewServer(grpcOpts...)
	svc := service_configuration.NewServer(opts...)
	configuration.RegisterConfigurationServer(srv, svc)
	orchestrator.RegisterOrchestratorServer(srv, svc)

	// Enable reflection, primary for testing in early stages
	reflection.Register(srv)

	// TODO(anatheka): Missing communication security module in addCollectionModules()
	if viper.GetBool(CollectionModuleAutoCreateFlag) {
		log.Info("Adding collection modules...")
		autoCreateCollectionModules(svc)
	}

	log.Infof("Starting gRPC server for Requirements Manager on port %d ...", port)
	if err = srv.Serve(sock); err != nil {
		log.Fatalf("Failed to serve gRPC endpoint: %s", err)
		return err
	}

	return nil
}

func autoCreateCollectionModules(svc configuration.ConfigurationServer) {
	var (
		err error
		res *configuration.ListCollectionModulesResponse
		mod *collection.CollectionModule
	)

	// Remove existing ones
	res, err = svc.ListCollectionModules(context.TODO(), &configuration.ListCollectionModulesRequest{})
	if err != nil {
		log.Errorf("Could not list collection modules: %v", err)
		return
	}

	for _, mod = range res.Modules {
		_, _ = svc.RemoveCollectionModule(context.TODO(), &configuration.RemoveCollectionModuleRequest{ModuleId: mod.Id})
	}

	// Add Workload Configuration Collection Module
	mod = &collection.CollectionModule{
		Id:      config.DefaultCollectionWorkloadID,
		Name:    "Workload Security",
		Metrics: []*assessment.Metric{{Id: "AtRestEncryption"}},
		Address: fmt.Sprintf("%s:%d", viper.GetString(CollectionWorkloadServiceHostFlag),
			viper.GetUint(CollectionWorkloadServicePortFlag)),
		ConfigMessageTypeUrl: protobuf.TypeURL(&collection.WorkloadSecurityConfig{}),
	}
	_, err = svc.AddCollectionModule(context.TODO(), &configuration.AddCollectionModuleRequest{Module: mod})
	if err != nil {
		log.Errorf("Could not add workload security collection module: %v", err)
	} else {
		log.Infof("Added workload security collection module (address: %s)", mod.Address)
	}

	// Add Authentication Security Test Collection Module
	mod = &collection.CollectionModule{
		Id:      config.DefaultCollectionAuthSecID,
		Name:    "Authentication Security",
		Metrics: []*assessment.Metric{{Id: "OAuthGrantTypes"}, {Id: "APIOAuthProtected"}},
		Address: fmt.Sprintf("%s:%d", viper.GetString(CollectionAuthSecServiceHostFlag),
			viper.GetUint(CollectionAuthSecServicePortFlag)),
		ConfigMessageTypeUrl: protobuf.TypeURL(&collection.AuthenticationSecurityConfig{}),
	}
	_, err = svc.AddCollectionModule(context.TODO(), &configuration.AddCollectionModuleRequest{Module: mod})
	if err != nil {
		log.Errorf("Could not add authentication security collection module: %v", err)
	} else {
		log.Infof("Added authentication security collection module (address: %s)", mod.Address)
	}

	// Add Remote Integrity Collection Module
	mod = &collection.CollectionModule{
		Id:      config.DefaultCollectionIntegrityID,
		Name:    "Remote Integrity",
		Metrics: []*assessment.Metric{{Id: "SystemComponentsIntegrity"}},
		Address: fmt.Sprintf("%s:%d", viper.GetString(CollectionIntegrityServiceHostFlag),
			viper.GetUint(CollectionIntegrityServicePortFlag)),
		ConfigMessageTypeUrl: protobuf.TypeURL(&collection.RemoteIntegrityConfig{}),
	}
	_, err = svc.AddCollectionModule(context.TODO(), &configuration.AddCollectionModuleRequest{Module: mod})
	if err != nil {
		log.Errorf("Could not add remote integrity collection module: %v", err)
	} else {
		log.Infof("Added remote integrity collection module (address: %s)", mod.Address)
	}

	// Add Communication Security Test Collection Module
	mod = &collection.CollectionModule{
		Id:      config.DefaultCollectionCommsecID,
		Name:    "Communication Security",
		Metrics: []*assessment.Metric{{Id: "TlsVersion"}, {Id: "TlsCipherSuite"}, {Id: "TlsCommonWeaknesses"}},
		Address: fmt.Sprintf("%s:%d", viper.GetString(CollectionCommSecServiceHostFlag),
			viper.GetUint(CollectionCommSecServicePortFlag)),
		ConfigMessageTypeUrl: protobuf.TypeURL(&collection.CommunicationSecurityConfig{}),
	}
	_, err = svc.AddCollectionModule(context.TODO(), &configuration.AddCollectionModuleRequest{Module: mod})
	if err != nil {
		log.Errorf("Could not add communication security collection module: %v", err)
	} else {
		log.Infof("Added communication security collection module (address: %s)", mod.Address)
	}
}

func main() {
	var cmd = newReqManagerCommand()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
