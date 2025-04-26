# Evidences
The collection modules create evidences based on the Evidence structure in api/common/evidence.proto. An example evidence is shown in the following:

```go
id := uuid.NewString()

common.Evidence{
	Id:             id, // A new uuid
	Name:           id, // For now the name is equal to the ID
	TargetService:  req.ServiceId, // The ServiceId from the collection.StartColletingRequest (api/collection/collection.proto)
	TargetResource: // Optional. A resource ID within the service, e.g., Resource.ID from clouditor.io/clouditor/voc/voc.go
	GatheredUsing:  req.MetricId, // The MetricId from the collection.StartColletingRequest (api/collection/collection.proto)
	GatheredAt:     // Timestamp of evidence gathering should come from the evidence collection tool, e.g., Clouditor Discovery component
	RawEvidence:    // Optional, could be a JSON representation of the raw underlying evidence in its original form
	Value:          // The measured value, e.g., the resource from clouditor.io/clouditor/voc/voc.go
}
```
