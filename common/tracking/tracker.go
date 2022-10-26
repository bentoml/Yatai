package tracking

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/version"
	log "github.com/sirupsen/logrus"
)

const (
	JITSU_SERVER              = "http://localhost:8000/api/v1/s2s/event"
	JITSUE_KEY                = "s2s.l5z3kpmcuevukrepeg70ht.j2imnduhnssuh4yb7gt9"
	YATAI_TRACKING_DEBUG_MODE = "__YATAI_TRACKING_DEBUG_MODE"
	YATAI_DONOT_TRACK         = "YATAI_DONOT_TRACK"
)

func is_debug_mode() bool {
	out := os.Getenv(YATAI_TRACKING_DEBUG_MODE)
	if strings.ToLower(out) == "true" {
		return true
	} else {
		return false
	}
}

func donot_track() bool {
	out := os.Getenv(YATAI_DONOT_TRACK)
	if strings.ToLower(out) == "true" {
		return true
	} else {
		return false
	}
}
func TrackDeploymentSchema(deploymentSchema *schemasv1.DeploymentSchema, deploymentType DeploymentEventType) {
	deploymentSchemaParsed := DeploymentEvent{
		TriggerEvent: TriggerEvent{
			UserUID: deploymentSchema.Creator.Uid,
		},
		CommonProperties: CommonProperties{
			YataiVersion:    version.Version,
			Timestamp:       time.Now(),
			OrganisationUID: deploymentSchema.Cluster.Organization.Uid,
		},
		ClusterUID:          deploymentSchema.Cluster.Uid,
		DeploymentUID:       deploymentSchema.Uid,
		DeploymentEventType: deploymentType,
		DeploymentStatus:    deploymentSchema.Status,
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

	track(deploymentSchemaParsed, "deployment")
}

// Sent the marshaled data to tracking server
func track(data interface{}, eventType string) {
	if donot_track() {
		return
	}
	trackerLog := log.WithField("eventType", eventType)

	jsonData, err := json.Marshal(data)
	if err != nil {
		if is_debug_mode() {
			trackerLog.Error(err, "Failed to marshal data")
		}
		return
	}

	if is_debug_mode() {
		var prettyJSON bytes.Buffer
		_ = json.Indent(&prettyJSON, jsonData, "", " ")
		trackerLog.Info("Tracking Payload: ", prettyJSON.String())
	}

	//TODO: change to t.bentoml.com
	request_url := fmt.Sprintf("%s?token=%s", JITSU_SERVER, JITSUE_KEY)
	resp, err := http.Post(request_url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil && is_debug_mode() {
		trackerLog.Error(err, "failed to send data to tracking server.")
	}
	defer resp.Body.Close()

	if is_debug_mode() {
		if resp.StatusCode == 200 {
			trackerLog.Info("Tracking Request sent.")
		} else {
			trackerLog.Errorf("Tracking Request failed. Status [%s]", resp.Status)
			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)
			trackerLog.Error(bodyString)
		}
	}
}
