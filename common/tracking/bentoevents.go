package tracking

import (
	"context"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/api-server/version"
)

func TrackBentoEvent(bentoschema schemasv1.BentoSchema, eventType BentoEventType) {
	bentoEvent := BentoEvent{
		UserUID: bentoschema.Creator.Uid,
		// TODO:fix organisation ID
		CommonProperties:          NewCommonProperties(YataiBentoEvent, "", version.Version),
		BentoEventType:            eventType,
		BentoRepositoryUID:        bentoschema.BentoRepositoryUid,
		BentoVersion:              bentoschema.Version,
		BentoUploadStatus:         bentoschema.UploadStatus,
		BentoUploadFinishedReason: bentoschema.UploadFinishedReason,
		BentoTransmissionStrategy: bentoschema.TransmissionStrategy,
	}

	if bentoschema.Manifest != nil {
		bentoEvent.NumModels = len(bentoschema.Manifest.Models)
		bentoEvent.NumRunners = len(bentoschema.Manifest.Runners)
		bentoEvent.BentoSizeBytes = bentoschema.Manifest.SizeBytes
	}
	track(bentoEvent, string(YataiBentoEvent))
}

func TrackBentoEventModel(ctx context.Context, bentoModel *models.Bento, eventType BentoEventType) {
	b, _ := transformersv1.ToBentoSchema(ctx, bentoModel)
	TrackBentoEvent(*b, eventType)
}
