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

// Is used to communicate with the API
import * as data from '../data/data'
import * as auth from '../auth/auth'

// URL is used to get the data via REST. By default, we assume that the
// dashboard is deployed relative to the api-gateway on the same domain.
const URL = "/v1/";

//REST test
export async function test() {
    var res;
    await getDataFromApi("configuration/controls").then(resp => {
        res = resp.requirements[0].id
    });
    console.log(res)
    return res;

}


// 
// Get
//

// Get live evaluation data
export async function getControlNameById(id) {
    var controls = await getAllControls();
    controls.forEach(c => {
        if (c.id == id) return c.name;
    });
}

export async function getServiceNameById(id) {
    var controls = await getAllServices();
    controls.forEach(c => {
        if (c.id == id) return c.name;
    });
}

export async function getEvaluation() {
    if (!data.getTestdata()) {
        // First get ids of current services
        let ids = await getAllServices();
        // Get evaluation for each service
        var evalRes = [];
        ids.forEach(service => {
            // get evaluation result from api
            var evalService;
            getDataFromApi("evaluation/cloud_services/" + service.id + "/compliance").then(resp => {
                evalService = JSON.parse(resp);

                // resturcture controls
                var tmpcontrols;
                evalService.complianceResults.forEach(element => {
                    var tmpcontrol;
                    tmpcontrol.controlId = element.ControlId;
                    tmpcontrol.controlName = getControlNameById(element.controlIdid);
                    tmpcontrol.controlRes = element.status;
                    tmpcontrol.controlMessage = element.status;
                    tmpcontrol.controlEvidenceId = element.evaluations[0].evidenceId;

                    tmpcontrols.push(tmpcontrol);
                });

                // restructure evaluation result for one service
                var tmp =
                {
                    serviceId: service.id,
                    serviceName: getServiceNameById(service.id),
                    controls: tmpcontrols,
                };
                evalRes.push(tmp);
            });
        });
        return evalRes;
    } else {
        return evalResTest;
    }
}

// Get evaluation data of last 30 days
// Used for history view

var currentHistory = null
var lastUpdated = new Date('2022-07-01T00:00:00')

export async function updateHistory() {
    var evalRes30 = [];
    // init evalRes30
    for (let i = 1; i <= 30; i++) {
        var dayRes = {
            day: i,
            data: []
        }
        evalRes30.push(dayRes)
    }
    // Get current services
    var services = await getAllServices();
/*     console.log('services')
    console.log(services) */
    for (let service of services) {
        var currentServiceEval = (await getDataFromApi(`evaluation/cloud_services/${service.id}/compliance?order_by=time&asc=false&days=30`)).complianceResults;
 /*        console.log(`evaluation/cloud_services/${service.Id}/compliance?order_by=time&asc=false&days=30`)
        console.log ((await getDataFromApi(`evaluation/cloud_services/${service.id}/compliance?order_by=time&asc=false&days=30`)))
        console.log('current eval')
        console.log(currentServiceEval) */
        if (currentServiceEval != null) {
            // Transform data from API in same format as testdata (to reuse functions for diagram, most compliant service etc.)
            // Group by day
            for (let i = 1; i <= 30; i++) {
                var ResultsForDay = [];
                var day = new Date();
                day.setDate(day.getDate() - i);
                const date =
                    day.getFullYear() +
                    "-" +
                    (day.getMonth() + 1) +
                    "-" +
                    day.getDate();
                currentServiceEval.forEach(complianceResult => {
                    if (complianceResult.time.includes(date)) {
                        ResultsForDay.push(complianceResult)
                    }
                })

                // Group control:
                var checkedControls = new Map()
                ResultsForDay.forEach(day => {
                    if (!checkedControls.has(day.controlId)) {
                        // Add control to map
                        checkedControls.set(day.controlId, day.status)
                    }
                    else {
                        // Set control to false if one result is false
                        if (day.status === false) checkedControls.set(day.controlId, false)
                    }

                })


                // Summaraize result for day and service
                var groupedControls = [];
                checkedControls.forEach((id, status) => {
                    var tmp = {
                        controlId: id,
                        controlName: getControlNameById(id),
                        controlRes: status,
                    }
                    groupedControls.push(tmp);
                })



                var dayServiceRes = {
                    serviceId: service.Id,
                    serviceName: getServiceNameById(service.id),
                    controls: groupedControls
                }

                // Push result to corresponding day in evalRes30
                var dayIndex = evalRes30.findIndex(item => item.day === i);;
                evalRes30[dayIndex].data.push(dayServiceRes)
            }
        }
    };

    currentHistory = evalRes30;
    lastUpdated = new Date
/*     console.log('new history')
    console.log(evalRes30) */
}

export async function getHistory() {
    if (!data.getTestdata()) {
        // Update history if older than one minute
        const now = new Date
        const diff = parseInt(now - lastUpdated)
        if (currentHistory === null || diff > 1000 * 60 /* || true */) {  // add || true for local debigging
            await updateHistory();
            //console.log('updating history')
        }
        return currentHistory;
    } else {
        return evalResTest30;
    }
}

// Get lists of ALL services / controls
export async function getAllServices() {
    if (!data.getTestdata()) {
        var res = [];
        await getDataFromApi("configuration/cloud_services").then(resp => {
            resp.services.forEach(element => {
                res.push(element);
            });
        });
        return res;
    } else {
        return Promise.resolve([
            { name: 'Service-1', id: '1' },
            { name: 'Service-2', id: '2' },
            { name: 'Service-3', id: '3' },
            { name: 'Service-4', id: '4' }]
        );
    }
}


