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
import { ref } from "vue";
import { useRouter } from "vue-router";
import { AuthenticationSecurityConfig, CommunicationSecurityConfig, RemoteIntegrityConfig, ServiceConfiguration, WorkloadSecurityConfig } from "../data/model";
import * as data from "../data/newdata";

export interface ConfigurationDetailProps {
    id: string
}

const props = defineProps<ConfigurationDetailProps>();

const service = ref(await data.getCloudService(props.id));
const configurations = await data.listServiceConfigurations(service.value.id);

let kubernetes = ref("kubernetes");

// Split configurations into individual variables for easier template access
let workloadConfig = ref(configFor<WorkloadSecurityConfig>(configurations, "type.googleapis.com/cam.WorkloadSecurityConfig"))
let commsecConfig = ref(configFor<CommunicationSecurityConfig>(configurations, "type.googleapis.com/cam.CommunicationSecurityConfig"))
let integrityConfig = ref(configFor<RemoteIntegrityConfig>(configurations, "type.googleapis.com/cam.RemoteIntegrityConfig"))
let authConfig = ref(configFor<AuthenticationSecurityConfig>(configurations, "type.googleapis.com/cam.AuthenticationSecurityConfig"))

const router = useRouter();

async function save() {
    await data.configureCloudService(service.value.id, [
        {
            serviceId: service.value.id,
            rawConfiguration: workloadConfig.value,
        },
        {
            serviceId: service.value.id,
            rawConfiguration: commsecConfig.value,
        },
        {
            serviceId: service.value.id,
            rawConfiguration: integrityConfig.value
        },
        {
            serviceId: service.value.id,
            rawConfiguration: authConfig.value
        },
    ]);
    router.back();
}

function kubernetesFileChanged(event) {
    var file = event.target.files[0];
    const reader = new FileReader();
    if (file.name.includes(".yaml")) {
        reader.onload = (res) => {
            workloadConfig.value.kubernetes = res.target?.result;
        };
        reader.onerror = (err) => console.log(err);
        reader.readAsText(file);
    } else {
        alert("only .txt files will be processed");
    }
}

function showKubernetesFile() {
    alert(workloadConfig.value.kubernetes);
}

function certChanged(event) {
    let file = event.target.files[0];
    const reader = new FileReader();
    if (file.name.includes(".pem") || file.name.includes(".cert")) {
        reader.onload = (res) => {
            integrityConfig.value.certificate = res.target?.result;
        };
        reader.onerror = (err) => console.log(err);
        reader.readAsText(file);
    } else {
        alert("only .cert files will be processed");
    }
}

function showCert() {
    alert(integrityConfig.value.certificate);
}

function configFor<T extends
    AuthenticationSecurityConfig |
    RemoteIntegrityConfig |
    CommunicationSecurityConfig |
    WorkloadSecurityConfig
>(configurations: ServiceConfiguration[],
    module: "type.googleapis.com/cam.AuthenticationSecurityConfig" |
        "type.googleapis.com/cam.CommunicationSecurityConfig" |
        "type.googleapis.com/cam.RemoteIntegrityConfig" |
        "type.googleapis.com/cam.WorkloadSecurityConfig"
): T {
    for (let config of configurations) {
        if (config.rawConfiguration["@type"] == module) {
            return config.rawConfiguration as T;
        }
    }

    if (module == "type.googleapis.com/cam.WorkloadSecurityConfig") {
        return { openstack: {}, kubernetes: {}, aws: {}, "@type": module } as T;
    } else {
        return { "@type": module } as T;
    }
}

function id(local: string): string {
    return `${service.value.id}-${local}`
}
</script>

