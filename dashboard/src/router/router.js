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

// Defines the routes of the page and the corresponding views
import { createRouter, createWebHashHistory } from 'vue-router'
import Settings from '../views/Settings.vue'
import Evaluation from '../views/Evaluation.vue'
import EvaluationDetail from "../views/EvaluationDetail.vue"
import ConfigurationDetail from "../views/ConfigurationDetail.vue"
import Configuration from '../views/Configuration.vue'
import History from '../views/History.vue'
import Loggedin from '../views/Loggedin.vue'
import Loggedout from '../views/Loggedout.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      redirect: "/evaluation",
    },
    {
      path: '/settings',
      name: 'Settings',
      component: Settings
    },
    {
      path: '/evaluation',
      name: 'Evaluation',
      component: Evaluation
    },
    {
      path: '/evaluation/:id',
      name: 'EvaluationDetail',
      component: EvaluationDetail,
      props: true
    },
    {
      path: '/configuration',
      name: 'Configuration',
      component: Configuration
    },
    {
      path: '/configuration/:id',
      name: 'ConfigurationDetail',
      component: ConfigurationDetail,
      props: true
    },
    {
      path: '/history',
      name: 'History',
      component: History
    },
    {
      path: '/loggedin',
      name: 'Loggedin',
      component: Loggedin
    },
    {
      path: '/loggedout',
      name: 'Loggedout',
      component: Loggedout
    }
  ]
})

export default router