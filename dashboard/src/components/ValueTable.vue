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
export interface ValueTableProps {
  value: any;
}
defineProps<ValueTableProps>();
</script>
<template>
  <table>
    <tr class="values" v-for="entry of Object.entries(value)">
      <td class="pe-4">{{ entry[0] }}</td>
      <td>{{ entry[1] }}</td>
    </tr>
  </table>
</template>

<style>
table {
  table-layout: fixed;
  width: 100%;
  border-collapse: collapse;
}
tr.values {
  border: solid thin;
}
table td {
  word-wrap: break-word;
  overflow-wrap: break-word; 
}
</style>