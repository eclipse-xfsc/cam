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

package authsec

import "clouditor.io/clouditor/voc"

// Value represents the Value of an evidence in the case of the Authentication Security CM
type Value struct {
	// Clouditor's Resource properties ID and Types have to be set that Evaluation will not fail
	voc.Resource

	// OAuthGrantTypes metric properties
	*OAuthGrantTypes `json:"oAuth,omitempty"`
	// APIOAuthProtected metric properties
	*APIOAuthProtected `json:"apiOAuthProtected,omitempty"`
}

type OAuthGrantTypes struct {
	GrantTypes                                         []string `json:"grantTypes"`
	IDTokenSigningAlgValuesSupported                   []string `json:"idTokenSigningAlgValuesSupported"`
	UserinfoSigningAlgValuesSupported                  []string `json:"userinfoSigningAlgValuesSupported"`
	RequestObjectSigningAlgValuesSupported             []string `json:"requestObjectSigningAlgValuesSupported"`
	TokenEndpointAuthSigningAlgValuesSupported         []string `json:"tokenEndpointAuthSigningAlgValuesSupported"`
	RevocationEndpointAuthSigningAlgValuesSupported    []string `json:"revocationEndpointAuthSigningAlgValuesSupported"`
	IntrospectionEndpointAuthSigningAlgValuesSupported []string `json:"introspectionEndpointAuthSigningAlgValuesSupported"`
	IDTokenEncryptionAlgValuesSupported                []string `json:"idTokenEncryptionAlgValuesSupported"`
	IDTokenEncryptionEncValuesSupported                []string `json:"idTokenEncryptionEncValuesSupported"`
	UserinfoEncryptionAlgValuesSupported               []string `json:"userinfoEncryptionAlgValuesSupported"`
	UserinfoEncryptionEncValuesSupported               []string `json:"userinfoEncryptionEncValuesSupported"`
	RequestObjectEncryptionAlgValuesSupported          []string `json:"requestObjectEncryptionAlgValuesSupported"`
	RequestObjectEncryptionEncValuesSupported          []string `json:"requestObjectEncryptionEncValuesSupported"`
}

type APIOAuthProtected struct {
	Url    string `json:"url"`
	Status string `json:"status"`
}
