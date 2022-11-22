package tracking

import (
	"context"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/version"
)

// track a DeploymentEvent(create/update/terminate/delete)
func TrackDeploymentEvent(ctx context.Context, deploymentSchema *schemasv1.DeploymentSchema, eventType YataiEventType) {
	trackingLogger := NewTrackerLogger().WithField("eventType", eventType)
	defaultOrg, err := services.OrganizationService.GetDefault(ctx)
	if err != nil {
		trackingLogger.Error(err)
	}
	deploymentSchemaParsed := DeploymentEvent{
		UserUID: deploymentSchema.Creator.Uid,
		CommonProperties: NewCommonProperties(
			eventType, defaultOrg.Uid, deploymentSchema.Cluster.Organization.Uid, version.Version),
		ClusterUID:       deploymentSchema.Cluster.Uid,
		DeploymentUID:    deploymentSchema.Uid,
		DeploymentStatus: deploymentSchema.Status,
	}

	// ignore DeploymentTarget information if *LatestRevision is nil
	if deploymentSchema.LatestRevision != nil {
		var deploymentTargetTypes []modelschemas.DeploymentTargetType
		var apiResources []modelschemas.DeploymentTargetResources
		var apiHPAConfs []modelschemas.DeploymentTargetHPAConf
		var runnerResourcesList = make([]map[string]modelschemas.DeploymentTargetResources, len(deploymentSchema.LatestRevision.Targets))
		var runnerHPAConfigList = make([]map[string]modelschemas.DeploymentTargetHPAConf, len(deploymentSchema.LatestRevision.Targets))

		for i, deploymentTarget := range deploymentSchema.LatestRevision.Targets {
			if deploymentTarget.Config == nil {
				continue
			}

			deploymentTargetTypes = append(deploymentTargetTypes, deploymentTarget.DeploymentTargetTypeSchema.Type)
			if deploymentTarget.Config.Resources != nil {
				apiResources = append(apiResources, *deploymentTarget.Config.Resources)
			}
			if deploymentTarget.Config.HPAConf != nil {
				apiHPAConfs = append(apiHPAConfs, *deploymentTarget.Config.HPAConf)
			}

			runnerResourcesList[i] = make(map[string]modelschemas.DeploymentTargetResources)
			runnerHPAConfigList[i] = make(map[string]modelschemas.DeploymentTargetHPAConf)
			for runnerName, runnerConfig := range deploymentTarget.Config.Runners {
				if runnerConfig.Resources != nil {
					runnerResourcesList[i][runnerName] = *runnerConfig.Resources
				}
				if runnerConfig.HPAConf != nil {
					runnerHPAConfigList[i][runnerName] = *runnerConfig.HPAConf
				}
			}
		}
		deploymentSchemaParsed.DeploymentTargetTypes = deploymentTargetTypes
		deploymentSchemaParsed.ApiServerResources = apiResources
		deploymentSchemaParsed.ApiServerHPAConfig = apiHPAConfs
		deploymentSchemaParsed.RunnerResourcesList = runnerResourcesList
		deploymentSchemaParsed.RunnerHPAConfigList = runnerHPAConfigList
		deploymentSchemaParsed.DeploymentRevisionID = deploymentSchema.LatestRevision.Uid
	}

	track(ctx, deploymentSchemaParsed, eventType)
}
