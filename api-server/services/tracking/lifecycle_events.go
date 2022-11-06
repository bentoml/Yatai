package tracking

import (
	"context"
	"time"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/version"
	"github.com/tianweidut/cron"
)

func AddLifeCycleTrackingCron(ctx context.Context) {
	TrackLifeCycle(ctx, YataiLifeCycleStartup)
	ctx = context.WithValue(ctx, "uptimeStamp", time.Now())

	c := cron.New()
	err := c.AddFunc("@every 1m", func() {
		TrackLifeCycle(ctx, YataiLifeCycleUpdate)
	})

	if err != nil {
		NewTrackerLogger().Errorf("cron add func failed: %s", err.Error())
	}

	c.Start()
}

func TrackLifeCycle(ctx context.Context, event YataiEventType) {
	trackerLog := NewTrackerLogger().WithField("eventType", event)

	orgs, _, err := services.OrganizationService.List(ctx, services.ListOrganizationOption{})

	if err != nil {
		trackerLog.Error("unable to get OrganizationService.List. ", err.Error())
	}

	// defaultOrg
	defaultOrg, err := services.OrganizationService.GetDefault(ctx)

	// sent tracking info for each organization
	for _, org := range orgs {
		// bento
		bentoRepos, numBentoRepos, _ := services.BentoRepositoryService.List(ctx, services.ListBentoRepositoryOption{OrganizationId: &org.ID})
		var numTotalBentos uint
		for _, bentoRepo := range bentoRepos {
			_, numBentos, _ := services.BentoService.List(ctx, services.ListBentoOption{BentoRepositoryId: &bentoRepo.ID})
			numTotalBentos += numBentos
		}

		// model
		modelRepos, numModelRepos, _ := services.ModelRepositoryService.List(ctx, services.ListModelRepositoryOption{OrganizationId: &org.ID})
		var numTotalModels uint
		for _, modelRepo := range modelRepos {
			_, numModels, _ := services.ModelService.List(ctx, services.ListModelOption{ModelRepositoryId: &modelRepo.ID})
			numTotalModels += numModels
		}

		// users
		members, _ := services.OrganizationMemberService.List(ctx, services.ListOrganizationMemberOption{OrganizationId: &org.ID})
		// clusters
		_, numClusters, _ := services.ClusterService.List(ctx, services.ListClusterOption{OrganizationId: &org.ID})

		// deployments
		deployments, numDeployments, _ := services.DeploymentService.List(ctx, services.ListDeploymentOption{OrganizationId: &org.ID})
		var numRunningDeployments uint
		for _, deployment := range deployments {
			if deployment.Status == modelschemas.DeploymentStatusRunning {
				numRunningDeployments++
			}
		}

		uptimeStamp := ctx.Value("uptimeStamp")
		var uptimeDurationSeconds time.Duration
		if uptimeStamp != nil {
			timeNow := time.Now()
			uptimeDurationSeconds = timeNow.Sub(uptimeStamp.(time.Time)) / time.Second
			ctx = context.WithValue(ctx, "uptimeStamp", timeNow)
		}
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
