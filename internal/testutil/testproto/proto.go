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

package testproto

import (
	"testing"

	"github.com/eclipse-xfsc/cam/internal/protobuf"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// NewAny creates a new anypb.Any out of a protobuf message and panics on error.
func NewAny(t *testing.T, m protoreflect.ProtoMessage) (a *anypb.Any) {
	var err error

	a, err = anypb.New(m)
	if err != nil {
		t.Error(err)
	}

	return
}

// ToValue is a wrapper around [protobuf.ToValue] which fails the test on error.
func ToValue(t *testing.T, strct any) (value *structpb.Value) {
	var err error

	value, err = protobuf.ToValue(strct)
	if err != nil {
		t.Error(err)
	}

	return
}
