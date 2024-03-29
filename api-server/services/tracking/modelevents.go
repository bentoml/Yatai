package tracking

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/api-server/version"
)

func TrackModelEvent(ctx context.Context, modelModel *models.Model, eventType YataiEventType) {
	if modelModel == nil {
		return
	}
	trackingLogger := NewTrackerLogger().WithField("eventType", eventType)
	modelschema, err := transformersv1.ToModelSchema(ctx, modelModel)
	if err != nil {
		trackingLogger.Error("Error transforming modelschema: ", err)
		return
	}

	modelRepository, err := services.ModelRepositoryService.GetAssociatedModelRepository(ctx, modelModel)
	if err != nil {
		trackingLogger.Error(err)
		return
	}
	org, err := services.OrganizationService.GetAssociatedOrganization(ctx, modelRepository)
	if err != nil {
		trackingLogger.Error("Unnable to get associated org: ", err)
	}
	instanceOrg, err := services.OrganizationService.GetDefault(ctx)
	if err != nil {
		trackingLogger.Error("coun't get default Org: ", err)
	}

	modelEvent := ModelEvent{
		UserUID: modelschema.Creator.Uid,
		CommonProperties: NewCommonProperties(
			eventType, instanceOrg.Uid, org.Uid, version.Version),
		ModelUID:                  modelschema.ModelUid,
		ModelUploadStatus:         modelschema.UploadStatus,
		ModelUploadFinishedReason: modelschema.UploadFinishedReason,
		ModelTransmissionStrategy: modelschema.TransmissionStrategy,
	}

	if modelschema.Manifest != nil {
		modelEvent.ModelSizeBytes = modelschema.Manifest.SizeBytes
	}
	track(ctx, modelEvent, eventType)
}
