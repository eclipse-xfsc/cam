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
	"reflect"
	"testing"

	"clouditor.io/clouditor/api/discovery"
	"clouditor.io/clouditor/voc"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
	"github.com/stretchr/testify/assert"
)

func TestStorageDiscovery(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	HandleVolumesListSuccessfully(t)

	p := &gophercloud.ProviderClient{
		TokenID: client.TokenID,
		EndpointLocator: func(eo gophercloud.EndpointOpts) (string, error) {
			return testhelper.Endpoint(), nil
		},
	}

	// WithAuthOpts is mandatory
	a := &AuthOptions{
		IdentityEndpoint: fmt.Sprintf("https://%s:%s/v2.0", "identityHost", "portNumber"),
		Username:         "username",
		Password:         "passwort",
		TenantName:       "tenant",
		AllowReauth:      true,
	}

	d := NewStorageDiscovery(WithProvider(p), WithAuthOpts(a))

	list, err := d.List()
	assert.Nil(t, err)

	block, ok := list[0].(*voc.BlockStorage)

	assert.True(t, ok)
	assert.Equal(t, Volume.ID, string(block.ID))
	assert.Equal(t, Volume.Name, block.Name)
	assert.Equal(t, Volume.Encrypted, block.AtRestEncryption.GetAtRestEncryption().Enabled)
}

func TestNewStorageDiscovery(t *testing.T) {
	type args struct {
		opts []DiscoveryOption
	}
	tests := []struct {
		name string
		args args
		want discovery.Discoverer
	}{
		{
			name: "Missing WithAuthOpts",
			args: args{
				opts: []DiscoveryOption{},
			},
			want: nil,
		},
		{
			name: "With WithAuthOpts",
			args: args{
				opts: []DiscoveryOption{
					WithAuthOpts(&AuthOptions{
						IdentityEndpoint: "https://identityHost:portNumber/v2.0",
						Username:         "username",
						Password:         "passwort",
						TenantName:       "tenant",
						AllowReauth:      true,
					})},
			},
			want: &storageDiscovery{&Discovery{
				authOpts: &gophercloud.AuthOptions{
					IdentityEndpoint: "https://identityHost:portNumber/v2.0",
					Username:         "username",
					Password:         "passwort",
					TenantName:       "tenant",
					AllowReauth:      true,
				}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStorageDiscovery(tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStorageDiscovery() = %v, want %v", got, tt.want)
			}
		})
	}
}
