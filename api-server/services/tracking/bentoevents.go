package tracking

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/api-server/version"
)

func TrackBentoEvent(ctx context.Context, bentoModel *models.Bento, eventType YataiEventType) {
	if bentoModel == nil {
		return
	}
	trackingLogger := NewTrackerLogger().WithField("eventType", eventType)
	org, err := services.GetCurrentOrganization(ctx)
	if err != nil {
		trackingLogger.Error("Error getting orgID, using empty OrgUID")
		return
	}

	defaultOrg, err := services.OrganizationService.GetDefault(ctx)
	if err != nil {
		trackingLogger.Error(err)
		return
	}
	bentoschema, _ := transformersv1.ToBentoSchema(ctx, bentoModel)
	bentoEvent := BentoEvent{
		UserUID: bentoschema.Creator.Uid,
		CommonProperties: NewCommonProperties(
			eventType, org.Uid, defaultOrg.Uid, version.Version),
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
	track(ctx, bentoEvent, eventType)
}
