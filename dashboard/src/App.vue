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

* Main view in which all other views are included
<script setup>
import sidebar from "./components/SideBar.vue";
import topbar from "./components/TopBar.vue";
import footbar from "./components/FootBar.vue";
import router from "./router/router";
</script>

<template>
  <div>
    <topbar />
  </div>
  <div style="margin-left: 15rem; margin-right: 15rem" class="mb-4">
    <Suspense>
      <template #fallback> Loading... </template>

      <router-view />
    </Suspense>
  </div>
  <div>
    <footbar />
  </div>
</template>


<style>
@import url("https://fonts.googleapis.com/css2?family=Titillium+Web&display=swap");

#app {
  font-family: Titillium Web, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  /*text-align: center;*/
  color: #2c3e50;
  margin-top: 0px;
}
</style>
