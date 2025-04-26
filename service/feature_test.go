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

package service_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/configuration"
	"github.com/eclipse-xfsc/cam/internal/testutil"
	"golang.org/x/oauth2/clientcredentials"

	"clouditor.io/clouditor/api"
	"clouditor.io/clouditor/api/orchestrator"
	"clouditor.io/clouditor/persistence"
	"clouditor.io/clouditor/persistence/gorm"
	"clouditor.io/clouditor/service"

	"github.com/cucumber/godog"
	"github.com/golang-jwt/jwt/v4"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	oauth2 "github.com/oxisto/oauth2go"
	configurationservice "github.com/eclipse-xfsc/cam/service/configuration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// techFeature describes technical requirements and we need to document them how we fulfill them in a seperate document
// rather than just executing tests. However, for future use we also prepare some godog functions.
type techFeature struct{}

func (*techFeature) aCommonDatabaseLayerShouldBeUsed() error {
	// Implemented using Clouditor persistence layer
	return nil
}

func (*techFeature) anyComponent(in context.Context) (context.Context, error) {
	// Return any gRPC component
	return testutil.ToContext(in, configurationservice.NewServer()), nil
}

func (*techFeature) persistanceIsNeeded() error {
	// Always needed
	return nil
}

// interfacesFeature describes technical requirements and we need to document them how we fulfill them in a seperate document
// rather than just executing tests. However, for future use we also prepare some godog functions.
type interfacesFeature struct {
	exampleFunc string
}

func (i *interfacesFeature) aGRPCGatewayMustBeUsed(in context.Context) error {
	var svc = testutil.FromContext[configurationservice.Server](in)

	// Try to use gRPC-gateway
	var mux = runtime.NewServeMux()
	err := configuration.RegisterConfigurationHandlerServer(context.TODO(), mux, svc)
	if err != nil {
		return errors.New("could not use gRPC gateway for component")
	}

	var w = httptest.NewRecorder()
	var r = &http.Request{
		Method: "GET",
		URL:    &url.URL{Scheme: "http", Host: "localhost", Path: i.exampleFunc},
	}
	mux.ServeHTTP(w, r)

	if w.Code != 200 {
		return errors.New("RPC call is not correctly exposed as REST")
	}

	return nil
}

func (*interfacesFeature) anRPCCallMustBeUsed() error {
	// Not checkable in source code
	return nil
}

func (*interfacesFeature) anyInterface(in context.Context) (context.Context, error) {
	// Return any gRPC interface implementation
	return testutil.ToContext(in, configurationservice.NewServer()), nil
}

func (*interfacesFeature) gRPCMustBeUsed(in context.Context) error {
	// The component must be registerable as a gRPC server
	var svc = testutil.FromContext[configurationservice.Server](in)

	srv := grpc.NewServer()

	// This will panic if it is not registrable
	configuration.RegisterConfigurationServer(srv, svc)

	return nil
}

func (*interfacesFeature) rPCMechanismsAreNeeded() error {
	// Always needed
	return nil
}

func (*interfacesFeature) theComponentHasEvents() error {
	// Not checkable in source code
	return nil
}

func (i *interfacesFeature) theRPCCallMustBeExposedAsREST() error {
	// Pick a random call that needs to exposed as REST
	i.exampleFunc = "/v1/configuration/metrics"

	return nil
}

type authFeature struct {
	grpcPort   uint16
	authPort   uint16
	srv        *grpc.Server
	authSrv    *oauth2.AuthorizationServer
	authConfig *service.AuthConfig

	key *ecdsa.PrivateKey
}

type response struct {
	msg proto.Message
	err error
}

func (a *authFeature) anyGRPCInterface() error {
	var storage persistence.Storage
	var svc configuration.ConfigurationServer
	var lis net.Listener
	var authLis net.Listener
	var err error

	// Generate a new key
	a.key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	// Use an internal OAuth 2.0 server for testing
	a.authSrv = oauth2.NewServer(":0", oauth2.WithClient("client", "secret", ""), oauth2.WithSigningKeysFunc(func() (keys map[int]*ecdsa.PrivateKey) {
		return map[int]*ecdsa.PrivateKey{
			0: a.key,
		}
	}))
	authLis, err = net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	a.authPort = authLis.Addr().(*net.TCPAddr).AddrPort().Port()

	storage, _ = gorm.NewStorage(gorm.WithInMemory(),
		gorm.WithAdditionalAutoMigration(
			collection.ServiceConfiguration{},
		),
	)

	svc = configurationservice.NewServer(configurationservice.WithStorage(storage))

	// Create any service and expose it via a gRPC interface on a random port
	lis, err = net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	a.grpcPort = lis.Addr().(*net.TCPAddr).AddrPort().Port()

	// Configure gRPC authentication
	a.authConfig = service.ConfigureAuth(service.WithJWKSURL(fmt.Sprintf("http://localhost:%d/certs", a.authPort)))
	a.srv = grpc.NewServer(grpc_middleware.WithUnaryServerChain(
		grpc_auth.UnaryServerInterceptor(a.authConfig.AuthFunc),
	), grpc_middleware.WithStreamServerChain(
		grpc_auth.StreamServerInterceptor(a.authConfig.AuthFunc),
	))
	configuration.RegisterConfigurationServer(a.srv, svc)

	go a.srv.Serve(lis)
	go a.authSrv.Serve(authLis)

	return nil
}

