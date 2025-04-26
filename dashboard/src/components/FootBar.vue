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

* Generates the footbar. 
<script setup lang="ts">
function getYear() {
  var today = new Date();
  const date = today.getFullYear();
  return date;
}
</script>

<template>
  <div class="top">
    <img src="/GXFS-white.png" style="width: 150px" />
  </div>
  <div class="bottom p-2">
    <p>Gaia-X Federation Services - Continuous Automated Monitoring</p>
  </div>
</template>

<style scoped>
.top {
  text-align: center;
  background-color: rgb(0, 0, 148);
  color: rgb(233, 235, 233);
  width: 100%;
}

.bottom {
  text-align: center;
  background-color: rgb(0, 0, 113);
  color: rgb(233, 235, 233);
  width: 100%;
}

img {
  padding: 25px;
}
</style>