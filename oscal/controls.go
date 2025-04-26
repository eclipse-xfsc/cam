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

package oscal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"clouditor.io/clouditor/api/assessment"
	"clouditor.io/clouditor/api/orchestrator"
)

// LoadRequirements loads requirements (or in Gaia-X speach "controls") from a file
func LoadRequirements(file string) (requirements []*orchestrator.Requirement, err error) {
	var (
		b []byte
	)

	var outer struct {
		Catalog Catalog `json:"catalog"`
	}

	b, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error while loading %s: %w", file, err)
	}

	err = json.Unmarshal(b, &outer)
	if err != nil {
		return nil, fmt.Errorf("error in JSON marshal: %w", err)
	}

	// Loop through controls
	for _, control := range outer.Catalog.Controls {
		// And create a requirement for it
		var r = orchestrator.Requirement{
			Id:       control.ID,
			Name:     control.Title,
			Metrics:  metricFor(&control),
			Category: "EUCS",
		}

		requirements = append(requirements, &r)
	}

	return requirements, nil
}

func metricFor(control *Control) (metrics []*assessment.Metric) {
	for _, prop := range control.Props {
		// Look for "metrics" property
		if prop.Name == "metrics" {
			// Remove white-spaces, e.g. "TlsVersion, TlsCipherSuite" -> "TlsVersion,TlsCipherSuite"
			var value = strings.ReplaceAll(prop.Value, " ", "")
			var metricIDs = strings.Split(value, ",")
			for _, metricID := range metricIDs {
				metrics = append(metrics, &assessment.Metric{
					Id: metricID,
				})
			}
		}
	}

	return
}