export async function getAllControls() {
    if (!data.getTestdata()) {
        var res = [];
        await getDataFromApi("configuration/controls").then(resp => {
            resp.requirements.forEach(element => {
                res.push(element);
            });
        });
        return res;
    } else {
        return [{ id: "Control-1", name: 'Control1' }, { id: "Control-2", name: 'Control2' }];
    }
}


//
// set; put; delelete
//

// Set services to monitor and controls
export function setMonitoring(sid, status, controls) {
    if (!data.getTestdata()) {
        if (!status) {
            var d = "";
            var path = "configuration/monitoring/" + sid + "/stop";
            setDataToApi(path, d);
        } else {
            var d = {
                serviceId: sid,
                controlIds: controls,
            };
            var path = "configuration/monitoring/" + sid + "/start";
            setDataToApi(path, d);
            console.log("start monitoring");
        }
    } else {
        alert("Testmode: Controls set");
    }
}


//
// API 
//

export async function getDataFromApi(path) {
    var res = "";
    var request = new Request(URL + path, {
        method: 'GET',
        headers: {
            'Authorization': 'Bearer ' + auth.getToken(),
            //  "Access-Control-Allow-Origin": "*",
        }
    });
    return await fetch(request)
        .then((resp) => resp.json())
        .catch((error) => {
            console.log(error);
        });
}

export async function setDataToApi(path, d) {
    var request = new Request(URL + path, {
        method: 'POST',
        body: JSON.stringify(d),
        headers: {
            'Authorization': 'Bearer ' + auth.getToken(),
            //  "Access-Control-Allow-Origin": "*",
        }
    });

    return fetch(request).then(res => res.json());
}

export async function putDataToApi(path, d) {
    var request = new Request(URL + path, {
        method: 'PUT',
        body: JSON.stringify(d),
        headers: {
            'Authorization': 'Bearer ' + auth.getToken(),
            //  "Access-Control-Allow-Origin": "*",
        }
    });
    return fetch(request).then(res => res.json());
}

export async function deleteDataToApi(path) {
    var request = new Request(URL + path, {
        method: 'DELETE',
        headers: {
            'Authorization': 'Bearer ' + auth.getToken(),
            //  "Access-Control-Allow-Origin": "*",
        }
    });
    console.log(request.headers);
    return await fetch(request)
        .then((resp) => resp.json())
        .catch((error) => {
            console.log(error);
        });


}


/***************TESTDATA***************/
// Testdata for evaluation with 4 services
const evalResTest = [
    {
        serviceId: "1",
        serviceName: "Service1",
        controls: [
            {
                controlId: "1",
                controlName: "Control1",
                controlRes: false,
                controlMessage: "not ok",
            },
            {
                controlId: "2",
                controlName: "Control2",
                controlRes: true,
                controlMessage: "ok",
            }
        ]
    },
    {
        serviceId: "2",
        serviceName: "Service2",
        controls: [
            {
                controlId: "1",
                controlName: "Control1",
                controlRes: false,
                controlMessage: "not ok",
            },
            {
                controlId: "2",
                controlName: "Control2",
                controlRes: true,
                controlMessage: "ok",
            },
            {
                controlId: "3",
                controlName: "Control3",
                controlRes: true,
                controlMessage: "ok",
            }
        ]
    },
    {
        serviceId: "3",
        serviceName: "Service3",
        controls: [
            {
                controlId: "1",
                controlName: "Control1",
                controlRes: false,
                controlMessage: "not ok",
            },
            {
                controlId: "2",
                controlName: "Control2",
                controlRes: true,
                controlMessage: "ok",
            }
        ]
    },
    {
        serviceId: "4",
        serviceName: "Service4",
        controls: [
            {
                controlId: "1",
                controlName: "Control1",
                controlRes: true,
                controlMessage: "ok",
            },
            {
                controlId: "2",
                controlName: "Control2",
                controlRes: true,
                controlMessage: "ok",
            }
        ]
    }
];

// Testdata for 30 last days 
const evalResTest30 = [
    {
        day: 1,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 2,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 3,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    }, {
        day: 4,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 5,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 6,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    }, {
        day: 7,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 8,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 9,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    }, {
        day: 10,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 11,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 12,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    }, {
        day: 13,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 14,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 15,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    }, {
        day: 16,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 17,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 18,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    }, {
        day: 19,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 20,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 21,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    }, {
        day: 22,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 23,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 24,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    }, {
        day: 25,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 26,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 27,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    }, {
        day: 28,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 29,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: false,
                        controlMessage: "not ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    },
    {
        day: 30,
        data: [
            {
                serviceId: "1",
                serviceName: "Service1",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "2",
                serviceName: "Service2",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "3",
                        controlName: "Control3",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "3",
                serviceName: "Service3",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            },
            {
                serviceId: "4",
                serviceName: "Service4",
                controls: [
                    {
                        controlId: "1",
                        controlName: "Control1",
                        controlRes: true,
                        controlMessage: "ok",
                    },
                    {
                        controlId: "2",
                        controlName: "Control2",
                        controlRes: true,
                        controlMessage: "ok",
                    }
                ]
            }
        ]
    }];
