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

	"clouditor.io/clouditor/api/assessment"
)

// loadMetrics loads metric definitions from a JSON file.
func LoadMetrics(file string) (metrics []*assessment.Metric, err error) {
	var (
		b []byte
	)

	b, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error while loading %s: %w", file, err)
	}

	err = json.Unmarshal(b, &metrics)
	if err != nil {
		return nil, fmt.Errorf("error in JSON marshal: %w", err)
	}

	return
}
