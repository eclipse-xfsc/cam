# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Contributors:
#	Fraunhofer AISEC

package gxfs.metrics.o_auth_grant_types

import data.clouditor.isIn

default applicable = false

default compliant = false

# allowedTypes: Scale defined in SPEC
allowedTypes := ["authorization_code", "implicit", "password", "client_credentials", "device_code",
                "refresh_token"]
grantTypes := input.oAuth.grantTypes

applicable {
	grantTypes != null
}

compliant {
    # First check that provided grant types are valid (in scale)
	isIn(allowedTypes, grantTypes)
	# Then check that they are not forbidden (== not in data.target_value)
	# TODO(lebogg): Make this part of operators.rego (using data.operator to distinguish)
	not isIn(data.target_value, grantTypes)
}