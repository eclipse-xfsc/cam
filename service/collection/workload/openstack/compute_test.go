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

func TestComputeDiscovery(t *testing.T) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	HandleServerListSuccessfully(t)
	HandleInterfaceListSuccessfully(t)

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

	d := NewComputeDiscovery(WithProvider(p), WithAuthOpts(a))

	list, err := d.List()
	assert.Nil(t, err)

	virtualMachine, ok := list[0].(*voc.VirtualMachine)

	assert.True(t, ok)
	assert.Equal(t, ServerHerp.ID, string(virtualMachine.ID))
	assert.Equal(t, ServerHerp.Name, virtualMachine.Name)
	assert.Equal(t, 1, len(virtualMachine.NetworkInterface))

	portID := virtualMachine.NetworkInterface[0]
	assert.Equal(t, voc.ResourceID(ServerHerpPortID), portID)

	assert.Equal(t, 1, len(virtualMachine.BlockStorage))
}

func TestNewComputeDiscovery(t *testing.T) {
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
			want: &computeDiscovery{&Discovery{
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
			if got := NewComputeDiscovery(tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewComputeDiscovery() = %v, want %v", got, tt.want)
			}
		})
	}
}
