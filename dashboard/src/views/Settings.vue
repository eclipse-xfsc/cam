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

* Here general Settings can be set
<script setup lang="ts">
import * as data from "../data/data";

function useTest(test) {
  data.setTestdata(test);
  location.reload();
}

function runLocal(port) {
  localStorage.setItem("runLocal", "true");
  localStorage.setItem("localSiteURL", 'http://localhost:' + port);
  location.reload();
}

function runOnline() {
  localStorage.setItem("runLocal", "false");
  location.reload();
}
</script>

<template>
  <h1>This might be deleted for production env.</h1>
  <!--To remove for production:-->

  <h3 class="mt-4">Use local testdata instead of API (test: {{ data.getTestdata() }})</h3>
  <button type="button" class="btn btn-success me-2" v-on:click="useTest(false)">
    Use API
  </button>
  <button type="button" class="btn btn-success me-2" v-on:click="useTest(true)">
    Use test
  </button>
  <h3 class="mt-4">Is the app running local or online?</h3>
  <button type="button" class="btn btn-success me-2" v-on:click="runLocal(3000)">
    local:3000
  </button>
  <button type="button" class="btn btn-success me-2" v-on:click="runLocal(8080)">
    local:8080
  </button>
  <button type="button" class="btn btn-success me-2" v-on:click="runOnline()">
    online
  </button>
</template>

<style>
.card-columns {
  display: flex;
  justify-content: space-between;
  flex-direction: row;
  flex-wrap: wrap;
}
</style>