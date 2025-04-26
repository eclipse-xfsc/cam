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

package common

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	ErrEvidenceMissing         = errors.New("evidence is missing")
	ErrEvidenceIdInvalidFormat = errors.New("evidence id not in expected format (UUID) or missing")
	ErrTimestampMissing        = errors.New("evidence timestamp is missing")
	ErrTargetServiceMissing    = errors.New("evidence target service is missing")
	ErrEvidenceWithError       = errors.New("evidence includes error")
	ErrValueMissing            = errors.New("evidence value is missing")
	ErrValueNotStruct          = errors.New("evidence value is no struct value")
	ErrValueNotMap             = errors.New("evidence (struct) value is not convertible to map")
)

func (e *Evidence) Validate() error {
	if e == nil {
		return ErrEvidenceMissing
	}
	if _, err := uuid.Parse(e.Id); err != nil {
		return ErrEvidenceIdInvalidFormat
	}
	if e.GatheredAt == nil {
		return ErrTimestampMissing
	}
	if e.TargetService == "" {
		return ErrTargetServiceMissing
	}
	if e.Error != nil {
		return ErrEvidenceWithError
	}
	if e.Value == nil {
		return ErrValueMissing
	}
	if structValue := e.Value.GetStructValue(); structValue == nil {
		return ErrValueNotStruct
	}
	if m := e.Value.GetStructValue().AsMap(); m == nil {
		return ErrValueNotMap
	}
	return nil
}

// ResourceTypes parses the embedded resource of this evidence and returns its types according to the ontology.
func (evidence *Evidence) ResourceTypes() (types []string, err error) {
	var (
		m     map[string]interface{}
		value *structpb.Value
	)

	value = evidence.Value
	if value == nil {
		return
	}

	m = value.GetStructValue().AsMap()

	if rawTypes, ok := m["type"].([]interface{}); ok {
		if len(rawTypes) != 0 {
			types = make([]string, len(rawTypes))
		} else {
			return nil, fmt.Errorf("list of types is empty")
		}
	} else {
		return nil, fmt.Errorf("got type '%T' but wanted '[]interface {}'. Check if resource types are specified ", rawTypes)
	}
	for i, v := range m["type"].([]interface{}) {
		if t, ok := v.(string); !ok {
			return nil, fmt.Errorf("got type '%T' but wanted 'string'", t)
		} else {
			types[i] = t
		}
	}

	return
}
