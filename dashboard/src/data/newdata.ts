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

import * as data from "./data";
import { CloudService, Compliance, Control, Evidence, ListCloudServicesResponse, MetricConfiguration, MonitoringStatus, ServiceConfiguration, StartMonitoringResponse } from "./model";
import * as REST from "./newREST";

export enum ComplianceStatus {
    Compliant = 1,
    NonComplaint,
    Pending
}

/**
 * List cloud service configurations.
 * 
 * @param serviceId the cloud service id
 * @returns the list of service configurations
 */
export async function listServiceConfigurations(serviceId: string): Promise<ServiceConfiguration[]> {
    if (data.getTestdata()) {
        return Promise.resolve([])
    } else {
        return REST.listServiceConfigurations(serviceId).then(res => res.configurations);
    }
}

/**
 * Configures a cloud service.
 * 
 * @param serviceId the cloud service id
 * @param configurations the configurations
 */
export async function configureCloudService(serviceId: string, configurations: ServiceConfiguration[]) {
    if (data.getTestdata()) {
        return Promise.resolve([])
    } else {
        return REST.configureCloudService(serviceId, {
            configurations: configurations
        })
    }
}

/**
 * Lists all cloud services that are registered in the system.
 * 
 * @returns a list of cloud services
 */
export async function listCloudServices(): Promise<CloudService[]> {
    if (data.getTestdata()) {
        return Promise.resolve([
            { name: 'Service-1', id: '1', requirements: { requirementIds: [] } },
            { name: 'Service-2', id: '2', requirements: { requirementIds: [] } },
            { name: 'Service-3', id: '3', requirements: { requirementIds: [] } },
            { name: 'Service-4', id: '4', requirements: { requirementIds: [] } }]);
    } else {
        return REST.listCloudServices().then(res => res.services);
    }
}

/**
 * Returns a specific cloud services.
 * 
 * @param serviceId the cloud service id
 * @returns the cloud service
 */
export async function getCloudService(serviceId: string): Promise<CloudService> {
    if (data.getTestdata()) {
        return Promise.resolve(
            { name: 'Service-1', id: '1', requirements: { requirementIds: [] } },
        );
    } else {
        return REST.getCloudService(serviceId);
    }
}

/**
 * Registeres a new cloud service.
 * 
 * @param name the name of the new cloud service
 * @returns the created cloud service
 */
export async function registerCloudService(name: string, description?: string | undefined): Promise<CloudService> {
    if (data.getTestdata()) {
        return Promise.resolve({
            id: "5",
            name: name,
            description: description,
            requirements: { requirementIds: [] }
        });
    } else {
        return REST.registerCloudService({
            name: name,
            description: description,
        })
    }
}

/**
 * Removes a certain cloud service.
 * 
 * @param serviceId the cloud service id
 */
export async function removeCloudService(serviceId: string): Promise<void> {
    if (data.getTestdata()) {
        return Promise.resolve();
    } else {
        return REST.removeCloudService(serviceId);
    }
}

/**
 * Updates a certain cloud service.
 */
export async function updateCloudService(serviceId: string, service: CloudService): Promise<CloudService> {
    if (data.getTestdata()) {
        return Promise.resolve(service);
    } else {
        return REST.updateCloudService(serviceId, service);
    }
}

/**
 * Returns the monitoring status of a cloud service.
 * 
 * @param serviceId the cloud service id
 * @returns the list of monitored controls. Empty, if monitoring is turned off. 
 */
export async function getMonitoringStatus(serviceId: string): Promise<MonitoringStatus> {
    if (data.getTestdata()) {
        let controlIds = ["Control-1", "Control-2"];
        let random = controlIds.sort(() => .5 - Math.random()).slice(0, Math.random() * controlIds.length);

        return Promise.resolve({
            serviceId: serviceId,
            controlIds: random,
        })
    } else {
        return REST.getMonitoringStatus(serviceId);
    }
}

/**
 * Triggers the monitoring of certain controls for a particular cloud service.
 * 
 * @param serviceId the cloud service id
 * @param controlIds the controls to monitor
 * @returns the current monitoring status immediately after the start
 */
