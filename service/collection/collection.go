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

package collection

import (
	"clouditor.io/clouditor/api"
	"clouditor.io/clouditor/voc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/eclipse-xfsc/cam/api/collection"
	"github.com/eclipse-xfsc/cam/api/common"
	"github.com/eclipse-xfsc/cam/api/evaluation"
)

const (
	TargetComponent = "Evaluation Manager"
)

// EnqueueEvidences creates evidences and sends them into the stream to the Evaluation Manager
func EnqueueEvidences(toolId string, req *collection.StartCollectingRequest, results []voc.IsCloudResource, stream *api.StreamChannelOf[evaluation.Evaluation_SendEvidencesClient, *common.Evidence], log *logrus.Entry) error {
	var (
		evidence *common.Evidence
	)

	for _, result := range results {
		value, err := voc.ToStruct(result)
		if err != nil {
			log.Errorf("Error transforming resource to a struct: %v", err)
			return err
		}

		// Create UUID for evidence ID and name
		id := uuid.NewString()

		evidence = &common.Evidence{
			Id:             id,
			Name:           id,
			TargetService:  req.ServiceId,
			TargetResource: string(result.GetID()),
			ToolId:         toolId,
			GatheredAt:     timestamppb.New(*result.GetCreationTime()),
			RawEvidence:    "", // Optional, could be a JSON representation of the raw underlying evidence
			Value:          value,
		}

		log.Debugf("Evidence: %v", evidence)

		// Send evidence to stream
		types, err := evidence.ResourceTypes()
		if err != nil {
			log.Debugf("could not extract resource types from evidence %s: %v", evidence.Id, err)
		}

		log.Infof("Sending evidence '%s' with resource type %s to evaluation manager stream", evidence.Id, types)
		stream.Send(evidence)
	}

	return nil
}