func (a *authFeature) aTokenWithExpirationOlderThanHArrives(in context.Context, hours int) (context.Context, error) {
	// Create an expired token
	t := jwt.NewWithClaims(jwt.SigningMethodES256, &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(-hours) * time.Hour)),
	})

	accessToken, err := t.SigningString()
	if err != nil {
		return nil, err
	}

	// We can use StaticTokenSource because it will never refresh our (invalid) token
	authorizer := &mockAuthorizer{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	})}

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", a.grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(authorizer),
	)
	if err != nil {
		return nil, err
	}

	client := configuration.NewConfigurationClient(conn)

	// Execute a command with authentication
	msg, err := client.ListMetrics(context.Background(), &orchestrator.ListMetricsRequest{})
	return testutil.ToContext(in, &response{msg, err}), nil
}

func (a *authFeature) aValidJWTArrives(in context.Context) (context.Context, error) {
	authorizer := api.NewOAuthAuthorizerFromClientCredentials(&clientcredentials.Config{
		ClientID:     "client",
		ClientSecret: "secret",
		TokenURL:     fmt.Sprintf("http://localhost:%d/token", a.authPort),
	})

	// To use a valid JWT, we can just make use of our api.Authorizer
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", a.grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(authorizer),
	)
	if err != nil {
		return nil, err
	}

	client := configuration.NewConfigurationClient(conn)

	// Execute a command without authentication
	msg, err := client.ListMetrics(context.Background(), &orchestrator.ListMetricsRequest{})
	return testutil.ToContext(in, &response{msg, err}), nil
}

func (a *authFeature) anUnauthorizedRequestArrives(in context.Context) (context.Context, error) {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", a.grpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := configuration.NewConfigurationClient(conn)

	// Execute a command without authentication
	msg, err := client.ListMetrics(context.Background(), &orchestrator.ListMetricsRequest{})
	return testutil.ToContext(in, &response{msg, err}), nil
}

func (*authFeature) noImplicitFlowMustTakePlace() error {
	return godog.ErrPending
}

func (*authFeature) theDashboard() error {
	return godog.ErrPending
}

func (*authFeature) theRequestMustBeFulfilled(in context.Context) error {
	var response = testutil.FromContext[response](in)

	if s, ok := status.FromError(response.err); ok {
		if s.Code() == codes.OK {
			return nil
		}
	}

	return fmt.Errorf("unexpected error: %w", response.err)
}

func (*authFeature) theRequestMustBeRejected(in context.Context) error {
	var response = testutil.FromContext[response](in)

	if s, ok := status.FromError(response.err); ok {
		if s.Code() == codes.Unauthenticated && s.Message() == "invalid auth token" {
			return nil
		}
	}

	return fmt.Errorf("unexpected error: %w", response.err)
}

func (*authFeature) theUserLogins() error {
	return godog.ErrPending
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format: "pretty",
			Paths: []string{
				"../features/tech.feature",
				"../features/interfaces.feature",
				"../features/auth.feature",
			},
			TestingT: t,
			Tags:     "Backend,Database",
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func (a *authFeature) shutdown(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	if a.srv != nil {
		a.srv.Stop()
	}

	if a.authSrv != nil {
		a.authSrv.Shutdown(context.Background())
	}

	if a.authConfig != nil && a.authConfig.Jwks != nil {
		a.authConfig.Jwks.EndBackground()
	}

	return nil, nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	techFeature := &techFeature{}

	ctx.Step(`^a common database layer should be used$`, techFeature.aCommonDatabaseLayerShouldBeUsed)
	ctx.Step(`^any component$`, techFeature.anyComponent)
	ctx.Step(`^persistance is needed$`, techFeature.persistanceIsNeeded)

	interfacesFeature := &interfacesFeature{}

	ctx.Step(`^a gRPC gateway must be used$`, interfacesFeature.aGRPCGatewayMustBeUsed)
	ctx.Step(`^an RPC call must be used$`, interfacesFeature.anRPCCallMustBeUsed)
	ctx.Step(`^any interface$`, interfacesFeature.anyInterface)
	ctx.Step(`^gRPC must be used$`, interfacesFeature.gRPCMustBeUsed)
	ctx.Step(`^RPC mechanisms are needed$`, interfacesFeature.rPCMechanismsAreNeeded)
	ctx.Step(`^the component has events$`, interfacesFeature.theComponentHasEvents)
	ctx.Step(`^the RPC call must be exposed as REST$`, interfacesFeature.theRPCCallMustBeExposedAsREST)

	authFeature := &authFeature{}

	ctx.After(authFeature.shutdown)
	ctx.Step(`^any gRPC interface$`, authFeature.anyGRPCInterface)
	ctx.Step(`^a token with expiration older than (\d+)h arrives$`, authFeature.aTokenWithExpirationOlderThanHArrives)
	ctx.Step(`^a valid JWT arrives$`, authFeature.aValidJWTArrives)
	ctx.Step(`^an unauthorized request arrives$`, authFeature.anUnauthorizedRequestArrives)
	ctx.Step(`^no implicit flow must take place$`, authFeature.noImplicitFlowMustTakePlace)
	ctx.Step(`^the dashboard$`, authFeature.theDashboard)
	ctx.Step(`^the request must be fulfilled$`, authFeature.theRequestMustBeFulfilled)
	ctx.Step(`^the request must be rejected$`, authFeature.theRequestMustBeRejected)
	ctx.Step(`^the user logins$`, authFeature.theUserLogins)
}

type mockAuthorizer struct {
	oauth2.TokenSource
}

func (p *mockAuthorizer) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	token, err := p.Token()
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"authorization": token.Type() + " " + token.AccessToken,
	}, nil
}

func (*mockAuthorizer) RequireTransportSecurity() bool {
	return false
}
