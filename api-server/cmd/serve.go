package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/common/command"

	"github.com/bentoml/yatai/api-server/routes"
	"github.com/bentoml/yatai/api-server/services"

	"github.com/tianweidut/cron"

	"github.com/bentoml/yatai/common/sync/errsgroup"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func addCron() {
	c := cron.New()
	ctx := context.Background()
	logger := logrus.New().WithField("cron", "sync env")

	err := c.AddFunc("@every 1m", func() {
		ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
		defer cancel()
		logger.Info("listing unsynced deployments")
		deployments, err := services.DeploymentService.ListUnsynced(ctx)
		if err != nil {
			logger.Errorf("list unsynced deployments: %s", err.Error())
		}
		logger.Info("updating unsynced deployments syncing_at")
		now := time.Now()
		nowPtr := &now
		for _, deployment := range deployments {
			_, err := services.DeploymentService.UpdateStatus(ctx, deployment, services.UpdateDeploymentStatusOption{
				SyncingAt: &nowPtr,
			})
			if err != nil {
				logger.Errorf("update deployment %d status: %s", deployment.ID, err.Error())
			}
		}
		logger.Info("updated unsynced deployments syncing_at")
		var eg errsgroup.Group
		eg.SetPoolSize(1000)
		for _, deployment := range deployments {
			deployment := deployment
			eg.Go(func() error {
				_, err := services.DeploymentService.SyncStatus(ctx, deployment)
				return err
			})
		}

		logger.Info("syncing unsynced app deployment deployments...")
		err = eg.WaitWithTimeout(10 * time.Minute)
		logger.Info("synced unsynced app deployment deployments...")
		if err != nil {
			logger.Errorf("sync deployments: %s", err.Error())
		}
	})

	if err != nil {
		logger.Errorf("cron add func failed: %s", err.Error())
	}

	go func() {
		ticker := time.NewTicker(time.Second * 20)
		defer ticker.Stop()
		for {
			func() {
				ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
				defer cancel()
				logger.Info("listing image buld status unsynced bentoVersions")
				bentoVersions, err := services.BentoVersionService.ListImageBuildStatusUnsynced(ctx)
				if err != nil {
					logger.Errorf("list unsynced bento versions: %s", err.Error())
				}
				logger.Info("updating unsynced bento versions image_build_status_syncing_at")
				now := time.Now()
				nowPtr := &now
				for _, bentoVersion := range bentoVersions {
					_, err := services.BentoVersionService.Update(ctx, bentoVersion, services.UpdateBentoVersionOption{
						ImageBuildStatusSyncingAt: &nowPtr,
					})
					if err != nil {
						logger.Errorf("update bento version %d status: %s", bentoVersion.ID, err.Error())
					}
				}
				logger.Info("updated unsynced bento version image_build_status_syncing_at")
				var eg errsgroup.Group
				eg.SetPoolSize(1000)
				for _, deployment := range bentoVersions {
					deployment := deployment
					eg.Go(func() error {
						_, err := services.BentoVersionService.SyncImageBuilderStatus(ctx, deployment)
						return err
					})
				}

				logger.Info("syncing unsynced bento version image build status...")
				err = eg.WaitWithTimeout(10 * time.Minute)
				logger.Info("synced unsynced bento version image build status...")
				if err != nil {
					logger.Errorf("sync bento version: %s", err.Error())
				}
			}()
			<-ticker.C
		}
	}()

	if err != nil {
		logger.Errorf("cron add func failed: %s", err.Error())
	}

	c.Start()
}

type ServeOption struct {
	ConfigPath string
}

func (opt *ServeOption) Validate(ctx context.Context) error {
	return nil
}

func (opt *ServeOption) Complete(ctx context.Context, args []string, argsLenAtDash int) error {
	return nil
}

