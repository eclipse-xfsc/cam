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

// Is used to get calculate / derive the data to display from the received data via REST

import * as REST from './REST'

// Access global variables:
export function setTestdata(status) {
    // localStorage can only save strings
    if (status == true) localStorage.setItem('testdata', 'true')
    else localStorage.setItem('testdata', 'false')
}

export function getTestdata() {
    if (localStorage.getItem('testdata') == 'true') {
        return true;
    }
    else {
        return false
    }

}

export function getSiteURL() {
    if (localStorage.getItem('runLocal') == 'true') {
        return localStorage.getItem('localSiteURL')
    }
    else {
        return localStorage.getItem('prodURL')
    }
}

// General configuration information


export async function getAllControls() {
    return REST.getAllControls();
}

export function setMonitoring(sid, status, controls) {
    return REST.setMonitoring(sid, status, controls);
}

// Evaluations
export async function test() {
    return await REST.test();
}

export async function getHistory() {
    return REST.getHistory();
}

export function getPercentagePerControl(pevals, pcontrols) {
    var counter = 0;
    var compcounter = 0;
    if (!pevals[0].hasOwnProperty('day')) {
        pevals.forEach((service) => {
            service.controls.forEach((control) => {
                counter++;
                if ((control.controlRes) & (pcontrols.includes(control.controlName))) compcounter++;
            });
        });
    } else {
        pevals.forEach((day) => {
            day.data.forEach((service) => {
                service.controls.forEach((control) => {
                    counter++;
                    if ((control.controlRes) & (pcontrols.includes(control.controlName))) compcounter++;
                });
            });
        });
    }
    return Math.round((compcounter * 100) / counter);
}

export async function getControlsFromEval(pevals) {
    var res = [];
    if (!pevals[0].hasOwnProperty('day')) {
        pevals.forEach((service) => {
            service.controls.forEach((control) => {
                if (!res.includes(control.controlName)) res.push(control.controlName);
            });
        });
    }
    else {
        pevals.forEach((day) => {
            day.data.forEach((service) => {
                service.controls.forEach((control) => {
                    if (!res.includes(control.controlName)) res.push(control.controlName);
                });
            });
        });
    }
    return res;
}

export async function getSupEval(end, start) {
    var res = [];
    var history = [];
    await getHistory().then(resp => {
        resp.forEach(element => {
            history.push(element);
        });
    });;

    if (start > end) {
        var tmp = end;
        end = start;
        start = tmp;
    }
    history.forEach((d) => {
        if (d.day >= start & d.day <= end) res.push(d);
    });

    return res;
}

// Calculated values for history (last 30 days)

// Returns an array with the number of compliant services per day
export async function getCompPerDay() {
    var comp = [];
    var history = [];
    await getHistory().then(resp => {
        resp.forEach(element => {
            history.push(element);
        });
    });;
/*     console.log('history')
    console.log(history) */
    history.forEach((day) => {
        var daycounter = 0;
        day.data.forEach((service) => {
            var allCompl = true
            if (service.controls.length === 0) allCompl = false;
            service.controls.forEach((control) => {
                if (!control.controlRes) allCompl = false;
            });
            if (allCompl) daycounter++;
        });
        comp.push(daycounter);
    });

    // fill up with 0 for days without data
    for (var i = history.length; i <= 30; i++) {
        comp.push(0);
    }
    return comp;
}

// Counts the compliance of each service over all days
export async function countCompliancePerService() {
    var counter = [];
    var comp = [];
    var history = [];
/*     await getHistory().then(resp => {
        resp.forEach(element => {
            history.push(element);
        });
    });;  */
    //history =  await getHistory();
    await getHistory().then(resp => {
        resp.forEach(element => {
            history.push(element);
        });
    });;
    history.forEach((day) => {
        var daycounter = 0;
        day.data.forEach((service) => {
            var allCompl = true
            service.controls.forEach((control) => {
                if (!control.controlRes) allCompl = false;
            });
            if (allCompl) {
                if (counter.some(s => s.id === service.serviceId)) {
                    // update counter
                    counter.find(s => s.id === service.serviceId).count++;
                } else {
                    // add to counter
                    var tmp = new Object;
                    tmp.name = service.serviceName;
                    tmp.id = service.serviceId;
                    tmp.count = 1;
                    counter.push(tmp);
                }
            }
        });
        comp.push(daycounter);
    }); 
    return counter;
}

export async function getDaysCompliant() {
    var compcounter = 0;
    var history = [];
    await getHistory().then(resp => {
        resp.forEach(element => {
            history.push(element);
        });
    });;
    history.forEach((day) => {
        var servicecounter = 0;
        day.data.forEach((service) => {
            var allCompl = true
            if (service.controls.length === 0) allCompl = false;
            service.controls.forEach((control) => {
                if (!control.controlRes) allCompl = false;
            });
            if (allCompl) servicecounter++;
        });
        if (servicecounter == day.data.length) compcounter++;
    });
    return compcounter;
}

export async function getMostCompliantService() {
    var result = "test";
    // find n max
    var max = -1;
    var comp = [];
     await countCompliancePerService().then(resp => {
        resp.forEach(element => {
            comp.push(element);
        });
    });;
    comp.forEach((service) => {
        if (service.count > max) max = service.count;
    });

    // add servicename to result
    comp.forEach((service) => {
        if (service.count == max) {
            if (result == "") {
                result = result + service.name;
            } else {
                result = result + ", " + service.name;
            }

        }
    });
    return result;
}

export async function getLeastCompliantService() {
    var result = "";
    // find n min
    var min = 1000000;
    var comp = [];
    await countCompliancePerService().then(resp => {
        resp.forEach(element => {
            comp.push(element);
        });
    });;
    comp.forEach((service) => {
        if (service.count < min) min = service.count;
    });

    // add servicename to result
    comp.forEach((service) => {
        if (service.count == min) {
            if (result == "") {
                result = result + service.name;
            } else {
                result = result + ", " + service.name;
            }

        }
    });
    return result;
}

export async function getEvalForDay(day) {
    var history = [];
    await getHistory().then(resp => {
        resp.forEach(element => {
            history.push(element);
        });
    });;
    return history.find(e => (e.day == day)).data;
}







