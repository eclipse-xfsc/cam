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

* Lists for a given service all controls
* With checkboxes the controls to apply to this service can be selected
* Used for the configuration of services

<script setup lang="ts">
import { ref } from "vue";
import * as olddata from "../data/data";
import * as data from "../data/newdata";
import type { CloudService } from "../data/model";

interface ServiceConfigProps {
  service: CloudService
}

// Define properties
const props = defineProps<ServiceConfigProps>()

// Define emitted events
defineEmits(['delete', 'edit'])

// Fetch control and configuration data
const controls = await olddata.getAllControls();

// Fetch current monitoring status
let status = ref(await data.getMonitoringStatus(props.service.id));
let isRunning = ref(status.value.controlIds.length > 0);
let selectedControls = ref(status.value.controlIds);

async function startMonitoring() {
  // Update requirements of Cloud Service to keep it in sync
  props.service.requirements = {
    requirementIds: selectedControls.value
  };
  await data.updateCloudService(props.service.id, props.service);

  // Start the monitoring and update the monitoring status accordingly
  status.value = await data.startMonitoring(props.service.id, selectedControls.value);
  isRunning.value = true;
}

async function stopMonitoring() {
  await olddata.setMonitoring(props.service.id, false, selectedControls.value);
  status.value.nextRun = undefined;
  status.value.lastRun = undefined;
  isRunning.value = false;
}
</script>

<template>
  <form>
    <div class="card" style="width: 20rem">
      <div class="card-body p-2">
        <h5 class="card-title">{{ service.name }}
          <font-awesome-icon v-on:click="$emit('edit')" icon="fa-solid fa-pen-to-square" />
        </h5>
        <div class="text-muted">{{ service.description }}</div>
        <h6 class="card-subtitle mt-2 mb-2 text-muted">
          Select controls to monitor
        </h6>

        <!-- Select controls: -->
        <ul class="list-group">
          <li v-for="item in controls" class="list-group-item d-flex justify-content-between align-items-center"
            :class="{ 'active': selectedControls.includes(item.id) }">
            <label :for="item.id + props.service.id" class="pe-2">{{ item.name }}</label>
            <input type="checkbox" :id="item.id + props.service.id" :value="item.id" v-model="selectedControls"
              class="me-1" :disabled="isRunning" />
          </li>
        </ul>

        <template v-if="isRunning == true">
          <button type="button" class="btn btn-danger mt-2" v-on:click="stopMonitoring">
            Stop Monitoring
          </button>
        </template>
        <template v-else>
          <button type="button" class="btn btn-success mt-2" v-on:click="startMonitoring"
            :disabled="selectedControls.length == 0">
            Start Monitoring
          </button>
        </template>

        <RouterLink :to="'/configuration/' + service.id">
          <button type="button" class="btn btn-primary mt-2 ms-2">
            <font-awesome-icon icon="fa-solid fa-gear" />
          </button>
        </RouterLink>
        <button type="button" class="btn btn-danger mt-2 ms-2" v-on:click="$emit('delete')">
          <font-awesome-icon icon="fa-solid fa-trash" />
        </button>

        <div v-if="status.lastRun" class="text-muted mt-2" style="font-size: 0.75rem">Last run: {{ new
            Date(status.lastRun)
        }}</div>

        <div v-if="status.nextRun" class="text-muted mt-2" style="font-size: 0.75rem">Next run: {{ new
            Date(status.nextRun)
        }}</div>
      </div>
    </div>
  </form>
</template>

<style scoped>
.card {
  float: none;
}

.form-group {
  margin-top: 0.5rem;
}
</style>