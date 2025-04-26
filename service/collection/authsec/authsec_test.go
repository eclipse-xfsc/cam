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

package authsec

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/internal/testutil/testevaluation"
	"github.com/eclipse-xfsc/cam/internal/testutil/testproto"
)

const (
	clientId     = "myAwesomeClient"
	clientSecret = "MyAwesomeSecret"
	apiToken     = "coolBearerToken"
)

var (
	tcpAddr *net.TCPAddr
)

// A simple HTTP server for testing
type httpServerRequestHandler struct{}

func (s httpServerRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infoln("Received Request for " + r.RequestURI)

	// OAuth 2.0 Metadata Document
	if r.RequestURI == "/.well-known/oauth-authorization-server" {
		resp, _ := json.Marshal(struct {
			Issuer                           string   `json:"issuer"`
			AuthorizationEndpoint            string   `json:"authorization_endpoint"`
			TokenEndpoint                    string   `json:"token_endpoint"`
			GrantTypesSupported              []string `json:"grant_types_supported"`
			ResponseTypesSupported           []string `json:"response_types_supported"`
			SubjectTypesSupported            []string `json:"subject_types_supported"`
			IdTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
		}{
			"http://" + tcpAddr.String(),
			"http://" + tcpAddr.String() + "/authorize",
			"http://" + tcpAddr.String() + "/token",
			[]string{"authorization_code", "client_credentials"},
			[]string{"code"},
			[]string{"public"},
			[]string{"RS256"},
		})
		w.Header().Add("Content-Type", "application/json")
		w.Write(resp)
		return
	}

	// OAuth 2.0 Token Endpoint
	if r.RequestURI == "/token" {
		// Accept Client Secret Post Authentication for the Client Credentials Grant. Ignore any scope values
		if clientId != r.PostFormValue("client_id") || clientSecret != r.PostFormValue("client_secret") {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("{\"error\":\"invalid_client\"}"))
			return
		}
		if r.PostFormValue("grant_type") != "client_credentials" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("{\"error\":\"unsupported_grant_type\"}"))
			return
		}
		scopes := r.PostFormValue("scope")
		resp, _ := json.Marshal(struct {
			AccessToken string `json:"access_token"`
			ExpiresIn   int    `json:"expires_in"`
			TokenType   string `json:"token_type"`
			Scope       string `json:"scope"`
		}{
			apiToken,
			60,
			"bearer",
			scopes,
		})
		w.Header().Add("Content-Type", "application/json")
		w.Write(resp)
		return
	}

	// Protected API Endpoint
	if r.RequestURI == "/api/protected-resource" {
		// Check the Authorization Header according to RFC 6750
		authHeaders := r.Header["Authorization"]
		if len(authHeaders) == 1 && ("Bearer "+apiToken) == authHeaders[0] {
			// authorized
			w.Write([]byte("Hello There :)"))
			return
		} else {
			// unauthorized
			w.Header().Add("WWW-Authenticate", "Bearer realm=\"Test\"")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	// Unprotected API Endpoint
	if r.RequestURI == "/api/unprotected-resource" {
		w.Write([]byte("Hello There :)"))
		return
	}

	w.WriteHeader(http.StatusNotFound)

}

func TestMain(m *testing.M) {

	// Setup logger
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.TraceLevel)
	log := logrus.WithField("testing", "collection-authsec")

	// Setup up dummy http server to collect information from
	addr := "127.0.0.1:0"
	log.Infof("Starting on %v", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("Failed to start server on %v: %v", addr, err)
		return
	}
	s := &http.Server{
		Handler: httpServerRequestHandler{},
	}
	log.Infof("Waiting for requests on %v", listener.Addr())

	// Get the tcp Address with dynamically chosen port
	tcpAddr = listener.Addr().(*net.TCPAddr)

	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	// Mock the Evaluation Manager to stream evidences to
	evaluationManagerServer, _, _ := testevaluation.StartBufConnServerToEvaluation()

	code := m.Run()

	evaluationManagerServer.Stop()

	os.Exit(code)
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	s := grpc.NewServer()

	collection.RegisterCollectionServer(s, NewServer(WithAdditionalGRPCOpts(grpc.WithContextDialer(testevaluation.BufConnDialer))))

	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestAcquireAccessToken(t *testing.T) {
	tests := []struct {
		name          string
		endpoint      string
		clientId      string
		clientSecret  string
		scopes        []string
		shouldFail    bool
		expectedToken string
	}{
		{
			name:          "Successful",
			endpoint:      "http://" + tcpAddr.String() + "/token",
			clientId:      clientId,
			clientSecret:  clientSecret,
			scopes:        []string{"Testscope"},
			shouldFail:    false,
			expectedToken: apiToken,
		},
		{
			name:         "Incorrect Client Secret",
			endpoint:     "http://" + tcpAddr.String() + "/token",
			clientId:     clientId,
			clientSecret: "invalid",
			scopes:       []string{"Testscope"},
			shouldFail:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare params
			metadata := make(map[string]interface{})
			metadata["token_endpoint"] = tt.endpoint // Ignoring other metadata

			// Call
			token, err := acquireAccessToken(&metadata, tt.clientId, tt.clientSecret, tt.scopes)

			// Check result
			if (nil != err) != tt.shouldFail {
				if nil != err {
					t.Errorf("Wrongfully received error: %s", err)
				} else {
					t.Error("Did not receive error")
				}
			}
			if !tt.shouldFail && token.AccessToken != tt.expectedToken {
				t.Fail()
			}

		})
	}
}

