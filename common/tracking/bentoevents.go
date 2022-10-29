package tracking

import (
	"context"
	"time"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/api-server/version"
)

func TrackBentoEvent(bentoschema schemasv1.BentoSchema, eventType BentoEventType) {
	bentoEvent := BentoEvent{
		TriggerEvent: TriggerEvent{
			UserUID: bentoschema.Creator.Uid,
		},
		CommonProperties: CommonProperties{
			YataiVersion:    version.Version,
			Timestamp:       time.Now(),
            //TODO get org ID
			OrganizationUID: "",
		},
		BentoEventType:       eventType,
		BentoRepositoryUID:   bentoschema.BentoRepositoryUid,
		BentoVersion:         bentoschema.Version,
		UploadStatus:         bentoschema.UploadStatus,
		UploadFinishedReason: bentoschema.UploadFinishedReason,
	}

	if bentoschema.Manifest != nil {
		bentoEvent.NumModels = len(bentoschema.Manifest.Models)
		bentoEvent.NumRunners = len(bentoschema.Manifest.Runners)
		bentoEvent.BentoSizeBytes = bentoschema.Manifest.SizeBytes
	}
	track(bentoEvent, "bentoEvent")
}

func TrackBentoEventModel(ctx context.Context, bentoModel *models.Bento, eventType BentoEventType) {
	b, _ := transformersv1.ToBentoSchema(ctx, bentoModel)
	TrackBentoEvent(*b, eventType)
}
