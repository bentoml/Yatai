package tracking

import (
	"context"
	"sync"
	"time"

	"github.com/tianweidut/cron"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/version"
)

const (
	LIFECYCLE_CRON_SCHEDULE       = "@every 6h"
	LIFECYCLE_CRON_SCHEDULE_DEBUG = "@every 1m"
)

var (
	yataiUpTimestamp   time.Time
	yataiUpTimestampMu sync.Mutex
)

func resetYataiUpTimestamp() {
	yataiUpTimestampMu.Lock()
	defer yataiUpTimestampMu.Unlock()
	yataiUpTimestamp = time.Now()
}

func init() {
	resetYataiUpTimestamp()
}

func AddLifeCycleTrackingCron(ctx context.Context, c *cron.Cron) {
	TrackLifeCycle(ctx, YataiLifeCycleStartup)

	var cron_schedule string
	if !isTrackingDebug() {
		cron_schedule = LIFECYCLE_CRON_SCHEDULE
	} else {
		cron_schedule = LIFECYCLE_CRON_SCHEDULE_DEBUG
	}
	err := c.AddFunc(cron_schedule, func() {
		TrackLifeCycle(ctx, YataiLifeCycleUpdate)
	})

	if err != nil {
		NewTrackerLogger().Errorf("cron add func failed: %s", err.Error())
	}
}

func TrackLifeCycle(ctx context.Context, event YataiEventType) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	trackerLog := NewTrackerLogger().WithField("eventType", event)

	orgs, _, err := services.OrganizationService.List(ctx, services.ListOrganizationOption{})

	if err != nil {
		trackerLog.Error("unable to get organizations ", err.Error())
		return
	}

	// defaultOrg
	defaultOrg, err := services.OrganizationService.GetDefault(ctx)
	if err != nil {
		trackingLogger.Error("Unnable to get defaultOrg: ", err)
		return
	}

	// sent tracking info for each organization
	for _, org := range orgs {
		// bento
		bentoRepos, numBentoRepos, err := services.BentoRepositoryService.List(ctx, services.ListBentoRepositoryOption{OrganizationId: &org.ID})
		if err != nil {
			trackerLog.Error("unable to get bento repositories ", err.Error())
			return
		}
		var numTotalBentos uint
		for _, bentoRepo := range bentoRepos {
			_, numBentos, err := services.BentoService.List(ctx, services.ListBentoOption{BentoRepositoryId: &bentoRepo.ID})
			if err != nil {
				trackerLog.Error("unable to get bentos ", err.Error())
				return
			}
			numTotalBentos += numBentos
		}

		// model
		modelRepos, numModelRepos, err := services.ModelRepositoryService.List(ctx, services.ListModelRepositoryOption{OrganizationId: &org.ID})
		if err != nil {
			trackerLog.Error("unable to get model repositories ", err.Error())
			return
		}
		var numTotalModels uint
		for _, modelRepo := range modelRepos {
			_, numModels, err := services.ModelService.List(ctx, services.ListModelOption{ModelRepositoryId: &modelRepo.ID})
			if err != nil {
				trackerLog.Error("unable to get models ", err.Error())
				return
			}
			numTotalModels += numModels
		}

		// users
		members, err := services.OrganizationMemberService.List(ctx, services.ListOrganizationMemberOption{OrganizationId: &org.ID})
		if err != nil {
			trackerLog.Error("unable to get organization members ", err.Error())
			return
		}
		// clusters
		_, numClusters, err := services.ClusterService.List(ctx, services.ListClusterOption{OrganizationId: &org.ID})
		if err != nil {
			trackerLog.Error("unable to get clusters ", err.Error())
			return
		}

		// deployments
		deployments, numDeployments, err := services.DeploymentService.List(ctx, services.ListDeploymentOption{OrganizationId: &org.ID})
		if err != nil {
			trackerLog.Error("unable to get deployments ", err.Error())
			return
		}
		var numRunningDeployments uint
		for _, deployment := range deployments {
			if deployment.Status == modelschemas.DeploymentStatusRunning {
				numRunningDeployments++
			}
		}

		timeNow := time.Now()
		uptimeDurationSeconds := timeNow.Sub(yataiUpTimestamp) / time.Second
		resetYataiUpTimestamp()
		lifecycleEvent := LifeCycleEvent{
			CommonProperties: NewCommonProperties(
				event, org.Uid, defaultOrg.Uid, version.Version),
			UptimeDurationSeconds: uptimeDurationSeconds,
			NumBentoRepositories:  numBentoRepos,
			NumTotalBentos:        numTotalBentos,
			NumModelRepositories:  numModelRepos,
			NumTotalModels:        numTotalModels,
			NumUsers:              uint(len(members)),
			NumClusters:           numClusters,
			NumDeployments:        numDeployments,
		}

		track(ctx, lifecycleEvent, event)
	}
}
