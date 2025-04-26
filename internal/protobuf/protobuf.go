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

package protobuf

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func ToValue(strct any) (value *structpb.Value, err error) {
	var b []byte

	value = new(structpb.Value)

	if b, err = json.Marshal(strct); err != nil {
		err = fmt.Errorf("JSON marshal failed: %w", err)
		return
	}
	if err = json.Unmarshal(b, &value); err != nil {
		err = fmt.Errorf("JSON unmarshal failed: %w", err)
		return
	}
	return
}

func ToStruct[T any](value *structpb.Value) (strct T, err error) {
	var b []byte

	if value == nil {
		err = fmt.Errorf("empty value")
		return
	}

	if b, err = json.Marshal(value); err != nil {
		err = fmt.Errorf("JSON marshal failed: %w", err)
		return
	}
	if err = json.Unmarshal(b, &strct); err != nil {
		err = fmt.Errorf("JSON unmarshal failed: %w", err)
		return
	}
	return
}

// ToByteArray converts the protobuf value to a byte array
func ToByteArray(v *structpb.Value) (b []byte, err error) {
	value := v.AsInterface()

	if value == nil {
		err = fmt.Errorf("converting raw configuration to map is empty or nil")
		return
	}

	switch v := value.(type) {
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("got type %T but wanted string", v)
	}
}

// TypeURL returns the protobuf type URL of the message. According to the protobuf specification, this always begins
// with type.googleapis.com.
func TypeURL(m proto.Message) string {
	return fmt.Sprintf("type.googleapis.com/%s", m.ProtoReflect().Descriptor().FullName())
}
