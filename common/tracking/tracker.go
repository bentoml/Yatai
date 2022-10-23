package tracking

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/version"
)

const TRACKING_URL = "http://localhost:8000"

func TrackDeploymentSchema(deploymentSchema *schemasv1.DeploymentSchema) {
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

	deploymentSchemaParsed := DeploymentEvent{
		TriggerEvent: TriggerEvent{
			UserUID: deploymentSchema.Creator.Uid,
		},
		CommonProperties: CommonProperties{
			YataiVersion:    version.Version,
			Timestamp:       time.Now(),
			OrganisationUID: deploymentSchema.Cluster.Organization.Uid,
		},
		ClusterUID:            deploymentSchema.Cluster.Uid,
		DeploymentUID:         deploymentSchema.Uid,
		DeploymentEventType:   DeploymentEventTypeCreate,
		DeploymentStatus:      deploymentSchema.Status,
		DeploymentRevisionID:  deploymentSchema.LatestRevision.Uid,
		DeploymentTargetTypes: deploymentTargetTypes,
		ApiServerResources:    apiResources,
		ApiServerHPAConfig:    apiHPAConfs,
		RunnerResourcesList:   runnerResourcesList,
		RunnerHPAConfigList:   runnerHPAConfigList,
	}

	data, err := json.Marshal(deploymentSchemaParsed)
	if err != nil {
		panic(err)
	}
	track(data)
}

// Sent the marshalled data to tracking server
func track(data []byte) {
	const (
		JITSU_SERVER = "http://localhost:8000/api/v1/s2s/event"
		JITSUE_KEY   = "s2s.l5z3kpmcuevukrepeg70ht.j2imnduhnssuh4yb7gt9"
	)
	request_url := fmt.Sprintf("%s?token=%s", JITSU_SERVER, JITSUE_KEY)
	resp, err := http.Post(request_url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Print(err)
		panic("this is bad!")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("success")
	} else {
		log.Println("fail")
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		log.Println(bodyString)
	}
}
