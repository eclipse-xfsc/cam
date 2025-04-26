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

package cam

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"

	"github.com/eclipse-xfsc/cam/api/configuration"
	"github.com/eclipse-xfsc/cam/api/evaluation"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ready chan bool

//go:embed dashboard/dist/*
var content embed.FS

type OAuth2Config struct {
	Authority             string `json:"authority"`
	ClientID              string `json:"client_id"`
	RedirectURI           string `json:"redirect_uri"`
	PostLogoutRedirectURI string `json:"post_logout_redirect_uri"`
}

// RunGateway starts a new gRPC gateway connecting the service specified in configurationEndpoint and evaluationEndpoint
// and exposes it on the httpPort specified.
func RunGateway(configurationEndpoint string,
	evaluationEndpoint string,
	config OAuth2Config,
	httpPort int, log *logrus.Entry) (err error) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		apiMux *runtime.ServeMux
		srv    *http.Server
		nl     net.Listener
		opts   []grpc.DialOption
	)

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	apiMux = runtime.NewServeMux()

	opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = configuration.RegisterConfigurationHandlerFromEndpoint(ctx, apiMux, configurationEndpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to connect to configuration service: %w", err)
	}

	log.Infof("Registered proxy to configuration service @ %s", configurationEndpoint)

	err = evaluation.RegisterEvaluationHandlerFromEndpoint(ctx, apiMux, evaluationEndpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to connect to evaluation service: %w", err)
	}

	log.Infof("Registered proxy to evaluation service @ %s", evaluationEndpoint)

	fsys, _ := fs.Sub(content, "dashboard/dist")

	fs := http.FileServer(http.FS(fsys))

	mux := http.NewServeMux()
	mux.Handle("/v1/", apiMux)
	mux.HandleFunc("/config.json", func(w http.ResponseWriter, r *http.Request) {
		err = json.NewEncoder(w).Encode(&config)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
	})
	mux.Handle("/", fs)

	srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: mux,
	}

	log.Printf("Starting API gateway on :%d", httpPort)

	nl, err = net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}

	go func() {
		ready <- true
	}()

	return srv.Serve(nl)
}
