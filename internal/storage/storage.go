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

package storage

import (
	"clouditor.io/clouditor/persistence"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VerifyExistence verifies that at least one entry exists for the field given by column and value. Returns gRPC error
// if not.
func VerifyExistence(storage persistence.Storage, entryType string, table any, column, value string) (err error) {
	count, err := storage.Count(table, "%s = ?", column, value)
	if err != nil {
		err = status.Errorf(codes.Internal, "DB error: %v", err)
		return
	}
	if count == 0 {
		err = status.Errorf(codes.NotFound, "No %s found for '%s' (%s)", entryType, value, column)
		return
	}
	return
}
