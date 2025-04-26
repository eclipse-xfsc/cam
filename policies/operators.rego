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

package clouditor


# operator and target_value are declared here to add them to the output of each single policy (so assessment can use it)
operator = data.operator
target_value = data.target_value

compare(operator, target_value, actual_value) {
    operator == "=="
    target_value == actual_value
}{
    operator == ">="
    actual_value >= target_value
}{
     operator == "<="
     actual_value <= target_value
}{
     operator == "<"
     actual_value < target_value
}{
     operator == ">"
     actual_value > target_value
}


# Params: target_values (multiple target values), actual_value (single actual value)
isIn(target_values, actual_value) {
    # Assess actual value with each compliant value in target values
	actual_value == target_values[_]
}

# Params: target_values (multiple target values), actual_values (multiple actual values)
isIn(target_values, actual_values) {
    # Current implementation: It is enough that one output is one of target_values
    actual_values[_] == target_values[_]
}