<template>
    <h2 class="pb-2">{{ service.name }}</h2>
    <div class="text-muted mb-4">{{ service.description }}</div>
    <div class="card-columns">
        <div class="card config-card mt-2">
            <div class="card-body">
                <h5>Workload</h5>

                <input type="radio" id="kubernetes" value="kubernetes" v-model="kubernetes" class="me-1" />
                <label for="kubernetes" class="me-2">Kubernetes</label>

                <input type="radio" id="openstack" value="openstack" v-model="kubernetes" class="me-1" />
                <label for="openstack" class="me-2">OpenStack</label>

                <input type="radio" id="aws" value="aws" v-model="kubernetes" class="me-1" />
                <label for="aws" class="me-2">AWS</label>

                <!-- kubernetes: -->
                <div v-if="kubernetes == 'kubernetes'">
                    <h6 class="card-title mt-2">Kubernetes config</h6>
                    <p>Config file</p>
                    <input type="file" ref="kubernetesFile" @change="kubernetesFileChanged" />
                    <button type="button" class="btn btn-success" v-on:click="showKubernetesFile"
                        style="margin-top:10px;">
                        Show current file
                    </button>
                </div>
                <!-- openstack: -->
                <div v-else-if="kubernetes == 'openstack'">
                    <h6 class="card-title mt-2">OpenStack config</h6>

                    <div class="form-group">
                        <label :for="id('openstack-endpoint')">Identity Endpoint</label>
                        <input :id="id('openstack-endpoint')" v-model="workloadConfig.openstack.endpoint"
                            class="form-control" />
                    </div>
                    <div class="form-group">
                        <label :for="id('openstack-tenant')">Tenant Name</label>
                        <input :id="id('openstack-tenant')" v-model="workloadConfig.openstack.tenant"
                            class="form-control" />
                    </div>
                    <div class="form-group">
                        <label :for="id('openstack-region')">Region</label>
                        <input :id="id('openstack-region')" v-model="workloadConfig.openstack.region"
                            class="form-control" />
                    </div>
                    <div class="form-group">
                        <label :for="id('openstack-username')">Username</label>
                        <input :id="id('openstack-username')" v-model="workloadConfig.openstack.username"
                            class="form-control" />
                    </div>
                    <div class="form-group">
                        <label :for="id('openstack-password')">Password</label>
                        <input :id="id('openstack-password')" v-model="workloadConfig.openstack.password"
                            type="password" class="form-control" />
                    </div>
                </div>
                <!-- AWS: -->
                <div v-else>
                    <h6 class="card-title mt-2">AWS config</h6>

                    <div class="form-group">
                        <label :for="id('aws-region')">Region</label>
                        <input :id="id('aws-region')" v-model="workloadConfig.aws.region" class="form-control" />
                    </div>
                    <div class="form-group">
                        <label :for="id('aws-access-key-id')">Access Key ID</label>
                        <input :id="id('aws-access-key-id')" v-model="workloadConfig.aws.accessKeyId"
                            class="form-control" />
                    </div>
                    <div class="form-group">
                        <label :for="id('aws-secret-access-key')">Secret Access Key</label>
                        <input :id="id('aws-secret-access-key')" v-model="workloadConfig.aws.secretAccessKey"
                            type="password" class="form-control" />
                    </div>
                </div>
            </div>
        </div>

        <!-- remote integrity: -->
        <div class="card config-card mt-2">
            <div class="card-body">
                <h5>Remote integrity</h5>
                <p>Target</p>
                <p><input v-model="integrityConfig.target" /></p>
                <p>Certificate</p>
                <input type="file" ref="certFile" @change="certChanged" />
                <button type="button" class="btn btn-success" v-on:click="showCert" style="margin-top:10px;">
                    Show certificate
                </button>
            </div>
        </div>

        <!-- Authentication Security: -->
        <div class="card config-card mt-2">
            <div class="card-body">
                <h5>Authentication Security</h5>
                <div class="form-group">
                    <label :for="id('auth-issuer')" class="mb-2">Issuer</label>
                    <input :id="id('auth-issuer')" v-model="authConfig.issuer" class="form-control" />
                </div>
                <div class="form-group">
                    <label :for="id('auth-metadata')" class="mb-2">Metadata</label>
                    <input :id="id('auth-metadata')" v-model="authConfig.metadataDocument" class="form-control" />
                </div>
                <div class="form-group">
                    <label :for="id('auth-api-endpoint')" class="mb-2">API Endpoint</label>
                    <input :id="id('auth-api-endpoint')" v-model="authConfig.apiEndpoint" class="form-control" />
                </div>
                <div class="form-group">
                    <label :for="id('auth-api-client-id')" class="mb-2">API OAuth 2.0 Client ID</label>
                    <input :id="id('auth-api-client-id')" v-model="authConfig.clientId" class="form-control" />
                </div>
                <div class="form-group">
                    <label :for="id('auth-api-client-secret')" class="mb-2">API OAuth 2.0 Client Secret</label>
                    <input type="password" :id="id('auth-api-client-secret')" v-model="authConfig.clientSecret"
                        class="form-control" />
                </div>
                <div class="form-group">
                    <label :for="id('auth-api-client-scopes')" class="mb-2">API OAuth 2.0 Scopes</label>
                    <input :id="id('auth-api-client-scopes')" v-model="authConfig.scopes" class="form-control" />
                </div>
            </div>
        </div>

        <!-- Communication Security: -->
        <div class="card config-card mt-2">
            <div class="card-body">
                <h5>Communication Security</h5>
                <div class="form-group">
                    <label :for="id('commsec-endpoint')" class="mb-2">Endpoint</label>
                    <input :id="id('commsec-endpoint')" v-model="commsecConfig.endpoint" class="form-control" />
                </div>
            </div>
        </div>
    </div>
    <!-- save service: -->
    <button type="button" class="btn btn-success mt-2" v-on:click="save">
        Save configuration
    </button>
</template>

<style>
.config-card {
    width: 24rem;
    margin-bottom: 1rem;
}
</style>