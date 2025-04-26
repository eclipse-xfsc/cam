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

	"clouditor.io/clouditor/api/orchestrator"

	"github.com/eclipse-xfsc/cam/oscal"
)

// ListControls is a wrapper around Clouditor orchestrator
func (s *Server) ListControls(ctx context.Context, req *orchestrator.ListRequirementsRequest) (*orchestrator.ListRequirementsResponse, error) {
	return s.OrchestratorServer.ListRequirements(ctx, req)
}

// loadRequirements loads requirements (or in Gaia-X speach "controls").
func loadRequirements() (requirements []*orchestrator.Requirement, err error) {
	var (
		file = "gxfs.json"
	)

	log.Infof("Loading OSCAL catalog from %s", file)

	return oscal.LoadRequirements(file)
}
