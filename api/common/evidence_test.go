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

package common_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/eclipse-xfsc/cam/api/common"
	"github.com/eclipse-xfsc/cam/internal/testutil"
	"github.com/eclipse-xfsc/cam/internal/testutil/testproto"
)

const MockEvidenceID = "00000000-0000-0000-0000-000000000000"

func Test_PersistEvidence(t *testing.T) {
	var (
		err error
		e2  common.Evidence
	)

	storage := testutil.NewInMemoryStorage(t)

	e := common.Evidence{
		Id:    MockEvidenceID,
		Name:  "my evidence",
		Value: testproto.ToValue(t, struct{ Test string }{Test: "test"}),
	}

	err = storage.Save(&e)
	assert.NoError(t, err)

	err = storage.Get(&e2, "id = ?", MockEvidenceID)
	assert.NoError(t, err)

	assert.True(t, proto.Equal(&e, &e2))
}
