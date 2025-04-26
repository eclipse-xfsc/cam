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

* Shows the evaluation results of the last 30 days
<script setup lang="ts">
import $ from "jquery";
import ServiceSummary from "../components/ServiceSummary.vue";
import HistoryChart from "../components/HistoryChart.vue";
import * as data from "../data/data";
import * as auth from "../auth/auth";
import ControlPercentage from "../components/ControlPercentage.vue";
import * as newdata from "../data/newdata";
import { CloudService, Control } from "../data/model";
import { ref, Ref } from "vue";


var startDate = new Date();
var startDay = 0;
var endDate = new Date();
var endDay = 0;

var services: Ref<CloudService[]> = ref([]);
var controls: Ref<Control[]> = ref([]);

async function load() {
  if (auth.isAuthenticated() || data.getTestdata()) {
    services.value = await newdata.listCloudServices();
    controls.value = await newdata.listControls();
  }
}

load();

function getDay(n) {
  var target = new Date();
  target.setDate(target.getDate() - n);
  const date =
    target.getDate() +
    "." +
    (target.getMonth() + 1) +
    "." +
    target.getFullYear();
  return date;
}

async function startDateChanged(event) {
  startDate.setDate(startDate.getDate() - event.target.value);
  startDay = event.target.value;
  load();
}

async function endDateChanged(event) {
  endDate.setDate(endDate.getDate() - event.target.value);
  endDay = event.target.value;
  load();
}
</script>


<template>
  <div v-if="!auth.isAuthenticated() && !data.getTestdata()">
    <p>Please log in or use testdata to continue</p>
  </div>
  <div v-else id="overview">
    <div v-if="services.length == 0">
      <p>No services could be loaded</p>
    </div>
    <div v-else>
      <!-- History chart -->
      <div>
        <h1>Compliant services per day</h1>
        <HistoryChart />
        <p>
          Note: A service will only be considered as compliant if all its
          controls were compliant for the whole day.
        </p>
      </div>
      <div>
        <!-- Details for a specific date -->
        &nbsp;&nbsp;&nbsp;&nbsp;
        <h1>Evaluation for a time period</h1>
        <p>
          Show evaluation between
          <select
            name="startDate"
            @change="startDateChanged($event)"
            id="startDate"
          >
            <option v-for="n in 31" v-bind:key="n" v-bind:value="n - 1">
              -{{ n - 1 }} days | {{ getDay(n) }}
            </option>
          </select>
          and
          <select name="endDate" @change="endDateChanged($event)" id="endDate">
            <option v-for="n in 31" v-bind:key="n" v-bind:value="n - 1">
              -{{ n - 1 }} days | {{ getDay(n) }}
            </option>
          </select>
        </p>
        <!-- show evaluation result for selected date -->
        <div class="card-rows">
          <p v-for="service in services" :key="service.id">
            <ServiceSummary
              :service="service"
              :controls="controls"
              :startDay="startDay"
              :endDay="endDay"
            />
          </p>
        </div>
        <p>
          Note: Controls will only be considered as compliant if all evaluations
          in the select time period were compliant.
        </p>
      </div>
      &nbsp;&nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp;&nbsp; &nbsp;&nbsp;&nbsp;&nbsp;
    </div>
  </div>
</template>


<style>
.card-columns {
  display: flex;
  justify-content: space-between;
  flex-direction: row;
  flex-wrap: wrap;
}
</style>