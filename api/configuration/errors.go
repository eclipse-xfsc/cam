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

import "errors"

var (
	// ErrControlIDsEmpty indicates the request doesn't include any control IDs
	ErrControlIDsEmpty = errors.New("control IDs must be specified")
	// ErrRequestEmpty indicates the request is empty
	ErrRequestEmpty = errors.New("request is empty")
	// ErrServiceIDMissing indicates the request doesn't include a service ID
	ErrServiceIDMissing = errors.New("serviceID is missing")
)