func TestCheckAPIAccess(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		endpoint       string
		token          string
		shouldFail     bool
		expectedReturn bool
	}{
		{
			name:           "Protected Authorized Access",
			method:         "GET",
			endpoint:       "http://" + tcpAddr.String() + "/api/protected-resource",
			token:          apiToken,
			shouldFail:     false,
			expectedReturn: true,
		},
		{
			name:           "Protected Unauthorized Access",
			method:         "GET",
			endpoint:       "http://" + tcpAddr.String() + "/api/protected-resource",
			token:          "",
			shouldFail:     false,
			expectedReturn: false,
		},
		{
			name:           "Unprotected Unauthorized Access",
			method:         "GET",
			endpoint:       "http://" + tcpAddr.String() + "/api/unprotected-resource",
			token:          "",
			shouldFail:     false,
			expectedReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare params
			endpoint, _ := url.Parse(tt.endpoint)
			var token *oauth2.Token = nil
			if tt.token != "" {
				token = &oauth2.Token{
					AccessToken: tt.token,
				}
			}

			// Call
			access, err := CheckAPIAccess(*endpoint, "GET", token)

			// Check result
			if (nil != err) != tt.shouldFail {
				if nil != err {
					t.Errorf("Wrongfully received error: %s", err)
				} else {
					t.Error("Did not receive error")
				}
			}
			if !tt.shouldFail && access != tt.expectedReturn {
				t.Fail()
			}

		})
	}
}

func TestStartCollecting(t *testing.T) {
	tests := []struct {
		name          string
		configuration *collection.ServiceConfiguration
		serviceId     string
		evalManager   string
		response      string
		wantErr       bool
	}{
		{
			name: "OAuthGrantTypes Success",
			configuration: &collection.ServiceConfiguration{
				ServiceId: "00000000-0000-0000-0000-000000000000",
				RawConfiguration: testproto.NewAny(t, &collection.AuthenticationSecurityConfig{
					Issuer:           "http://" + tcpAddr.String(),
					MetadataDocument: "",
				})},
			serviceId:   tcpAddr.String(),
			evalManager: "bufnet",
			response:    tcpAddr.String(),
			wantErr:     false,
		},
	}

	// Create gRPC Mock-Client for testing the server
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := collection.NewCollectionClient(conn)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := collection.StartCollectingRequest{
				ServiceId:     tt.serviceId,
				EvalManager:   tt.evalManager,
				Configuration: tt.configuration,
			}
			_, err = client.StartCollecting(ctx, &request)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
