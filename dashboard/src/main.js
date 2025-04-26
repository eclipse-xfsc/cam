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

import "./style.scss"
import "bootstrap/dist/js/bootstrap.js"
import { createApp } from 'vue'
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faGear, faGears, faTrash, faPenToSquare, faBook } from '@fortawesome/free-solid-svg-icons'
import App from './App.vue'
import router from './router/router'
import { Log, OidcClient } from 'oidc-client-ts';
import * as data from './data/data';

Log.setLogger(console);
Log.setLevel(Log.INFO);

library.add(faTrash, faGear, faGears, faPenToSquare, faBook)

const app = createApp(App)
    .component('font-awesome-icon', FontAwesomeIcon)

// Set here if the app is running on localhost; true: localhost
if (localStorage.getItem('runLocal') == null) localStorage.setItem('runLocal', 'false')
if (localStorage.getItem('testdata') == null) localStorage.setItem('testdata', 'false')

if (localStorage.getItem('localSiteURL') == null) localStorage.setItem('localSiteURL', 'http://localhost:3000')
if (localStorage.getItem('localAPIURL') == null) localStorage.setItem('localAPIURL', 'http://localhost:8080')
if (localStorage.getItem('prodURL') == null) localStorage.setItem('prodURL', 'https://cam.gxfs.dev')

fetch("/config.json")
    .then((response) => response.json())
    .then((config) => {
        const settings = {
            authority: config.authority,
            client_id: config.client_id,
            redirect_uri: config.redirect_uri,
            post_logout_redirect_uri: config.post_logout_redirect_uri,
            response_type: "code",
            scope: "profile",
            //monitorSession: false,
            //response_mode: "fragment",
            //filterProtocolClaims: true
        };

        console.log(settings);

        app.provide("client", new OidcClient(settings));

        app.use(router)
        app.mount('#app')
    })
