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
import { CloudService, EvaluationResult, Metric } from "../data/model";
import MetricConfiguration from "../components/MetricConfiguration.vue";
import Evidence from "../components/Evidence.vue";
import RawEvidence from "../components/RawEvidence.vue";

export interface MetricEvalProps {
  evaluation?: EvaluationResult;
  metric: Metric;
  service: CloudService;
}
const props = defineProps<MetricEvalProps>();

function evaluationClass(result?: EvaluationResult) {
  return {
    "bg-secondary": result == undefined,
    "bg-danger": result && !result.status,
    "bg-success": result && result.status,
  };
}

function id(pre: string, local: string): string {
  return `${pre}-${props.service.id}-${local}`;
}
</script>

<template>
  <li
    class="
      list-group-item
      d-flex
      justify-content-between
      align-items-center
      text-white
    "
    :class="evaluationClass(evaluation)"
  >
    {{ metric.id }}
    <div>
      <button
        type="button"
        class="btn btn-primary me-2"
        data-bs-toggle="modal"
        :data-bs-target="'#' + id(metric.id, 'metricConfig')"
      >
        <font-awesome-icon icon="fa-solid fa-gear" />
      </button>

      <button
        type="button"
        class="btn btn-primary me-2"
        data-bs-toggle="modal"
        :data-bs-target="'#' + id(metric.id, 'evidence')"
        :disabled="!evaluation"
      >
        <font-awesome-icon icon=" fa-solid fa-book" />
      </button>

      <!-- TODO(oxisto): prefer lazy loading of evidences -->
      <Evidence
        :id="id(metric.id, 'evidence')"
        :evidence-id="evaluation.evidenceId"
        :service="service"
        :metric="metric"
        v-if="evaluation"
      >
      </Evidence>
      <RawEvidence
        :id="id(metric.id, 'rawevidence')"
        :evidence-id="evaluation.evidenceId"
        :service="service"
        v-if="evaluation"
      >
      </RawEvidence>
      <MetricConfiguration
        :id="id(metric.id, 'metricConfig')"
        :metric-id="metric.id"
        :service-id="service.id"
      />
    </div>
  </li>
</template>