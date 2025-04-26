#!/usr/bin/env bash
NAMESPACE="$1"
access_token=`curl -SsX POST --user gaiax-fs-req:$CAM_REQ_MANAGER_OAUTH2_CLIENT_SECRET --data "grant_type=client_credentials" --data "scope=gaiax" $(curl -sS https://as.aisec.fraunhofer.de/.well-known/oauth-authorization-server/auth | jq -r .token_endpoint) | jq -r .access_token`

curl -sL -H "Authorization: Bearer $access_token" --insecure cam-api-gateway.${NAMESPACE}:8080/v1/configuration/cloud_services | jq
curl -sL -H "Authorization: Bearer $access_token" --insecure cam-api-gateway.${NAMESPACE}:8080/v1/configuration/controls | jq
curl -sL -H "Authorization: Bearer $access_token" --insecure cam-api-gateway.${NAMESPACE}:8080/v1/configuration/metrics | jq
curl -sL -H "Authorization: Bearer $access_token" --insecure cam-api-gateway.${NAMESPACE}:8080/v1/evaluation/evidences | jq
