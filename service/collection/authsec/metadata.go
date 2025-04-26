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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/eclipse-xfsc/cam/api/common"
)

// getMetadata tries to automatically find the server metadata document for a given issuer identifier.
// It prioritizes the URLs given in RFC 8414, but uses the wide-spread OpenID metadata document as a fallback.
// The actual HTTP call can be found in readMetadata below.
func getMetadata(issuer *url.URL, url *url.URL) (*map[string]interface{}, *common.Error) {
	logrus.Traceln("Loading RFC 8414 Metadata document for: " + issuer.String())
	var err error

	// Use the given URL, if applicable
	if nil != url {
		return readMetadata(issuer, url)
	}

	logrus.Traceln("Automatically deriving Metadata URL")

	logrus.Traceln("Trying RFC 8414 Metadata document location")
	url, err = url.Parse(issuer.Scheme + "://" + issuer.Host + "/.well-known/oauth-authorization-server" + issuer.Path)
	if nil != err {
		return nil, &common.Error{
			Code:        common.Error_ERROR_UNKNOWN, // TODO: Should be "internal error"
			Description: "Error forming metadata URL: " + err.Error(),
		}
	}
	metadata, errStruct := readMetadata(issuer, url)
	if nil == errStruct || common.Error_ERROR_CONNECTION_FAILURE == errStruct.Code {
		return metadata, errStruct
	}

	logrus.Traceln("Trying RFC 8414 Fallback Metadata document location")
	url, err = url.Parse(issuer.Scheme + "://" + issuer.Host + "/.well-known/openid-configuration" + issuer.Path)
	if nil != err {
		return nil, &common.Error{
			Code:        common.Error_ERROR_UNKNOWN, // TODO: Should be "internal error"
			Description: "Error forming metadata URL: " + err.Error(),
		}
	}
	metadata, errStruct = readMetadata(issuer, url)
	if nil == errStruct || common.Error_ERROR_CONNECTION_FAILURE == errStruct.Code {
		return metadata, errStruct
	}

	logrus.Traceln("Trying legacy OpenID location")
	url, err = url.Parse(issuer.Scheme + "://" + issuer.Host + issuer.Path + "/.well-known/openid-configuration")
	if nil != err {
		return nil, &common.Error{
			Code:        common.Error_ERROR_UNKNOWN, // TODO: Should be "internal error"
			Description: "Error forming metadata URL: " + err.Error(),
		}
	}
	return readMetadata(issuer, url)
}

