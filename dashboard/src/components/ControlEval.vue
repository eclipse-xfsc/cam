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

* Shows the evaluation result for one control
<script setup lang="ts">
import { CloudService, Compliance, Control, EvaluationResult } from "../data/model";
import * as data from "../data/newdata";
import MetricEval from "./MetricEval.vue";

interface ControlEvalProps {
  service: CloudService
  control: Control
  result?: Compliance
}

const props = defineProps<ControlEvalProps>();

function complianceClass(result: Compliance) {
  return {
    'bg-danger': data.isControlCompliant(result) == data.ComplianceStatus.NonComplaint,
    'bg-success': data.isControlCompliant(result) == data.ComplianceStatus.Compliant,
    'bg-secondary': data.isControlCompliant(result) == data.ComplianceStatus.Pending,
  }
}

function evaluationFor(compliance: Compliance, metricId: string): EvaluationResult | undefined {
  let result: EvaluationResult;

  for (let res of compliance.evaluations) {
    if (res.metricId == metricId) {
      return res
    }
  }

  return undefined;
}
</script>

<template>
  <div class="card h-100" style="width: 24rem">
    <div class="card-body control-card-body">
      <h6 class="card-subtitle mb-2 text-muted">{{ control.name }}</h6>
      <template v-if="result">
        <span class="dot mb-2" :class="complianceClass(result)"></span>
        <ul class="list-group">
          <template v-for="metric in control.metrics">
            <MetricEval :metric="metric" :service="service" :evaluation="evaluationFor(result, metric.id)" />
          </template>
        </ul>
      </template>
      <template v-else>
        <span class="dot" style="background-color: gray"></span>
        <ul class="list-group">
          <template v-for="metric in control.metrics">
            <MetricEval :metric="metric" :service="service" />
          </template>
        </ul>
      </template>
    </div>
  </div>
</template >

<style>
.dot {
  height: 25px;
  width: 25px;
  border-radius: 50%;
  display: inline-block;
}

.control-card-body {
  text-align: center;
}
</style>