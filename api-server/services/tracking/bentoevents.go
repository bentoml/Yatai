package tracking

import (
	"context"

	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/api-server/version"
	log "github.com/sirupsen/logrus"
)

func TrackBentoEvent(bentoschema schemasv1.BentoSchema, eventType YataiEventType) {
	bentoEvent := BentoEvent{
		UserUID: bentoschema.Creator.Uid,
		// TODO:fix organisation ID
		CommonProperties:          NewCommonProperties(eventType, "", version.Version),
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
	track(bentoEvent, eventType)
}

func TrackBentoEventModel(ctx context.Context, bentoModel *models.Bento, eventType YataiEventType) {
	org, err := services.GetCurrentOrganization(ctx)
	var orgUID string
	if err != nil {
		log.Error("Error getting orgID, using empty OrgUID")
	} else {
		orgUID = org.Uid
		log.Info("Got OrgID: ", orgUID)
	}
	b, _ := transformersv1.ToBentoSchema(ctx, bentoModel)
	TrackBentoEvent(*b, eventType)
}
