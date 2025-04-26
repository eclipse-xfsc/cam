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
	"sync/atomic"
	"time"

	"clouditor.io/clouditor/api"
	"clouditor.io/clouditor/api/assessment"
	"clouditor.io/clouditor/api/orchestrator"
	"clouditor.io/clouditor/persistence"
	orchestratorservice "clouditor.io/clouditor/service/orchestrator"

	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/eclipse-xfsc/cam/api/configuration"
	"github.com/eclipse-xfsc/cam/oscal"
	"github.com/eclipse-xfsc/cam/service"
)

const (
	// DefaultInterval sets the default interval in seconds for collecting evidences
	DefaultInterval    = 5 * 60
	DefaultMetricsFile = "metrics.json"
)

var (
	log = logrus.WithField("service", "configuration")
)

// Server is an implementation of the ConfigurationServer interface. Furthermore, it also implements a Clouditor
// Orchestrator, which is used to deal with non-Gaia-X specific functionalities, such as managing the target cloud
// services, metrics and controls (requirements).
type Server struct {
	configuration.UnimplementedConfigurationServer

	// piggyback a Clouditor Orchestrator to take care of metrics and cloud services
	orchestrator.OrchestratorServer

	// evalManagerAddress defines the target address (evaluation) for the collection modules
	evalManagerAddress string

	// interval defines the interval in minutes for collecting evidences
	interval int

	// storage is our storage backend
	storage persistence.Storage

	// authorizer is used to authenticate API calls to other services
	authorizer api.Authorizer

	// monitoring contains lists of controls (value) per service (key) that are currently monitored
	monitoring map[string]*MonitorScheduler

	ccw map[string]*complianceCalcWindow
}

type complianceCalcWindow struct {
	start   time.Time
	counter atomic.Int64
}

type MonitorScheduler struct {
	scheduler         *gocron.Scheduler
	monitoredControls []string
}

// WithEvalManagerAddress is a Server option setting the address for the evaluation manager
func WithEvalManagerAddress(url string) service.ServiceOption[Server] {
	return func(srv *Server) {
		srv.evalManagerAddress = url
	}
}

// WithInterval is a Server option setting the interval in minutes for collecting evidences
func WithInterval(interval int) service.ServiceOption[Server] {
	return func(srv *Server) {
		srv.interval = interval
	}
}

// WithStorage is an option to set the storage. If not set, NewServer will use inmemory storage.
func WithStorage(storage persistence.Storage) service.ServiceOption[Server] {
	return func(srv *Server) {
		srv.storage = storage
	}
}

// WithOAuth2Authorizer is an option to use an OAuth 2.0 authorizer
func WithOAuth2Authorizer(config *clientcredentials.Config) service.ServiceOption[Server] {
	return func(srv *Server) {
		srv.SetAuthorizer(api.NewOAuthAuthorizerFromClientCredentials(config))
	}
}

// SetAuthorizer implements UsesAuthorizer
func (srv *Server) SetAuthorizer(auth api.Authorizer) {
	srv.authorizer = auth
}

// Authorizer implements UsesAuthorizer
func (srv *Server) Authorizer() api.Authorizer {
	return srv.authorizer
}

// NewServer creates a new Server that implements the configuration interface.
func NewServer(opts ...service.ServiceOption[Server]) (srv *Server) {
	var (
		controls []*orchestrator.Requirement
		err      error
	)
	srv = &Server{
		interval:   DefaultInterval,
		monitoring: make(map[string]*MonitorScheduler),
		ccw:        make(map[string]*complianceCalcWindow),
	}

	// Apply any options
	for _, o := range opts {
		o(srv)
	}

	// Initialize in-memory storage backend if storage is not set already
	if srv.storage == nil {
		log.Errorf("Storage not initialized. Configuration Server will probably not function correctly: %v", err)

	}

	// Load controls (requirements in Clouditor terminology)
	log.Info("Loading controls")
	controls, err = loadRequirements()
	if err != nil {
		log.Errorf("Could not load controls (=requirements). Configuration Server will probably not function"+
			" correctly: %v", err)
	}
	// Create a new embedded Clouditor orchestrator, which takes care of the heavy lifting of metrics and controls
	srv.OrchestratorServer = orchestratorservice.NewService(
		orchestratorservice.WithExternalMetrics(func() ([]*assessment.Metric, error) {
			return oscal.LoadMetrics(DefaultMetricsFile)
		}),
		orchestratorservice.WithRequirements(controls),
		orchestratorservice.WithStorage(srv.storage),
	)

	// Create a hook function for incoming assessment results, so that we can trigger the compliance calculation
	srv.OrchestratorServer.(*orchestratorservice.Service).RegisterAssessmentResultHook(srv.handleIncomingAssessmentResults)

	if srv.evalManagerAddress == "" {
		log.Error("Address for Eval Manager not set: CMs will probably not work properly (It can be set via " +
			"`WithEvalManagerAddress` option)")
	}

	return
}
