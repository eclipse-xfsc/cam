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

<script setup lang="ts">
import * as data from "../data/data";
import { Compliance } from "../data/model";
import * as newdata from "../data/newdata";
import { ComplianceStatus } from "../data/newdata";
import PieChart from "./PieChart.vue";

export interface ComplianceData {
  numServices: number;
  numPending: number;
  numCompliant: number;
  numNonCompliant: number;
  data: Map<string, Compliance[]>[];
}

const overview = await calculateComplianceOverview();

const chartdata = {
  labels: ["Compliant", "Not Compliant", "Pending"],
  datasets: [
    {
      backgroundColor: ["#41B883", "#E46651", "#c0c0c0"],
      data: [
        overview.numCompliant,
        overview.numNonCompliant,
        overview.numPending,
      ],
    },
  ],
};

/**
 * Calculates compliance overview data for all services.
 */
async function calculateComplianceOverview(): Promise<ComplianceData> {
  // Retrieve all services
  const services = await newdata.listCloudServices();
  let numNonCompliant = 0;
  let numPending = 0;
  let data: Map<string, Compliance[]>[] = [];

  // Loop through all services and retrieve compliance results
  for (const service of services) {
    let groupedResults = await newdata.listGroupedCompliance(service.id);
    let latest = await newdata.retrieveLatestCompliance(groupedResults);

    // Add it to the compliance result
    data.push(groupedResults);

    let status = newdata.isServiceCompliant(latest);
    if (status == ComplianceStatus.Pending) {
      numPending++;
      continue;
    } else if (status == ComplianceStatus.NonComplaint) {
      numNonCompliant++;
      continue;
    }
  }

  return {
    numServices: services.length,
    numCompliant: services.length - numPending - numNonCompliant,
    numNonCompliant: numNonCompliant,
    numPending: numPending,
    data: data,
  };
}
</script>

<template>
  <div v-if="numServices <= 0">
    <p>No services could be loaded</p>
  </div>
  <div v-else>
    <div id="overview">
      <h3>Monitoring {{ overview.numServices }} services</h3>
      <h3>{{ overview.numCompliant }} compliant</h3>
      <h3>{{ overview.numPending }} pending for evaluation</h3>
      <h3>{{ overview.numNonCompliant }} not compliant</h3>
    </div>

    <div id="diagramm" class="mt-4">
      <PieChart v-if="chartdata" :chartData="chartdata" />
    </div>
  </div>
</template>
