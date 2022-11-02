package tracking

import (
	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/version"
)

// track a DeploymentEvent(create/update/terminate/delete)
func TrackDeploymentEvent(deploymentSchema *schemasv1.DeploymentSchema, eventType YataiEventType) {
	deploymentSchemaParsed := DeploymentEvent{
		UserUID:          deploymentSchema.Creator.Uid,
		CommonProperties: NewCommonProperties(eventType, deploymentSchema.Cluster.Organization.Uid, version.Version),
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
			deploymentTargetTypes = append(deploymentTargetTypes, deploymentTarget.DeploymentTargetTypeSchema.Type)
			apiResources = append(apiResources, *deploymentTarget.Config.Resources)
			apiHPAConfs = append(apiHPAConfs, *deploymentTarget.Config.HPAConf)

			runnerResourcesList[i] = make(map[string]modelschemas.DeploymentTargetResources)
			runnerHPAConfigList[i] = make(map[string]modelschemas.DeploymentTargetHPAConf)
			for runnerName, runnerConfig := range deploymentTarget.Config.Runners {
				runnerResourcesList[i][runnerName] = *runnerConfig.Resources
				runnerHPAConfigList[i][runnerName] = *runnerConfig.HPAConf
			}
		}
		deploymentSchemaParsed.DeploymentTargetTypes = deploymentTargetTypes
		deploymentSchemaParsed.ApiServerResources = apiResources
		deploymentSchemaParsed.ApiServerHPAConfig = apiHPAConfs
		deploymentSchemaParsed.RunnerResourcesList = runnerResourcesList
		deploymentSchemaParsed.RunnerHPAConfigList = runnerHPAConfigList
		deploymentSchemaParsed.DeploymentRevisionID = deploymentSchema.LatestRevision.Uid
	}

	track(deploymentSchemaParsed, eventType)
}
