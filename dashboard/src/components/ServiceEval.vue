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
# Fraunhofer AISEC

* Show the evaluation result for one service with all controls
<script setup lang="ts">
import ControlEval from "../components/ControlEval.vue";
import PieChart from "../components/PieChart.vue";
import { CloudService, Compliance, Control } from "../data/model";
import * as data from "../data/newdata";

interface EvalProps {
  service: CloudService;
  // TODO(oxisto): should be converted to a global store or something similar
  controls: Control[];
  day?: number;
}

const props = defineProps<EvalProps>();
const compliance = await data.calculateCompliance(props.service, props.day);

function controlFor(controlId: string): Control {
  return (
    props.controls.find((control) => control.id == controlId) ?? {
      id: controlId,
      name: "Unknown",
      metrics: [],
    }
  );
}
</script>

<template>
  <div class="card">
    <div class="card-body">
      <h5 class="card-title">
        <RouterLink to="/evaluation/{{service.id}}">{{ service.name }}</RouterLink>
      </h5>
      <div class="text-muted">{{ service.description }}</div>
      <span v-if="compliance.status == data.ComplianceStatus.Compliant" class="text-success">Compliant</span>
      <span v-else-if="compliance.status == data.ComplianceStatus.NonComplaint" class="text-danger">Not Compliant</span>
      <span v-else style="color: gray">Pending</span>

      <PieChart v-if="compliance.chartdata" :chartData="compliance.chartdata" :width="200" :height="200" />

      <!-- Show result for each control of this service -->
      <p v-for="controldId in props.service.requirements?.requirementIds ?? []" class="mt-2">
        <ControlEval :service="service" :control="controlFor(controldId)" :result="compliance.latest.get(controldId)" />
      </p>
    </div>
  </div>
</template>

<style>
</style>