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

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// CheckAPIAccess calls a REST API endpoint, optionally using an OAuth Access token according to RFC 6750
// Returns whether the request was successful, determined by the response code
func CheckAPIAccess(endpoint url.URL, method string, token *oauth2.Token) (bool, error) {
	// We do not use request parameters
	requestParams := url.Values{}

	// Prepare request
	req, err := http.NewRequest(method, endpoint.String(), strings.NewReader(requestParams.Encode()))
	if nil != err {
		return false, err
	}

	// If given a token, use it for authentication (See RFC 6750)
	if nil != token {
		token.SetAuthHeader(req)
	}

	// Fire request
	response, err := http.DefaultClient.Do(req)
	if nil != err {
		return false, err
	}

	// Was the request successful?
	responseCode := response.StatusCode
	return responseCode >= 200 && responseCode <= 299, nil
}

// acquireAccessToken uses the OAuth 2.0 Client Credentials Grant to acquire an Access Token
// Client Authentication is done via client_secret_basic or client_secret_post
func acquireAccessToken(metadata *map[string]interface{}, clientID string, clientSecret string, scopes []string) (*oauth2.Token, error) {

	// Determine the endpoint to use
	tokenEndpoint, err := shouldBeURLOrNil(metadata, "token_endpoint")
	if nil != err {
		return nil, err
	}

	// Configure the Client Credentials Implementation
	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenEndpoint.String(),
		Scopes:       scopes,
	}

	// Request Token
	token, err := config.Token(context.TODO())
	if nil != err {
		return nil, err
	}

	return token, nil
}
