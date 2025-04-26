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
import { serve } from 'esbuild';
import { ref } from 'vue';
import { Metric } from '../data/model';
import * as data from "../data/newdata";

export interface MetricConfigurationProps {
    serviceId: string
    metricId: string
}

const props = defineProps<MetricConfigurationProps>();

const config = ref(await data.getMetricConfiguration(props.serviceId, props.metricId));

async function save() {
    // do some magic with target value
    if (config.value.targetValue.includes(",")) {
        // split it 
        config.value.targetValue = config.value.targetValue.split(",");
    } else if (config.value.targetValue == "true") {
        config.value.targetValue = true;
    } else if (config.value.targetValue == "false") {
        config.value.targetValue = true;
    }

    // make sure we set the metric ID and service ID within the config itself also
    config.value.serviceId = props.serviceId;
    config.value.metricId = props.metricId;

    config.value = await data.updateMetricConfiguration(props.serviceId, props.metricId, config.value);
}
</script>
<template>
    <div class="modal fade text-black" tabindex="-1" aria-labelledby="metricConfigLabel" aria-hidden="true">
        <div class="modal-dialog modal-xl">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="metricConfigLabel">Metric Configuration ({{ metricId }})</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div class="input-group mb-3">
                        <span class="input-group-text" id="operator">Operator</span>
                        <input type="text" class="form-control" placeholder="==" aria-label="operator"
                            aria-describedby="operator" v-model="config.operator">
                    </div>
                    <div class="input-group mb-3">
                        <span class="input-group-text" id="targetValue">Target Value</span>
                        <input type="text" class="form-control" placeholder="" aria-label="targetValue"
                            aria-describedby="targetValue" v-model="config.targetValue">
                    </div>
                    <div class="input-group mb-3">
                        <span class="input-group-text" id="targetValue">Is Default?</span>
                        <input type="text" class="form-control" placeholder="" aria-label="targetValue"
                            aria-describedby="targetValue" :value="config.isDefault" disabled>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
                    <button type="button" class="btn btn-primary" v-on:click="save">Save changes</button>
                </div>
            </div>
        </div>
    </div>
</template>