func initSelfHost(ctx context.Context) error {
	var adminUser *models.User
	users, total, err := services.UserService.List(ctx, services.ListUserOption{
		BaseListOption: services.BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		Perm:  modelschemas.UserPermPtr(modelschemas.UserPermAdmin),
		Order: utils.StringPtr("id ASC"),
	})
	if err != nil {
		return errors.Wrap(err, "list users")
	}
	if total == 0 {
		adminUser, err = services.UserService.Create(ctx, services.CreateUserOption{
			Name:     "admin",
			Password: "admin",
		})
		if err != nil {
			return errors.Wrap(err, "create admin user")
		}
	} else {
		adminUser = users[0]
	}

	var defaultOrg *models.Organization
	orgs, total, err := services.OrganizationService.List(ctx, services.ListOrganizationOption{
		BaseListOption: services.BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		Order: utils.StringPtr("id ASC"),
	})
	if err != nil {
		return errors.Wrapf(err, "list organizations")
	}

	if total == 0 {
		defaultOrg, err = services.OrganizationService.Create(ctx, services.CreateOrganizationOption{
			CreatorId: adminUser.ID,
			Name:      "default",
		})
		if err != nil {
			return errors.Wrapf(err, "create default organization")
		}
		_, err = services.OrganizationMemberService.Create(ctx, adminUser.ID, services.CreateOrganizationMemberOption{
			CreatorId:      adminUser.ID,
			UserId:         adminUser.ID,
			OrganizationId: defaultOrg.ID,
			Role:           modelschemas.MemberRoleAdmin,
		})
		if err != nil {
			return errors.Wrapf(err, "create default organization member")
		}
	} else {
		defaultOrg = orgs[0]
	}

	var defaultCluster *models.Cluster
	_, total, err = services.ClusterService.List(ctx, services.ListClusterOption{
		BaseListOption: services.BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		Order: utils.StringPtr("id ASC"),
	})
	if err != nil {
		return errors.Wrapf(err, "list clusters")
	}

	if total == 0 {
		defaultCluster, err = services.ClusterService.Create(ctx, services.CreateClusterOption{
			CreatorId:      adminUser.ID,
			OrganizationId: defaultOrg.ID,
			Name:           "default",
		})
		if err != nil {
			return errors.Wrapf(err, "create default cluster")
		}
		_, err = services.ClusterMemberService.Create(ctx, adminUser.ID, services.CreateClusterMemberOption{
			CreatorId: adminUser.ID,
			UserId:    adminUser.ID,
			ClusterId: defaultCluster.ID,
			Role:      modelschemas.MemberRoleAdmin,
		})
		if err != nil {
			return errors.Wrapf(err, "create default cluster member")
		}
	}

	return nil
}

func (opt *ServeOption) Run(ctx context.Context, args []string) error {
	if !command.GlobalCommandOption.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	content, err := os.ReadFile(opt.ConfigPath)
	if err != nil {
		return errors.Wrapf(err, "read config file: %s", opt.ConfigPath)
	}

	err = yaml.Unmarshal(content, config.YataiConfig)
	if err != nil {
		return errors.Wrapf(err, "unmarshal config file: %s", opt.ConfigPath)
	}

	err = config.PopulateYataiConfig()
	if err != nil {
		return errors.Wrapf(err, "populate config file: %s", opt.ConfigPath)
	}

	err = services.MigrateUp()
	if err != nil {
		return errors.Wrap(err, "migrate up db")
	}

	if !config.YataiConfig.IsSass {
		err = initSelfHost(ctx)
		if err != nil {
			return errors.Wrap(err, "init self host")
		}
	}

	addCron()

	router, err := routes.NewRouter()
	if err != nil {
		return err
	}

	logrus.Infof("listening on 0.0.0.0:%d", config.YataiConfig.Server.Port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.YataiConfig.Server.Port),
		Handler: router,
	}
	return srv.ListenAndServe()
}

func getServeCmd() *cobra.Command {
	var opt ServeOption
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "run yatai api server",
		Long:  "",
		RunE:  command.MakeRunE(&opt),
	}
	cmd.Flags().StringVarP(&opt.ConfigPath, "config", "c", "./yatai-config.dev.yaml", "")
	return cmd
}