// readMetadata fetches an OAuth Server Metadata document from a given URL
// The response *should* be a JSON object with the "issuer" key matching the issuer we are looking for
func readMetadata(issuer *url.URL, url *url.URL) (*map[string]interface{}, *common.Error) {
	response, err := http.Get((*url).String())
	if nil != err {
		logrus.Warnln(err)
		return nil, &common.Error{
			Code:        common.Error_ERROR_CONNECTION_FAILURE,
			Description: "Error fetching Metadata: " + err.Error(),
		}
	}
	response_code := response.StatusCode
	if response_code < 200 || response_code >= 300 {
		return nil, &common.Error{
			Code:        common.Error_ERROR_PROTOCOL_VIOLATION,
			Description: "Request for Metadata document returned status code " + strconv.Itoa(response.StatusCode),
		}
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if nil != err {
		return nil, &common.Error{
			Code:        common.Error_ERROR_PROTOCOL_VIOLATION,
			Description: "Request for Metadata document returned invalid body",
		}
	}

	var metadata map[string]interface{}
	err = json.Unmarshal([]byte(body), &metadata)
	if nil != err {
		return nil, &common.Error{
			Code:        common.Error_ERROR_PROTOCOL_VIOLATION,
			Description: "Metadata document is not JSON",
		}
	}
	if nil == metadata["issuer"] {
		return nil, &common.Error{
			Code:        common.Error_ERROR_PROTOCOL_VIOLATION,
			Description: "Metadata is missing Issuer",
		}
	}
	issuerString, ok := metadata["issuer"].(string)
	if !ok {
		return nil, &common.Error{
			Code:        common.Error_ERROR_PROTOCOL_VIOLATION,
			Description: "Metadata Issuer is not a string",
		}
	}
	if issuerString != issuer.String() {
		return nil, &common.Error{
			Code:        common.Error_ERROR_PROTOCOL_VIOLATION,
			Description: "Metadata has incorrect Issuer: " + issuerString,
		}
	}

	return &metadata, nil
}

// checkMetadataRFC8414 sanity-checks a metadata document
// and fills in additional metadata default values according to RFC 8414
// This function helps with input validation of fetched documents.
func checkMetadataRFC8414(metadata *map[string]interface{}) (*map[string]interface{}, error) {
	var err error
	var val []string
	var url *url.URL

	// Grant Types Supported
	setDefaultValue(metadata, "grant_types_supported", []string{"authorization_code", "implicit"})
	grant_types_supported, err := shouldBeStringArrayOrNil(metadata, "grant_types_supported")
	if nil != err {
		return nil, err
	}

	// Authorization Endpoint
	url, err = shouldBeURLOrNil(metadata, "authorization_endpoint")
	if nil != err {
		return nil, err
	}
	if nil == url {
		if stringHas(grant_types_supported, "authorization_code") || stringHas(grant_types_supported, "implicit") {
			return nil, errors.New("metadata is missing authorization_endpoint")
		}
	}

	// Token Endpoint
	url, err = shouldBeURLOrNil(metadata, "token_endpoint")
	if nil != err {
		return nil, err
	}
	if nil == url {
		for i := 0; i < len(grant_types_supported); i++ {
			if grant_types_supported[i] != "implicit" {
				return nil, errors.New("metadata is missing token_endpoint")
			}
		}
	}

	// Response Types Supported
	val, err = shouldBeStringArrayOrNil(metadata, "response_types_supported")
	if nil != err {
		return nil, err
	}
	if nil == val {
		return nil, errors.New("metadata is missing response_types_supported")
	}

	// Response Modes Supported
	setDefaultValue(metadata, "response_modes_supported", []string{"query", "fragment"})

	// Client Auth Secured endpoints
	for _, ep := range []string{"token", "revocation", "introspection"} {
		// X Endpoint Auth Methods Supported
		setDefaultValue(metadata, ep+"_endpoint_auth_methods_supported", []string{"client_secret_basic"})
		endpoint_auth_methods_supported, err := shouldBeStringArrayOrNil(metadata, ep+"_endpoint_auth_methods_supported")
		if nil != err {
			return nil, err
		}

		// X Endpoint Auth Signing Alg Values Supported
		val, err = shouldBeStringArrayOrNil(metadata, ep+"_endpoint_auth_signing_alg_values_supported")
		if nil != err {
			return nil, err
		}
		if nil == val {
			for i := 0; i < len(endpoint_auth_methods_supported); i++ {
				if stringHas(endpoint_auth_methods_supported, "client_secret_jwt") || stringHas(endpoint_auth_methods_supported, "private_key_jwt") {
					return nil, errors.New("Metadata is missing " + ep + "_endpoint_auth_signing_alg_values_supported")
				}
			}
		} else if stringHas(val, "none") {
			return nil, errors.New("'none' value included in " + ep + "_endpoint_auth_signing_alg_values_supported")
		}
	}

	// other optional URLs (type checks only)
	for _, key := range []string{"jwks_uri", "service_documentation", "op_policy_url", "op_tos_url", "registration_endpoint", "revocation_endpoint", "introspection_endpoint"} {
		_, err = shouldBeURLOrNil(metadata, key)
		if nil != err {
			return nil, err
		}
	}

	// other optional String Arrays (type checks only)
	for _, key := range []string{"scopes_supported", "response_modes_supported", "ui_locales_supported", "code_challenge_methods_supported"} {
		_, err = shouldBeStringArrayOrNil(metadata, key)
		if nil != err {
			return nil, err
		}
	}

	// Return modified metadata
	return metadata, nil
}

// checkMetadataOIDCDiscovery sanity-checks an OpenID Connect Provider metadata document
// and fills in additional metadata default values according to OpenID Connect Discovery.
// This function helps with input validation of fetched documents.
// Assumes that checkMetadataRFC8414 was successful.
func checkMetadataOIDCDiscovery(metadata *map[string]interface{}) (*map[string]interface{}, error) {
	var err error
	var val []string
	//var url *url.URL

	// Subject Types Supported
	val, err = shouldBeStringArrayOrNil(metadata, "subject_types_supported")
	if nil != err {
		return nil, err
	}
	if nil == val {
		return nil, errors.New("metadata is missing subject_types_supported")
	}

	// {ID Token, Userinfo, Request Object} {Signing Alg, Encryption Alg, Encryption Enc} Values Supported
	for _, target := range []string{"id_token", "userinfo", "request_object"} {
		for _, claim := range []string{"signing_alg", "encryption_alg", "encryption_enc"} {
			val, err = shouldBeStringArrayOrNil(metadata, target+"_"+claim+"_values_supported")
			if nil != err {
				return nil, err
			}
		}
	}
	val, _ = shouldBeStringArrayOrNil(metadata, "id_token_signing_alg_values_supported")
	if nil == val {
		return nil, errors.New("metadata is missing id_token_signing_alg_values_supported")
	}

	// other optional String Arrays (type checks only)
	for _, key := range []string{"acr_values_supported", "display_values_supported", "claim_types_supported", "claims_supported", "claims_locales_supported", "ui_locales_supported"} {
		_, err = shouldBeStringArrayOrNil(metadata, key)
		if nil != err {
			return nil, err
		}
	}

	// Boolean flags
	setDefaultValue(metadata, "claims_parameter_supported", false)
	setDefaultValue(metadata, "request_parameter_supported", false)
	setDefaultValue(metadata, "request_uri_parameter_supported", true)
	setDefaultValue(metadata, "require_request_uri_registration", false)
	for _, key := range []string{"claims_parameter_supported", "request_parameter_supported", "request_uri_parameter_supported", "require_request_uri_registration"} {
		_, err = shouldBeBoolOrNil(metadata, key)
		if nil != err {
			return nil, err
		}
	}

	// return modified Metadata
	return metadata, nil
}

// shouldBeStringArrayOrNil extracts an array of strings from a JSON object.
// Returns nil if the given key was non-existent or its value was nil.
// This function is heavily inspired by https://github.com/gookit/goutil/blob/master/arrutil/convert.go
func shouldBeStringArrayOrNil(metadata *map[string]interface{}, key string) ([]string, error) {
	val := (*metadata)[key]
	if nil != val {
		var ret []string
		refl := reflect.ValueOf(val)
		if refl.Kind() != reflect.Slice {
			return nil, errors.New(key + " should be an Array of Strings")
		}

		for i := 0; i < refl.Len(); i++ {
			s, ok := refl.Index(i).Interface().(string)
			if !ok {
				return nil, errors.New(key + " should be an Array of Strings")
			}

			ret = append(ret, s)
		}
		return ret, nil
	}
	return nil, nil
}

// shouldBeURLOrNil extracts a URL from a JSON object
// Returns nil if the given key was non-existent or its value was nil.
func shouldBeURLOrNil(metadata *map[string]interface{}, key string) (*url.URL, error) {
	if nil != (*metadata)[key] {
		val, ok := (*metadata)[key].(string)
		if !ok {
			return nil, errors.New(key + " should be a URL")
		}
		res, err := url.Parse(val)
		if nil != err {
			return nil, errors.New(key + " should be a URL")
		}
		return res, nil
	}
	return nil, nil
}

// shouldBeBoolOrNil extracts a boolean value from a JSON object
// Returns nil if the given key was non-existent or its value was nil.
func shouldBeBoolOrNil(metadata *map[string]interface{}, key string) (bool, error) {
	if nil != (*metadata)[key] {
		val, ok := (*metadata)[key].(bool)
		if !ok {
			return false, errors.New(key + " should be a Boolean Value")
		}
		return val, nil
	}
	return false, nil
}

// setDefaultValue can be used to provide a default value to a key in a JSON object
func setDefaultValue(metadata *map[string]interface{}, key string, value interface{}) {
	if nil == (*metadata)[key] {
		(*metadata)[key] = value
	}
}

// stringHas determines if a array of strings contains a particular string
func stringHas(arr []string, searchterm string) bool {
	for _, s := range arr {
		if searchterm == s {
			return true
		}
	}
	return false
}
