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

package openstack

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	volumestesting "github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes/testing"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	serverstesting "github.com/gophercloud/gophercloud/openstack/compute/v2/servers/testing"
	"github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
)

var (
	HandleServerListSuccessfully  = serverstesting.HandleServerListSuccessfully
	HandleVolumesListSuccessfully = volumestesting.MockListResponse

	Volume = volumes.Volume{
		Name:      "vol-001",
		ID:        "289da7f8-6440-407c-9fb4-7db01ec49164",
		Encrypted: false,
	}

	ServerHerp = serverstesting.ServerHerp
	ServerDerp = serverstesting.ServerDerp
	ServerMerp = serverstesting.ServerMerp

	ServerHerpPortID = "10bac140-7239-4cde-8b04-402a0124bee8"
	ServerDerpPortID = "804d86fc-1498-45f1-89ab-25ff325595af"
	ServerMerpPortID = "b5d09f0a-84d8-4127-8041-566511ae0bf9"

	SubnetID = "36682d2f-c267-44d4-90af-f60f38787ed1"
)

func HandleInterfaceListSuccessfully(t *testing.T) {
	// Unfortunately, the internal testing fixtures of gophercloud are not consistent,
	// i.e., they are not using the same resource IDs across their services. This makes
	// it hard for our unit tests, since we rely on retrieving a consistent picture of a
	// resource, such as a virtual machine, across different services such as networking,
	// storage and compute. Therefore we need to set up our own interface list for the test servers
	// instead of relying on the test fixture in compute/v2/extensions/attachinterfaces/testing.
	servers := []servers.Server{
		ServerHerp,
		ServerDerp,
		ServerMerp,
	}

	portIDs := []string{
		ServerHerpPortID,
		ServerDerpPortID,
		ServerMerpPortID,
	}

	for i, server := range servers {
		portID := portIDs[i]

		testhelper.Mux.HandleFunc(fmt.Sprintf("/servers/%s/os-interface", server.ID), func(w http.ResponseWriter, r *http.Request) {
			testhelper.TestMethod(t, r, "GET")
			testhelper.TestHeader(t, r, "X-Auth-Token", client.TokenID)

			w.Header().Add("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"interfaceAttachments": [
					{
						"port_state": "ACTIVE",
						"fixed_ips": [
							{
								"subnet_id": "%s",
								"ip_address": "%s"
							}
						],
						"port_id": "%s"
					}
				]
			}`,
				SubnetID,
				server.Addresses["private"].([]interface{})[0].(map[string]interface{})["addr"],
				portID,
			)
		})
	}
}
