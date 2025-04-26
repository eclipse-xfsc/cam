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

package integrity

import (
	"encoding/json"
	"fmt"

	"clouditor.io/clouditor/voc"
	"google.golang.org/protobuf/types/known/structpb"
)

type Value struct {
	voc.Resource

	SystemComponentsIntegrity `json:"systemComponentsIntegrity"`
}

type SystemComponentsIntegrity struct {
	Status bool `json:"status"`
}

func toStruct(v Value) (s *structpb.Value, err error) {
	var b []byte

	s = new(structpb.Value)

	if b, err = json.Marshal(v); err != nil {
		err = fmt.Errorf("JSON marshal failed: %w", err)
		return
	}
	if err = json.Unmarshal(b, &s); err != nil {
		err = fmt.Errorf("JSON unmarshal failed: %w", err)
		return
	}
	return
}

/*
func toStruct(value *structpb.Value) (conf *collection.IntegrityConfig, err error) {
	strct, err := testutil.ToStruct(value)
	if err != nil {
		return
	}

	var ok bool
	conf, ok = strct.(*collection.IntegrityConfig)
	if !ok {
		err = errors.New("")
		return
	}
	return
}

*/
