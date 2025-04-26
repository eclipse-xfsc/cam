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

* Creates a piechart which shows the sum of compliant and not compliant services
<script setup lang="ts">
import BarChart from "../components/barChart.ts";
import * as data from "../data/data";
import * as newdata from "../data/newdata";

var chartdata = null;
var compPerDay = null;

function getDates() {
  var labels = [];
  var i = 0;
  for (; i < 31; i++) {
    var target = new Date();
    target.setDate(target.getDate() - i);
    const date =
      target.getDate() +
      "." +
      (target.getMonth() + 1) +
      "." +
      target.getFullYear();
    labels.push(date);
  }
  return labels;
}

compPerDay = await newdata.getCompPerDay();
chartdata = {
  labels: getDates(),
  datasets: [
    {
      label: "Compliant services",
      backgroundColor: "#000071",
      data: compPerDay,
    },
  ],
};
</script>

<template>
  <BarChart v-if="chartdata" :chartData="chartdata" />
</template>