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

package strct

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

type AWSConfig struct {
	Region          string `json:"region"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

// ToConfig converts the protobuf value to an AWS Config
func ToConfig(v *structpb.Value) (config *AWSConfig, err error) {
	value := v.GetStructValue().AsMap()

	if len(value) == 0 {
		err = fmt.Errorf("converting raw configuration to map is empty or nil")
		return
	}

	// First, we have to marshal the configuration map
	body, err := json.Marshal(value)
	if err != nil {
		err = fmt.Errorf("could not marshal configuration")
		return
	}

	// Then, we can store it back to the gophercloud.AuthOptions
	if err = json.Unmarshal(body, &config); err != nil {
		err = fmt.Errorf("could not parse configuration: %w", err)
		return
	}

	return
}
