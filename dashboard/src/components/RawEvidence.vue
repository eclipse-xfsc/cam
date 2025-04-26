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
import { CloudService } from "../data/model";
import * as data from "../data/newdata";
import ValueTable from "./ValueTable.vue";

interface RawEvaluationProps {
  evidenceId: string;
  service: CloudService;
}

let props = defineProps<RawEvaluationProps>();

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
          <h5 class="modal-title" id="ModalLabel">Raw Evidence</h5>
          <button
            type="button"
            class="btn-close"
            data-bs-dismiss="modal"
            aria-label="Close"
          ></button>
        </div>
        <div class="modal-body" v-if="evidence.rawEvidence">
            <ValueTable :value="flatten(JSON.parse(evidence.rawEvidence), null)" />
        </div>
        <div class="modal-footer">
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