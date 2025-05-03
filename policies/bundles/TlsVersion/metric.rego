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

package xfsc.metrics.tls_version

import data.clouditor.compare
import data.clouditor.isIn

default compliant = false

default applicable = false

version := input.transportEncryption.tlsVersion

applicable {
	version != null
}

compliant {
	# If target_value is a list of strings/numbers. Currently it is enough that one of the (supported) versions matches
	# one of the target TLS Versions (e.g. supported TLS Versions [1.2, 1.3] is compliant for target value of [1.3])
	isIn(data.target_value, version)
}

compliant {
	# If target_value is the version number represented as int/float
	compare(data.operator, data.target_value, version)
}
