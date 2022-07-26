package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tianweidut/cron"
	"gopkg.in/yaml.v3"

	"github.com/bentoml/yatai/api-server/config"
	"github.com/bentoml/yatai/api-server/routes"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/common/command"
	"github.com/bentoml/yatai/common/sync/errsgroup"
)

func addCron(ctx context.Context) {
	c := cron.New()
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
				logger.Info("listing image build status unsynced bentos")
				bentos, err := services.BentoService.ListImageBuildStatusUnsynced(ctx)
				if err != nil {
					logger.Errorf("list unsynced bentos: %s", err.Error())
				}
				logger.Info("updating unsynced bentos image_build_status_syncing_at")
				now := time.Now()
				nowPtr := &now
				for _, bento := range bentos {
					_, err := services.BentoService.Update(ctx, bento, services.UpdateBentoOption{
						ImageBuildStatusSyncingAt: &nowPtr,
					})
					if err != nil {
						logger.Errorf("update bento %d status: %s", bento.ID, err.Error())
					}
				}
				logger.Info("updated unsynced bento image_build_status_syncing_at")
				var eg errsgroup.Group
				eg.SetPoolSize(1000)
				for _, bento := range bentos {
					bento := bento
					eg.Go(func() error {
						_, err := services.BentoService.SyncImageBuilderStatus(ctx, bento)
						return err
					})
				}

				logger.Info("syncing unsynced bento image build status...")
				err = eg.WaitWithTimeout(10 * time.Minute)
				logger.Info("synced unsynced bento image build status...")
				if err != nil {
					logger.Errorf("sync bento: %s", err.Error())
				}
			}()
			<-ticker.C
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * 20)
		defer ticker.Stop()
		for {
			func() {
				ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
				defer cancel()
				logger.Info("listing image build status unsynced models")
				models_, err := services.ModelService.ListImageBuildStatusUnsynced(ctx)
				if err != nil {
					logger.Errorf("list unsynced models: %s", err.Error())
				}
				logger.Info("updating unsynced models image_build_status_syncing_at")
				now := time.Now()
				nowPtr := &now
				for _, model := range models_ {
					_, err := services.ModelService.Update(ctx, model, services.UpdateModelOption{
						ImageBuildStatusSyncingAt: &nowPtr,
					})
					if err != nil {
						logger.Errorf("update model %d status: %s", model.ID, err.Error())
					}
				}
				logger.Info("updated unsynced model image_build_status_syncing_at")
				var eg errsgroup.Group
				eg.SetPoolSize(1000)
				for _, model := range models_ {
					model := model
					eg.Go(func() error {
						_, err := services.ModelService.SyncImageBuilderStatus(ctx, model)
						return err
					})
				}

				logger.Info("syncing unsynced model image build status...")
				err = eg.WaitWithTimeout(10 * time.Minute)
				logger.Info("synced unsynced model image build status...")
				if err != nil {
					logger.Errorf("sync model: %s", err.Error())
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
	defaultOrg, err := services.OrganizationService.GetDefault(ctx)
	if err != nil {
		return errors.Wrap(err, "get default org")
	}

	_, err = services.ClusterService.GetDefault(ctx, defaultOrg.ID)

	return err
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

	addCron(ctx)

	router, err := routes.NewRouter()
	if err != nil {
		return err
	}

	readHeaderTimeout := 10 * time.Second
	if config.YataiConfig.Server.ReadHeaderTimeout > 0 {
		readHeaderTimeout = time.Duration(config.YataiConfig.Server.ReadHeaderTimeout) * time.Second
	}

	logrus.Infof("listening on 0.0.0.0:%d", config.YataiConfig.Server.Port)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.YataiConfig.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
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
