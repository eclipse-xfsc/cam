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
  startDay: Number;
  endDay: Number;
}

const props = defineProps<EvalProps>();
let compliance = null;
if (props.startDay) {
  compliance = await data.calculateCompliance(
    props.service,
    Number(props.startDay),
    Number(props.endDay)
  );
} else {
  compliance = await data.calculateCompliance(props.service);
}
</script>

<template>
  <div v-if="!compliance.latest">
    <p>No data for this time period.</p>
  </div>

  <div v-else>
    <div class="card">
      <div class="card-body">
        <PieChart
          v-if="compliance.chartdata"
          :chartData="compliance.chartdata"
          :width="150"
          :height="150"
          style="float: right"
        />
        <h5 class="card-title">
          <RouterLink
            v-if="startDay"
            :to="
              '/evaluation/' +
              service.id +
              '?startDay=' +
              startDay +
              '&endDay=' +
              endDay
            "
          >
            {{ service.name }}</RouterLink
          >
          <RouterLink v-else :to="'/evaluation/' + service.id">
            {{ service.name }}</RouterLink
          >
        </h5>
        <div class="text-muted mb-3">{{ service.description }}</div>
        <span
          v-if="compliance.status == data.ComplianceStatus.Compliant"
          style="color: green"
        >
          Compliant
        </span>
        <span
          v-else-if="compliance.status == data.ComplianceStatus.NonComplaint"
          style="color: red"
        >
          Not Compliant
        </span>
        <span v-else style="color: gray"> Pending </span>
        <div>
          Monitoring
          {{ service.requirements?.requirementIds.length ?? 0 }} control(s)
        </div>
      </div>
      <div class="card-footer text-muted" style="font-size: 0.9rem">
        Last update: {{ compliance.latestDate?.toLocaleString() ?? "Pending" }}
      </div>
    </div>
  </div>
</template>