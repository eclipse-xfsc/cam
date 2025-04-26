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

* Shows the compliance in percentage for selected controls
<script setup lang="ts">
import { ref } from "vue";
import * as data from "../data/data";
import PieChart from "./PieChart.vue";

interface ControlPercentageProps {
  evals: any[]
}

let selectedControls = [];
let comp = ref(-1);
let chartdata = {
  labels: ["Compliant", "Not Compliant"],
  datasets: [
    {
      backgroundColor: ["#41B883", "#E46651", "#00D8FF", "#DD1B16"],
      data: [0, 0],
    },
  ],
};

const props = defineProps<ControlPercentageProps>();
const controls = await data.getControlsFromEval(props.evals);

function calc() {
  comp.value = data.getPercentagePerControl(
    props.evals,
    selectedControls
  );
  chartdata.datasets[0].data = [comp.value, 100 - comp.value];
}
</script>

<template>
  <p v-for="item in controls">
    <input type="checkbox" :value="item" v-model="selectedControls" />
    <span class="checkbox-label"> {{ item }} </span> <br />
  </p>

  <button type="button" class="btn btn-success" v-on:click="calc">
    Calculate compliance
  </button>

  <div v-if="comp == -1"></div>
  <div v-else>
    <PieChart :chartData="chartdata" />
    &nbsp;&nbsp;&nbsp;&nbsp;
    <h5>{{ comp }}% Compliant</h5>
    <h5>{{ 100 - comp }}% Not Compliant</h5>
  </div>
</template>

<style>
.dot {
  height: 25px;
  width: 25px;
  border-radius: 50%;
  display: inline-block;
}
</style>