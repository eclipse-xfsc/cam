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
	"errors"
)

var (
	ErrInvalidServiceID                             = errors.New("serviceID is invalid")
	ErrMissingServiceID                             = errors.New("serviceID is missing")
	ErrRequestServiceID                             = errors.New("serviceID in request is invalid")
	ErrMissingServiceConfigurationServiceID         = errors.New("serviceID in service configuration is invalid")
	ErrMissingEvalManager                           = errors.New("evaluation manager URL is missing")
	ErrMissingServiceConfiguration                  = errors.New("service configuration is missing")
	ErrMissingRawConfiguration                      = errors.New("service configuration is missing")
	ErrInvalidRemoteIntegrityRawConfiguration       = errors.New("no remote integrity raw configuration")
	ErrInvalidWorkloadConfigurationRawConfiguration = errors.New("no workload raw configuration")
	ErrInvalidKubernetesServiceConfiguration        = errors.New("kubernetes service configuration is invalid")
	ErrInvalidOpenstackServiceConfiguration         = errors.New("could not store openstack service configuration")
	ErrInvalidAWSServiceConfiguration               = errors.New("could not store aws service configuration")
	ErrConversionProtobufToByteArray                = errors.New("could not convert protobuf value to byte array")
	ErrKubernetesClientset                          = errors.New("could not get kubernetes clientset")
	ErrConversionProtobufToAuthOptions              = errors.New("could not convert protobuf value to openstack.authOptions")
)
