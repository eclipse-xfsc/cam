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

package xfsc.metrics.cyber_security_certification

import data.clouditor.compare

default applicable = false

default compliant = false

status := input.cyberSecurityCertification.status

applicable {
	status != null
}

compliant {
	compare(data.operator, data.target_value, status)
}
