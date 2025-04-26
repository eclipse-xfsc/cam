// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Contributors:
//	Fraunhofer AISEC

import {
    CloudService,
    Compliance,
    ConfigureServiceRequest,
    Evidence,
    ListCloudServicesResponse,
    ListComplianceResponse,
    ListServiceConfigurationsResponse,
    MetricConfiguration,
    MonitoringStatus,
    RegisterCloudServiceRequest,
    ServiceConfiguration,
    StartMonitoringRequest,
    StartMonitoringResponse
} from "./model";
import * as REST from "./REST";

export async function listControls() {
    return REST.getDataFromApi("configuration/controls").then(res => res.requirements);
}

export async function listCompliance(serviceId: string): Promise<ListComplianceResponse> {
    return REST.getDataFromApi(`evaluation/cloud_services/${serviceId}/compliance?order_by=time&asc=false&page_size=100`);
}

export async function listCompliance30Days(serviceId: string): Promise<ListComplianceResponse> {
    return REST.getDataFromApi(`evaluation/cloud_services/${serviceId}/compliance?order_by=time&asc=false&days=30`);
}

export async function getCompliance(controlId: string, serviceId: string): Promise<Compliance> {
    return REST.getDataFromApi(`evaluation/cloud_services/${serviceId}/compliance/${controlId}`);
}

export async function listCloudServices(): Promise<ListCloudServicesResponse> {
    return REST.getDataFromApi("configuration/cloud_services");
}

export async function getCloudService(serviceId: string): Promise<CloudService> {
    return REST.getDataFromApi(`configuration/cloud_services/${serviceId}`)
}

export async function listServiceConfigurations(serviceId: string): Promise<ListServiceConfigurationsResponse> {
    return REST.getDataFromApi(`configuration/cloud_services/${serviceId}/configurations`);
}

export async function registerCloudService(request: RegisterCloudServiceRequest): Promise<CloudService> {
    return REST.setDataToApi(`configuration/cloud_services`, request);
}

export async function removeCloudService(serviceId: string): Promise<void> {
    return REST.deleteDataToApi(`configuration/cloud_services/${serviceId}`);
}

export async function updateCloudService(serviceId: string, service: CloudService): Promise<CloudService> {
    return REST.putDataToApi(`configuration/cloud_services/${serviceId}`, service);
}

export function configureCloudService(serviceId: string, request: ConfigureServiceRequest): Promise<any> {
    return REST.putDataToApi(`configuration/cloud_services/${serviceId}/configurations`, request);
}

export async function startMonitoring(serviceId: string, request: StartMonitoringRequest): Promise<StartMonitoringResponse> {
    return REST.setDataToApi(`configuration/monitoring/${serviceId}/start`, request)
}

export async function getMonitoringStatus(serviceId: string): Promise<MonitoringStatus> {
    let resp = await REST.getDataFromApi(`configuration/monitoring/${serviceId}`)

    // Check for "not found". This means that the monitoring is stopped
    // TODO(lebogg): We will fix this in the backend
    if (resp.code == 5) {
        return {
            serviceId: serviceId,
            controlIds: []
        };
    }

    return resp
}

export async function getEvidence(evidenceId: string): Promise<Evidence> {
    return REST.getDataFromApi(`evaluation/evidences/${evidenceId}`);
}

export async function getMetricConfiguration(serviceId: string, metricId: string): Promise<MetricConfiguration> {
    return REST.getDataFromApi(`configuration/cloud_services/${serviceId}/metric_configurations/${metricId}`);
}

export async function updateMetricConfiguration(serviceId: string, metricId: string, config: MetricConfiguration): Promise<MetricConfiguration> {
    return REST.putDataToApi(`configuration/cloud_services/${serviceId}/metric_configurations/${metricId}`, config);
}