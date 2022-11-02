package tracking

import (
	"context"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/api-server/version"
)

func TrackModelEvent(modelschema schemasv1.ModelSchema, eventType ModelEventType) {
	modelEvent := ModelEvent{
		UserUID: modelschema.Creator.Uid,
		//TODO get org ID
		CommonProperties:          NewCommonProperties(YataiModelEvent, "", version.Version),
		ModelEventType:            eventType,
		ModelUID:                  modelschema.ModelUid,
		ModelUploadStatus:         modelschema.UploadStatus,
		ModelUploadFinishedReason: modelschema.UploadFinishedReason,
		ModelTransmissionStrategy: modelschema.TransmissionStrategy,
	}

	if modelschema.Manifest != nil {
		modelEvent.ModelSizeBytes = modelschema.Manifest.SizeBytes
	}
	track(modelEvent, string(YataiModelEvent))
}

func TrackModelEventModel(ctx context.Context, modelModel *models.Model, eventType ModelEventType) {
	b, _ := transformersv1.ToModelSchema(ctx, modelModel)
	TrackModelEvent(*b, eventType)
}
