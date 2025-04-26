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

* Generates the topbar. It displays the name of the current view
<script setup lang="ts">
import * as auth from "../auth/auth";
import { useRouter, useRoute } from 'vue-router'
import { RouterLink } from 'vue-router'
import { inject } from "vue";

const router = useRouter()
const route = useRoute()

function currentRouteName() {
  return route.name;
}

const client = inject("client");
</script>

<template>
  <div>

    <h4 class="p-2">
      <router-link to="/evaluation"
        :class="{ 'active': currentRouteName() === 'Evaluation' || currentRouteName() === 'EvaluationDetail' }">Overview
      </router-link>
      <router-link to="/history" :class="{ 'active': currentRouteName() === 'History' }">History</router-link>
      <router-link to="/configuration"
        :class="{ 'active': currentRouteName() === 'Configuration' || currentRouteName() === 'ConfigurationDetail' }">
        Configuration
      </router-link>
      <!--<router-link to="/settings" :class="{ 'active': currentRouteName() === 'Settings' }">Settings</router-link>-->
      <button v-if="auth.isAuthenticated()" type="button" class="btn btn-danger" v-on:click="auth.logout(client)"
        style="float: right;">Logout </button>
      <button v-else type="button" class="btn btn-success" v-on:click="auth.login(client)"
        style="float: right;">Login</button>
    </h4>
    <hr />
  </div>
</template>

<style scoped>
div {
  text-align: center;
  background-color: rgb(255, 255, 255);
  color: rgb(233, 235, 233);
}

hr {
  color: rgb(0, 0, 113);
  height: 5px;
}

h4 {
  color: rgb(0, 0, 0);
}

h4 a {
  color: rgb(0, 0, 0);
  text-decoration: none;
  font-weight: bold;
  font-size: 1.2rem;
  margin-left: 1rem;
  margin-right: 1rem;
}

.active {
  border-bottom: 4px solid #000094;
}
</style>