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

* In this view the user can configure the controls to be monitored for each service

<script setup lang="ts">
import { Ref, ref } from "vue";
import ServiceConfig from "../components/ServiceConfig.vue";
import { CloudService } from "../data/model";
import * as data from "../data/newdata";
import * as olddata from "../data/data";
import * as auth from "../auth/auth";

var services: Ref<CloudService[]> = ref([]);
if (auth.isAuthenticated() || olddata.getTestdata()) {
  services.value = await data.listCloudServices();
}

async function add() {
  let name = prompt("Service name:", "Service");

  if (name == null) {
    return add();
  }

  let description =
    prompt("Service description:", "My awesome service") ?? undefined;

  const service = await data.registerCloudService(name, description);
  services.value.push(service);
}

async function del(serviceId: string) {
  await data.removeCloudService(serviceId);

  let service = services.value.find((service) => service.id == serviceId);

  if (service != null) {
    let idx = services.value.indexOf(service);
    services.value.splice(idx, 1);
  }
}

async function edit(service: CloudService) {
  let name = prompt("New service name:", service.name);

  if (name == null) {
    return edit(service);
  }

  let description =
    prompt("Service description:", service.description) ?? undefined;

  service.name = name;
  service.description = description;

  service = await data.updateCloudService(service.id, service);
}
</script>

<template>
  <div v-if="!auth.isAuthenticated() && !olddata.getTestdata()">
    <p>Please log in or use testdata to continue</p>
  </div>
  <div v-else id="overview">
    <h2>Monitoring Configuration</h2>
    <p></p>
    <div class="card-columns">
      <p v-for="service in services">
        <ServiceConfig :service="service" @delete="del(service.id)" @edit="edit(service)" />
      </p>
      <div class="card" style="width: 20rem">
        <div class="card-body p-2">
          <button type="button" class="btn btn-primary mb-2" v-on:click="add">
            Add new service
          </button>
        </div>
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