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
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
)

type storageDiscovery struct {
	*Discovery
}

// NewStorageDiscovery creates a new OpenStack discoverer for storage resources based on the
// options provided in opts. WithAuthOpts is mandatory and must be provided.
func NewStorageDiscovery(opts ...DiscoveryOption) discovery.Discoverer {
	d := &storageDiscovery{&Discovery{}}

	for _, opt := range opts {
		opt(d.Discovery)
	}

	// WithAuthOpts is mandatory, since it cannot be checked directly whether WithAuthOpts was passed, we check if authOpts is set before returning the discoverer
	if d.authOpts == nil {
		return nil
	}

	return d
}

func (*storageDiscovery) Name() string {
	return "OpenStack Storage"
}

func (*storageDiscovery) Description() string {
	return "Discovery OpenStack Storage."
}

// List lists OpenStack storage resources and translates them into the Clouditor ontology
func (d *storageDiscovery) List() (list []voc.IsCloudResource, err error) {
	// Discover volumes
	volumes, err := d.discoverVolumes()
	if err != nil {
		return nil, fmt.Errorf("could not discover volumes: %w", err)
	}
	list = append(list, volumes...)

	return
}

func (d *storageDiscovery) discoverVolumes() (list []voc.IsCloudResource, err error) {
	var opts volumes.ListOptsBuilder = volumes.ListOpts{}
	list, err = genericList(d.Discovery, d.storageClient, volumes.List, d.handleVolume, volumes.ExtractVolumes, opts)

	return
}

// handleVolume creates a block storage resource based on the Clouditor Ontology
func (d *storageDiscovery) handleVolume(volume *volumes.Volume) (r *voc.BlockStorage, err error) {
	r = &voc.BlockStorage{
		Storage: &voc.Storage{
			Resource: &voc.Resource{
				ID:           voc.ResourceID(volume.ID),
				Name:         volume.Name,
				CreationTime: volume.CreatedAt.Unix(),
			},
			AtRestEncryption: voc.AtRestEncryption{Enabled: volume.Encrypted},
		},
	}

	return
}
