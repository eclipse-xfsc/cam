## Licensed under the Apache License, Version 2.0 (the "License");
## you may not use this file except in compliance with the License.
## You may obtain a copy of the License at
##
##	http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software
## distributed under the License is distributed on an "AS IS" BASIS,
## WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and
## limitations under the License.
##
## Contributors:
##	Fraunhofer AISEC

cli=\
cam

services=\
cam-api-gateway\
cam-req-manager\
cam-eval-manager\
cam-collection-authsec\
cam-collection-integrity\
cam-collection-workload

all: $(services) $(cli)

.PHONY: clean

clean:
	rm -rf build

# Build services only for Linux
$(services):
	echo "Building service: $@"
	GOARCH=arm64 GOOS=linux go build -o build/$@-linux-arm64 cmd/$@/$@.go
	GOARCH=amd64 GOOS=linux go build -o build/$@-linux-amd64 cmd/$@/$@.go

# Build CLI tools for mac and linux
$(cli):
	echo "Building CLI tool: $@"
	GOARCH=arm64 GOOS=darwin go build -o build/$@-darwin-arm64 cmd/$@/$@.go
	GOARCH=amd64 GOOS=darwin go build -o build/$@-darwin-amd64 cmd/$@/$@.go
	GOARCH=arm64 GOOS=linux go build -o build/$@-linux-amd64 cmd/$@/$@.go
	GOARCH=amd64 GOOS=linux go build -o build/$@-linux-amd64 cmd/$@/$@.go
