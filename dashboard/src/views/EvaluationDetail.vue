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

<script setup lang="ts">
import * as data from "../data/newdata";
import * as olddata from "../data/data";
import * as auth from "../auth/auth";
import { Ref, ref } from "vue";
import ControlEval from "../components/ControlEval.vue";
import PieChart from "../components/PieChart.vue";
import { CloudService, Control } from "../data/model";
import { useRouter, useRoute } from "vue-router";

export interface EvaluationDetailProps {
  id: string;
}

const props = defineProps<EvaluationDetailProps>();

let service: Ref<CloudService> = ref({} as CloudService);
let controls: Ref<Control[]> = ref([]);

const router = useRouter();
const route = useRoute();
let url = route.fullPath;
let startDay = route.query.startDay;
let endDay = route.query.endDay;

if (auth.isAuthenticated() || olddata.getTestdata()) {
  service.value = await data.getCloudService(props.id);
  controls.value = await data.listControls();
}

let compliance = null;
if (startDay) {
  compliance = await data.calculateCompliance(
    service.value,
    Number(startDay),
    Number(endDay)
  );
} else {
  compliance = await data.calculateCompliance(service.value);
}

function controlFor(controlId: string): Control {
  return (
    controls.value.find((control) => control.id == controlId) ?? {
      id: controlId,
      name: "Unknown",
      metrics: [],
    }
  );
}
</script>

<template>
  <div v-if="!auth.isAuthenticated() && !olddata.getTestdata()">
    <p>Please log in or use testdata to continue</p>
  </div>
  <div v-else id="evaluation">
    <div v-if="!compliance">
      <p>No data for this service at selected time.</p>
    </div>
    <div v-else id="data">
      <PieChart
        v-if="compliance.chartdata"
        :chartData="compliance.chartdata"
        :height="165"
        style="max-width: 400px; float: right"
      />

      <h2 class="pb-2">{{ service.name }}</h2>
      <div class="text-muted mb-4">{{ service.description }}</div>
      <span
        v-if="compliance.status == data.ComplianceStatus.Compliant"
        class="text-success"
      >
        Compliant
      </span>
      <span
        v-else-if="compliance.status == data.ComplianceStatus.NonComplaint"
        class="text-danger"
      >
        Not Compliant
      </span>
      <span v-else style="color: gray"> Pending </span>
      <div class="mb-4">
        Monitoring
        {{ service.requirements?.requirementIds.length ?? 0 }} control(s)
      </div>
      <div class="card-footer text-muted mb-2" style="font-size: 0.9rem">
        Last update: {{ compliance.latestDate?.toLocaleString() ?? "Pending" }}
      </div>
      <div class="card-columns">
        <p
          v-for="controldId in service.requirements?.requirementIds ?? []"
          class="mt-2"
        >
          <ControlEval
            :service="service"
            :control="controlFor(controldId)"
            :result="compliance.latest.get(controldId)"
          />
        </p>
      </div>
    </div>
  </div>
</template>
