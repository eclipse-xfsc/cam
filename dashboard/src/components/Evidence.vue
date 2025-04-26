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
import { CloudService, Metric } from "../data/model";
import * as data from "../data/newdata";
import ValueTable from "./ValueTable.vue";
import RawEvidence from "../components/RawEvidence.vue";

export interface EvaluationProps {
  evidenceId: string;
  service: CloudService;
  metric: Metric;
}

let props = defineProps<EvaluationProps>();

let evidence = await data.getEvidence(props.evidenceId);

function flatten(obj, parent?, result = {}) {
  for (let key in obj) {
    let property = parent ? parent + "." + key : key;

    if (typeof obj[key] == "object") {
      flatten(obj[key], property, result);
    } else {
      result[property] = obj[key];
    }
  }

  return result;
}

function id(pre: string, local: string): string {
  return `${pre}-${props.service.id}-${local}`;
}
</script>

<template>
  <div
    class="modal fade text-black"
    tabindex="-1"
    aria-labelledby="exampleModalLabel"
    aria-hidden="true"
  >
    <div class="modal-dialog modal-xl">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" id="exampleModalLabel">Evidence</h5>
          <button
            type="button"
            class="btn-close"
            data-bs-dismiss="modal"
            aria-label="Close"
          ></button>
        </div>
        <div class="modal-body">
          <table class="table">
            <tbody>
              <tr>
                <th>Evidence Id</th>
                <td>{{ evidence.id }}</td>
              </tr>
              <tr>
                <th style="min-width: 150px">Target Service</th>
                <td>{{ evidence.targetService }} ({{ service.name }})</td>
              </tr>
              <tr>
                <th>Target Resource</th>
                <td>{{ evidence.targetResource }}</td>
              </tr>
              <tr>
                <th>Tool ID</th>
                <td>{{ evidence.toolId }}</td>
              </tr>
              <tr>
                <th>Gathered At</th>
                <td>{{ evidence.gatheredAt }}</td>
              </tr>
              <tr>
                <th>Value</th>
                <td>
                  <ValueTable :value="flatten(evidence.value, null)" />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="modal-footer">
          <button
            type="button"
            class="btn btn-primary"
            data-bs-toggle="modal"
            :data-bs-target="'#' + id(props.metric.id, 'rawevidence')"
            :disabled="!evidence.rawEvidence"
          >
            Raw Evidence
          </button>
          <RawEvidence
            :id="id(props.metric.id, 'rawevidence')"
            :evidence-id="evaluation.evidenceId"
            :service="service"
            v-if="evaluation"
          >
          </RawEvidence>
          <!-- TODO(oxisto): prefer lazy loading of evidences -->
          <button
            type="button"
            class="btn btn-secondary"
            data-bs-dismiss="modal"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.modal-body {
  text-align: left;
}
</style>