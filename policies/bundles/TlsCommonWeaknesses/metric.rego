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

package gxfs.metrics.tls_common_weaknesses

import data.clouditor.isIn

default applicable = false

default compliant = false

weaknesses := input.transportEncryption.tlsCommonVulns

applicable {
	weaknesses != null
}

compliant {
    # Compliant if none of the identified weaknesses match any of the weaknesses in data.target_value
	not isIn(data.operator, data.target_value, weaknesses)
}
