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

import jwt_decode from "jwt-decode";

export function login(client) {
    signin(client)
}

export function logout(client) {
    setToken(null);
    signout(client);
}

export function isAuthenticated() {
    let token = localStorage.getItem("token")

    if (token == null) {
        // If the token is empty, we are not authenticated
        return false;
    }

    // Try to parse the token
    try {
        // Check the expiry of the token, to see whether we are still authenticated
        const claims = jwt_decode(token);
        return claims.exp >= new Date().getTime() / 1000;
    }
    catch (err) {
        // If the token is invalid, we are definitely not authenticated
        return false;
    }
}

export function setToken(ptoken) {
    localStorage.setItem('token', ptoken)
}

export function getToken() {
    return localStorage.getItem('token')
}

export function setIdToken(ptoken) {
    localStorage.setItem('id_token', ptoken)
}

export function getIdToken() {
    return localStorage.getItem('id_token')
}


// Copyright (c) Brock Allen & Dominick Baier. All rights reserved.
// Licensed under the Apache License, Version 2.0. See LICENSE in the project root for license information.

///////////////////////////////
// OidcClient config
///////////////////////////////

function log() {
    Array.prototype.forEach.call(arguments, function (msg) {
        if (msg instanceof Error) {
            msg = "Error: " + msg.message;
        }
        else if (typeof msg !== "string") {
            msg = JSON.stringify(msg, null, 2);
        }
    });
}



///////////////////////////////
// functions for UI elements
///////////////////////////////
function signin(client) {
    client.createSigninRequest({ state: { bar: 15 } }).then(function (req) {
        log("signin request", req, "<a href='" + req.url + "'>go signin</a>");
        // if (followLinks()) {
        window.location = req.url;
        //}
    }).catch(function (err) {
        console.error(err);
        log(err);
    });
}

function signout(client) {
    client.createSignoutRequest({ id_token_hint: getIdToken() }).then(function (req) {
        log("signout request", req, "<a href='" + req.url + "'>go signout</a>");
        // if (followLinks()) {
        window.location = req.url;
        //}
    }).catch(function (err) {
        console.error(err);
        log(err);
    });
}

export {
    log
};