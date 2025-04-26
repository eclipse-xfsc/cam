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

package collection

import (
	"fmt"

	"github.com/eclipse-xfsc/cam/service/collection/workload/openstack"

	"github.com/google/uuid"
)

// Validate validates the StartCollectingRequest
func (req *StartCollectingRequest) Validate() error {
	// Check service ID
	err := checkServiceID(req.ServiceId)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrRequestServiceID, err)
	}

	// Check if EvalManager is set (non-empty)
	if req.EvalManager == "" {
		return ErrMissingEvalManager
	}

	if req.Configuration == nil {
		return ErrMissingServiceConfiguration
	}

	err = checkServiceID(req.Configuration.ServiceId)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrMissingServiceConfigurationServiceID, err)
	}

	if req.Configuration.RawConfiguration == nil {
		return ErrMissingRawConfiguration
	}

	// Check if configuration is available for the workload configuration collection module
	if !req.Configuration.RawConfiguration.MessageIs(&WorkloadSecurityConfig{}) {
		return ErrInvalidWorkloadConfigurationRawConfiguration
	}

	return nil
}

// KubeConfig returns the file content of the kube config file
func (req *StartCollectingRequest) KubeConfig() (config []byte, err error) {
	if req.Configuration == nil || req.Configuration.RawConfiguration == nil {
		err = fmt.Errorf("configuration not available: %w", err)
		return
	}

	// Get kube config from configuration
	var rawConfig WorkloadSecurityConfig
	err = req.Configuration.RawConfiguration.UnmarshalTo(&rawConfig)
	if err != nil {
		err = ErrInvalidWorkloadConfigurationRawConfiguration
		return
	}

	value := rawConfig.Kubernetes.AsInterface()

	if value == nil {
		err = fmt.Errorf("converting raw configuration to map is empty or nil")
		return
	}

	switch v := value.(type) {
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("got type %T but wanted string", v)
	}
}

// ConfigType contains the possible config types for the ToStruct method
type ConfigType interface {
	string | *openstack.AuthOptions
}

func checkServiceID(serviceID string) error {
	// Check if ServiceId is missing
	if serviceID == "" {
		return ErrMissingServiceID
	}

	// Check if ServiceId is valid
	if !IsValidUUID(serviceID) {
		return ErrInvalidServiceID
	}

	return nil
}

// Check if string is a valid UUID
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
