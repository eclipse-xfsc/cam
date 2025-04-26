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

	"clouditor.io/clouditor/api/discovery"
	"clouditor.io/clouditor/voc"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/attachinterfaces"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
)

type computeDiscovery struct {
	*Discovery
}

// NewComputeDiscovery creates a new OpenStack discoverer for compute resources based on the
// options provided in opts. WithAuthOpts is mandatory and must be provided.
func NewComputeDiscovery(opts ...DiscoveryOption) discovery.Discoverer {
	d := &computeDiscovery{&Discovery{}}

	for _, opt := range opts {
		opt(d.Discovery)
	}

	// WithAuthOpts is mandatory, since it cannot be checked directly whether WithAuthOpts was passed, we check if authOpts is set before returning the discoverer
	if d.authOpts == nil {
		return nil
	}

	return d
}

func (*computeDiscovery) Name() string {
	return "OpenStack Compute"
}

func (*computeDiscovery) Description() string {
	return "Discovery OpenStack Compute."
}

// List lists OpenStack servers (compute resources) and translates them into the Clouditor ontology
func (d *computeDiscovery) List() (list []voc.IsCloudResource, err error) {
	// Discover servers
	servers, err := d.discoverServers()
	if err != nil {
		return nil, fmt.Errorf("could not discover servers: %w", err)
	}
	list = append(list, servers...)

	return
}

func (d *computeDiscovery) discoverServers() (list []voc.IsCloudResource, err error) {
	// TODO(oxisto): Limit the list to a specific tenant?
	var opts servers.ListOptsBuilder = &servers.ListOpts{}
	list, err = genericList(d.Discovery, d.computeClient, servers.List, d.handleServer, servers.ExtractServers, opts)

	return
}

func (d *computeDiscovery) discoverNetworkInterfaces(serverID string) (list []voc.ResourceID, err error) {
	if err = d.authorize(); err != nil {
		return nil, fmt.Errorf("could not authorize openstack: %w", err)
	}

	err = attachinterfaces.List(d.compute, serverID).EachPage(func(p pagination.Page) (bool, error) {
		ifc, err := attachinterfaces.ExtractInterfaces(p)
		if err != nil {
			return false, fmt.Errorf("could not extract network interface from page: %w", err)
		}

		for _, i := range ifc {
			list = append(list, voc.ResourceID(i.PortID))
		}

		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not list network interfaces: %w", err)
	}

	return
}

// handleServer creates a virtual machine resource based on the Clouditor Ontology
func (d *computeDiscovery) handleServer(server *servers.Server) (r *voc.VirtualMachine, err error) {
	r = &voc.VirtualMachine{
		Compute: &voc.Compute{
			Resource: &voc.Resource{
				ID:           voc.ResourceID(server.ID),
				Name:         server.Name,
				CreationTime: server.Created.Unix(),
				Type:         []string{"VirtualMachine", "Compute", "Resource"},
				GeoLocation: voc.GeoLocation{
					Region: "unknown", // TODO: Can we get the region?
				},
			}},
		BootLogging:  nil,
		OSLogging:    nil,
		BlockStorage: []voc.ResourceID{},
	}

	for _, v := range server.AttachedVolumes {
		r.BlockStorage = append(r.BlockStorage, voc.ResourceID(v.ID))
	}

	r.NetworkInterface, err = d.discoverNetworkInterfaces(server.ID)
	if err != nil {
		return nil, fmt.Errorf("could not discover attached network interfaces: %w", err)
	}

	return
}
