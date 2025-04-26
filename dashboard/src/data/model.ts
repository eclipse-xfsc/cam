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

export interface CloudService {
  id: string
  name: string
  description?: string
  requirements?: {
    requirementIds: string[]
  }
}

export interface Control {
  id: string
  name: string
  metrics: Metric[]
}

export interface Metric {
  id: string
}

export interface MetricConfiguration {
  isDefault: boolean
  operator: string
  targetValue: any
  updatedAt?: string
  metricId: string
  serviceId: string
}

export interface ConfigureServiceRequest {
  configurations: ServiceConfiguration[]
}

export interface BaseConfig {
  "@type": "type.googleapis.com/cam.CommunicationSecurityConfig" |
  "type.googleapis.com/cam.WorkloadSecurityConfig" |
  "type.googleapis.com/cam.RemoteIntegrityConfig" |
  "type.googleapis.com/cam.AuthenticationSecurityConfig"
}


export interface AuthenticationSecurityConfig extends BaseConfig {
  "@type": "type.googleapis.com/cam.AuthenticationSecurityConfig"
  issuer: String
  metadataDocument: String
  apiEndpoint: String
  clientId: String
  clientSecret: String
  scopes: String
}

export interface RemoteIntegrityConfig extends BaseConfig {
  "@type": "type.googleapis.com/cam.RemoteIntegrityConfig"
  certificate: string | ArrayBuffer | null | undefined;
  target: string
}

export interface CommunicationSecurityConfig extends BaseConfig {
  //"@type": "type.googleapis.com/cam.CommunicationSecurityConfig"
  endpoint: string
}

export interface WorkloadSecurityConfig extends BaseConfig {
  //"@type": "type.googleapis.com/cam.WorkloadSecurityConfig"
  kubernetes: String | ArrayBuffer | null | undefined,
  openstack: OpenStackConfiguration,
  aws: AWSConfiguration,
}

export interface OpenStackConfiguration extends BaseConfig {
  endpoint: string,
  tenant: string,
  region: string,
  username: string,
  password: string,
}

export interface AWSConfiguration extends BaseConfig {
  region: string,
  accessKeyId: string
  secretAccessKey: string
}

export interface ServiceConfiguration {
  serviceId: string
  rawConfiguration: CommunicationSecurityConfig |
  AuthenticationSecurityConfig |
  RemoteIntegrityConfig |
  WorkloadSecurityConfig
}

export interface GetMonitoringStatusResponse {
  serviceId: string
  controlIds: string[]
}

export interface StartMonitoringRequest {
  serviceId: string
  controlIds: string[]
}

export interface StartMonitoringResponse {
  status: MonitoringStatus
}

export interface ListServiceConfigurationsResponse {
  configurations: ServiceConfiguration[];
}

export interface RegisterCloudServiceRequest {
  name: string
  description?: string
}

export interface ListCloudServicesResponse {
  services: CloudService[];
}

export interface MonitoringStatus {
  serviceId: string
  controlIds: string[]
  lastRun?: string | Date
  nextRun?: string | Date
}

export interface ListComplianceResponse {
  complianceResults: Compliance[];
}

export interface EvaluationResult {
  id: string
  evidenceId: string
  metricId: string
  serviceId: string
  status: boolean
  time: string
}

export interface Compliance {
  id: string
  serviceId: string
  controlId: string
  evaluations: EvaluationResult[]
  status: boolean
  time: string
}

export interface Evidence {
  id: string
  value: object
  targetService: string
  targetResource: string
  toolId: string
  gatheredAt: string
}