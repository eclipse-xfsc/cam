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

// Shows the evaluation results for all services

<script setup lang="ts">
import * as data from "../data/newdata";
import * as olddata from "../data/data";
import * as auth from "../auth/auth";
import { Ref, ref } from "vue";
import { CloudService, Control } from "../data/model";
import ServiceSummary from "../components/ServiceSummary.vue";

let services: Ref<CloudService[]> = ref([]);
let controls: Ref<Control[]> = ref([]);

if (auth.isAuthenticated() || olddata.getTestdata()) {
  services.value = await data.listCloudServices();
  controls.value = await data.listControls();
}
</script>

<template>
  <div v-if="!auth.isAuthenticated() && !olddata.getTestdata()">
    <p>Please log in or use testdata to continue</p>
  </div>
  <div v-else id="evaluation">
    <h2 class="pb-2">Cloud Service Overview</h2>

    <!-- For each service show a service eval -->
    <div v-if="services.length == 0">
      <p>No services</p>
    </div>
    <div v-else>
      <div class="card-rows">
        <p v-for="service in services">
          <ServiceSummary :service="service" :controls="controls" />
        </p>
      </div>
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