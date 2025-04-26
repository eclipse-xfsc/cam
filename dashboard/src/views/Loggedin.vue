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

* Landing page for sucessful logins
<script setup lang="ts">
import { settings } from "../auth/settings";
import { log } from "../auth/auth";
import * as auth from "../auth/auth";
import { useRouter, useRoute } from 'vue-router'
import { OidcClient } from "oidc-client-ts";
import { inject } from "vue";

const router = useRouter()
const route = useRoute()
const url = route.fullPath;

var message = "Login sucessful, forwarding...";

const client: OidcClient | undefined = inject("client");

client?.processSigninResponse(url)
  .then(function (response) {
    log("signin response success", response);
    auth.setToken(response.access_token);
    auth.setIdToken(response.id_token);
    location.reload();
  })
  .catch(function (err) {
    console.error(err);
    log(err);
  });

// Go back to start page
router.push("/");

</script>

<template>
  <div id="login">
    <h1>{{ message }}</h1>
  </div>
</template>