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
	"os"

	"clouditor.io/clouditor/voc"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("component", "openstack-discovery")

const (
	RegionName = "OS_REGION_NAME"
)

type Discovery struct {
	provider *gophercloud.ProviderClient
	compute  *gophercloud.ServiceClient
	storage  *gophercloud.ServiceClient
	authOpts *gophercloud.AuthOptions
}

type AuthOptions struct {
	IdentityEndpoint string `json:"identityEndpoint"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	TenantName       string `json:"tenantName"`
	AllowReauth      bool   `json:"allowReauth"`
}

type DiscoveryOption func(*Discovery)

// WithAuthOpts is an option to set the authentication options
func WithAuthOpts(o *AuthOptions) DiscoveryOption {
	return func(d *Discovery) {
		d.authOpts = &gophercloud.AuthOptions{
			IdentityEndpoint: o.IdentityEndpoint, // "https://identityHost:portNumber/v2.0"
			Username:         o.Username,
			Password:         o.Password,
			TenantName:       o.TenantName,
			AllowReauth:      o.AllowReauth,
		}
	}
}

func WithProvider(p *gophercloud.ProviderClient) DiscoveryOption {
	return func(d *Discovery) {
		d.provider = p
	}
}

// computeClient returns the compute client if initialized
func (d *Discovery) computeClient() (client *gophercloud.ServiceClient, err error) {
	if d.compute == nil {
		return nil, fmt.Errorf("compute client not initialized")
	}
	return d.compute, nil
}

// storageClient returns the compute client if initialized
func (d *Discovery) storageClient() (client *gophercloud.ServiceClient, err error) {
	if d.storage == nil {
		return nil, fmt.Errorf("storage client not initialized")
	}
	return d.storage, nil
}

// authorize authorizes to Openstack and asserts the following clients
// * compute client
// * block storage client
func (d *Discovery) authorize() (err error) {

	if d.provider == nil {
		d.provider, err = openstack.AuthenticatedClient(*d.authOpts)

		if err != nil {
			return fmt.Errorf("error while authenticating: %w", err)
		}
	}

	if d.compute == nil {
		d.compute, err = openstack.NewComputeV2(d.provider, gophercloud.EndpointOpts{
			Region: os.Getenv(RegionName),
		})

		if err != nil {
			return fmt.Errorf("could not create compute client: %w", err)
		}
	}

	if d.storage == nil {
		d.storage, err = openstack.NewBlockStorageV3(d.provider, gophercloud.EndpointOpts{
			Region: os.Getenv(RegionName),
		})

		if err != nil {
			return fmt.Errorf("could not create block storage client: %w", err)
		}
	}

	return
}

type ClientFunc func() (*gophercloud.ServiceClient, error)
type ListFunc[O any] func(client *gophercloud.ServiceClient, opts O) pagination.Pager
type HandlerFunc[T any, R voc.IsCloudResource] func(in *T) (r R, err error)
type ExtractorFunc[T any] func(r pagination.Page) ([]T, error)

// genericList is a function leveraging type parameters that takes care of listing OpenStack
// resources using a ClientFunc, which returns the needed client, a ListFunc l, which returns paginated results,
// an extractor that extracts the results into gophercloud specific objects and a handler which converts them
// into an appropriate Clouditor vocabulary object.
func genericList[T any, O any, R voc.IsCloudResource](d *Discovery, clientGetter ClientFunc,
	l ListFunc[O],
	handler HandlerFunc[T, R],
	extractor ExtractorFunc[T],
	opts O,
) (list []voc.IsCloudResource, err error) {
	if err = d.authorize(); err != nil {
		return nil, fmt.Errorf("could not authorize openstack: %w", err)
	}

	client, err := clientGetter()
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	err = l(client, opts).EachPage(func(p pagination.Page) (bool, error) {
		x, err := extractor(p)

		if err != nil {
			return false, fmt.Errorf("could not extract items from paginated result: %w", err)
		}

		for _, s := range x {
			r, err := handler(&s)
			if err != nil {
				return false, fmt.Errorf("could not convert into Clouditor vocabulary: %w", err)
			}

			log.Debugf("Adding resource %+v", s)

			list = append(list, r)
		}

		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not list resources: %w", err)
	}

	return
}