export async function startMonitoring(serviceId: string, controlIds: string[]): Promise<MonitoringStatus> {
    if (data.getTestdata()) {
        return Promise.resolve({
            serviceId: serviceId,
            controlIds: controlIds,
        })
    } else {
        return REST.startMonitoring(serviceId, {
            serviceId: serviceId,
            controlIds: controlIds
        }).then(res => res.status);
    }
}

/**
 * Lists compliance results for a specific service.
 * 
 * @param serviceId the cloud service id
 * @returns a list of compliance results
 */
export async function listCompliance(serviceId: string): Promise<Compliance[]> {
    if (data.getTestdata()) {
        return Promise.resolve([
            { id: "1", serviceId: "1", controlId: "Control-1", status: true, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "2", serviceId: "1", controlId: "Control-2", status: true, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "3", serviceId: "2", controlId: "Control-1", status: false, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "4", serviceId: "2", controlId: "Control-2", status: true, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "5", serviceId: "3", controlId: "Control-1", status: false, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "6", serviceId: "3", controlId: "Control-2", status: false, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "7", serviceId: "4", controlId: "Control-1", status: true, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "8", serviceId: "4", controlId: "Control-2", status: false, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
        ])
    } else {
        return REST.listCompliance(serviceId).then(res => res.complianceResults);
    }
}

/**
 * Lists compliance results for a specific service.
 * 
 * @param serviceId the cloud service id
 * @param day date for the results
 * @returns a list of compliance results
 */
export async function listComplianceForRange(serviceId: string): Promise<Compliance[]> {
    if (data.getTestdata()) {
        return Promise.resolve([
            { id: "1", serviceId: "1", controlId: "Control-1", status: true, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "2", serviceId: "1", controlId: "Control-2", status: true, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "3", serviceId: "2", controlId: "Control-1", status: false, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "4", serviceId: "2", controlId: "Control-2", status: true, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "5", serviceId: "3", controlId: "Control-1", status: false, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "6", serviceId: "3", controlId: "Control-2", status: false, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "7", serviceId: "4", controlId: "Control-1", status: true, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
            { id: "8", serviceId: "4", controlId: "Control-2", status: false, evidenceId: "5", time: new Date().toISOString(), evaluations: [] },
        ])
    } else {
        return REST.listCompliance30Days(serviceId).then(res => res.complianceResults);
    }
}

/**
 * Lists compliance results grouped by the control.
 * 
 * @param serviceId the cloud service id
 * @returns a map of controls and compliance results
 */
export async function listGroupedCompliance(serviceId: string): Promise<Map<string, Compliance[]>> {
    const results = await listCompliance(serviceId);

    // Group the results by control
    var groupedResult = results.reduce((all: Map<string, Compliance[]>, current: Compliance) => {
        let controls: Compliance[] = all.get(current.controlId) ?? [];
        controls.push(current);
        all.set(current.controlId, controls);

        return all;
    }, new Map<string, Compliance[]>());

    return groupedResult;
}

/**
 * Lists compliance results grouped by the control.
 * 
 * @param serviceId the cloud service id
 * @param startDay start of time frame
 * @param endDay end of time frame
 * @returns a map of controls and compliance results
 */
export async function listGroupedComplianceForRange(serviceId: string, startDay: number, endDay: number): Promise<Map<string, Compliance[]>> {
    // Summarize results for selected time period
    const results = await listComplianceForRange(serviceId);
    // Remove empty results
    const r = results.map(function (t) {
        if (t !== null) return t;
    });

    // Remove all results which are not from requested time period
    var startDate = new Date();
    startDate.setDate(startDate.getDate() - startDay + 1);
    startDate.setHours(0);
    startDate.setMinutes(0);
    var endDate = new Date();
    endDate.setDate(endDate.getDate() - endDay + 2);
    endDate.setHours(0);
    endDate.setMinutes(0);

    if (startDate > endDate) {
        var tmp = startDate;
        startDate = endDate;
        endDate = tmp;
    }


    var filteredResults = r.map(function (e) {
        let d = new Date(e.time);
        if (d >= startDate && d <= endDate) return e
    })

    // Summarize remaining evaluations
    // Mark as false if one result in time period is not compliant
    var summs = [];
    var sum = {
        serviceId: String,
        controlId: String,
        allTrue: Boolean
    }
    filteredResults.forEach(ev => {
        // if control + service not already included: add
        if (summs.filter(e => e.serviceId == ev.serviceId && e.controlId == ev.controlId).length < 1) {
            var tmp = {
                serviceId: ev.serviceId,
                controlId: ev.controlId,
                // if one false result: allTrue is false
                allTrue: !(filteredResults.filter(e => e.serviceId == ev.serviceId && e.controlId == ev.controlId && ev.status == false).length > 0)
            }
            summs.push(tmp);
        }
    });


    // Add most recent evaluation with corresponding result (only true if all results in time frame were true) to results
    const summarizedResults = []
    summs.forEach(e => {
        var results = filteredResults.filter(fr => {
            return (fr.controlId == e.controlId && fr.serviceId == e.serviceId && fr.status == e.allTrue)
        });

        var mostRecentDate = new Date(Math.max.apply(null, results.map(e => {
            return new Date(e.time);
        })));
        var mostRecentObject = results.filter(e => {
            var d = new Date(e.time);
            return d.getTime() == mostRecentDate.getTime();
        })[0];
        summarizedResults.push(mostRecentObject);

    })



    // Group the results by control
    var groupedResult = summarizedResults.reduce((all: Map<string, Compliance[]>, current: Compliance) => {
        let controls: Compliance[] = all.get(current.controlId) ?? [];
        controls.push(current);
        all.set(current.controlId, controls);

        return all;
    }, new Map<string, Compliance[]>());

    return groupedResult;
}

/**
 * Retrieves the latest (in time) compliance results grouped by control ID.
 * 
 * @param groupedResults the compliance results grouped by control ID
 * @returns the latest (in time) compliance results grouped by control ID
 */
export function retrieveLatestCompliance(groupedResults: Map<string, Compliance[]>): Map<string, Compliance> {
    let latest = new Map<string, Compliance>();

    // The first element is always the "latest"
    groupedResults.forEach((value, key) => latest.set(key, value[0]));

    return latest;
}

/**
 * Checks, if a service is compliance depnding on the compliance status of its controls.
 * 
 * @param latest the latest compliance results grouped by control
 * @returns the compliance status
 */
export function isServiceCompliant(latest: Map<string, Compliance>): ComplianceStatus {
    let statuses = Array.from(latest.values()).map(isControlCompliant)

    // If we have no results that are either compliant or non-compliant, we are still pending
    if (!statuses.find((s) => s == ComplianceStatus.Compliant || s == ComplianceStatus.NonComplaint)) {
        return ComplianceStatus.Pending;
    }

    // If any control is non-compliant, the whole service is non-compliant
    if (statuses.find((s) => s == ComplianceStatus.NonComplaint)) {
        return ComplianceStatus.NonComplaint;
    }

    // Otherwise, the service is compliant
    return ComplianceStatus.Compliant;
}

/**
 * Checks, if a control is compliant. This is the case when the status is true AND if there are any evaluation results.
 * In the future phase 2 specification, we will deal with this differently and convert the status field into a enum.
 *
 * If the status is true and there are no evaluation results (yet), the status is pending, otherwise it is
 * non-compliant.
 *
 * @param result the compliance compliance result
 * @returns the compliance status
 */
export function isControlCompliant(result: Compliance): ComplianceStatus {
    if (result.status && result.evaluations.length > 0) {
        return ComplianceStatus.Compliant;
    } else if (result.status && result.evaluations.length == 0) {
        return ComplianceStatus.Pending;
    } else {
        return ComplianceStatus.NonComplaint;
    }
}

/**
 * Retrieves the compliance result for a particular service and control.
 * 
 * @param serviceId the cloud service id
 * @param controlId the control id
 * @returns the compliance result
 */
export async function getCompliance(serviceId: string, controlId: string): Promise<Compliance> {
    if (data.getTestdata()) {
        return Promise.resolve(
            { id: "1", serviceId: "1", controlId: "Control-1", status: true, evidenceId: "5", time: new Date().toISOString(), evaluations: [] }
        )
    } else {
        return REST.getCompliance(serviceId, controlId);
    }
}

/**
 * Lists all controls.
 * 
 * @returns returns a list of all controls
 */
export async function listControls(): Promise<Control[]> {
    if (data.getTestdata()) {
        return Promise.resolve([{ id: "Control-1", name: 'Control1', metrics: [] }, { id: "Control-2", name: 'Control2', metrics: [] }]);
    }
    else {
        return REST.listControls();
    }
}

/**
 * Retrieves a particular evidence.
 * 
 * @param controlId the evidence id
 * @result the evidence
 */
export async function getEvidence(evidenceId: string): Promise<Evidence> {
    if (data.getTestdata()) {
        return Promise.resolve({ id: "1", value: {}, targetResource: "1", targetService: "1", toolId: "2", gatheredAt: new Date().toISOString() })
    } else {
        return REST.getEvidence(evidenceId);
    }
}

/**
 * Retrieves a particular metric configuration.
 */
export async function getMetricConfiguration(serviceId: string, metricId: string): Promise<MetricConfiguration> {
    if (data.getTestdata()) {
        return Promise.resolve({ operator: "==", targetValue: 5, isDefault: true, serviceId: "1", metricId: "1" })
    } else {
        return REST.getMetricConfiguration(serviceId, metricId);
    }
}

export async function updateMetricConfiguration(serviceId: string, metricId: string, config: MetricConfiguration): Promise<MetricConfiguration> {
    if (data.getTestdata()) {
        return Promise.resolve({ operator: "==", targetValue: 5, isDefault: true, serviceId: "1", metricId: "1" })
    } else {
        return REST.updateMetricConfiguration(serviceId, metricId, config);
    }
}

export interface ServiceComplianceData {
    latest: Map<string, Compliance>
    latestDate?: Date,
    chartdata: { labels: string[], datasets: { backgroundColor: string[], data: number[] }[] }
    status: ComplianceStatus
}

export async function calculateCompliance(service: CloudService, startDay?: number, endDay?: number): Promise<ServiceComplianceData> {
    let groupedResults: Map<string, Compliance[]>;

    if (startDay) {
        // for history view
        groupedResults = await listGroupedComplianceForRange(
            service.id,
            startDay,
            endDay
        );
    } else {
        // for most recent evaluation
        groupedResults = await listGroupedCompliance(service.id);
    }


    const latest = await retrieveLatestCompliance(groupedResults);

    const latestValues = Array.from(latest.values());
    const numCompliant = latestValues.filter(
        (value) => isControlCompliant(value) == ComplianceStatus.Compliant
    ).length;
    const numNonCompliant = latestValues.filter(
        (value) => isControlCompliant(value) == ComplianceStatus.NonComplaint
    ).length;
    const numPending =
        (service.requirements?.requirementIds ?? []).length -
        (numCompliant + numNonCompliant);

    const chartdata = {
        labels: ["Compliant", "Not Compliant", "Pending"],
        datasets: [
            {
                backgroundColor: ["#198754", "#dc3545", "rgb(108, 117, 125)"],
                data: [numCompliant, numNonCompliant, numPending],
            },
        ],
    };

    const latestDate = Array.from(latest.values()).map(comp => new Date(comp.time)).sort((a, b) => { return a > b ? -1 : 1 })[0] ?? null;

    return {
        latest,
        latestDate,
        chartdata,
        status: isServiceCompliant(latest)
    }
}


export async function getCompPerDay() {
    // Init array
    var comp = [];
    for (var i = 0; i <= 30; i++) {
        comp.push(0);
    }

    let services = await listCloudServices();

    for (const service of services) {
        let serviceresult = await listComplianceForRange(service.id);

        // If for current day this service has no false evaluation, increment
        for (var i = 0; i <= 30; i++) {
            // get date
            let date = new Date();
            date.setDate(date.getDate() - i);
            let datestr = date.toISOString().substring(0, 10)

            // are all services compliant?
            var allcomp = true;
            // is there at least one result for this day?
            var atleastoneres = false;
            serviceresult.forEach(x => {
                if (x.time.includes(datestr)) atleastoneres = true;
                if (x.time.includes(datestr) && x.status===false) allcomp = false;
            })

            if (allcomp && atleastoneres) comp[i] = comp[i] + 1;
        }
    }

    return comp;
}