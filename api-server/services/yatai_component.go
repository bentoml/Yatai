package services

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/release"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/helmchart"
	"github.com/bentoml/yatai/schemas/modelschemas"
)

type yataiComponentService struct{}

var YataiComponentService = yataiComponentService{}

type CreateYataiComponentReleaseOption struct {
	ClusterId uint
	Type      modelschemas.YataiComponentType
}

type DeleteYataiComponentReleaseOption struct {
	ClusterId uint
	Type      modelschemas.YataiComponentType
}

func (s *yataiComponentService) Create(ctx context.Context, opt CreateYataiComponentReleaseOption) (comp *models.YataiComponent, err error) {
	cluster, err := ClusterService.Get(ctx, opt.ClusterId)
	if err != nil {
		return
	}

	clientGetter := helmchart.NewRESTClientGetter(consts.KubeNamespaceYataiOperators, cluster.KubeConfig)
	actionConfig := new(action.Configuration)
	err = actionConfig.Init(clientGetter, consts.KubeNamespaceYataiOperators, "", func(format string, v ...interface{}) {
		logrus.Infof(format, v...)
	})
	if err != nil {
		err = errors.Wrapf(err, "new action config")
		return
	}

	operatorReleaseName := string(opt.Type)

	comps, err := s.List(ctx, opt.ClusterId)
	if err != nil {
		return
	}
	var release_ *release.Release

	for _, c := range comps {
		if c.Release.Name == operatorReleaseName {
			release_ = c.Release
			break
		}
	}

	if release_ == nil {
		dirPath := fmt.Sprintf("scripts/helm-charts/yatai-%s-comp-operator/chart", opt.Type)
		var chartLoader loader.ChartLoader
		chartLoader, err = loader.Loader(dirPath)
		if err != nil {
			return
		}

		var chart_ *chart.Chart
		chart_, err = chartLoader.Load()
		if err != nil {
			return
		}

		install := action.NewInstall(actionConfig)

		install.Namespace = consts.KubeNamespaceYataiOperators
		install.ReleaseName = operatorReleaseName
		install.CreateNamespace = true

		var grafanaHostname string
		grafanaHostname, err = ClusterService.GetGrafanaHostname(ctx, cluster)
		if err != nil {
			return
		}

		release_, err = install.Run(chart_, map[string]interface{}{
			"logging": map[string]interface{}{
				"grafanaHostname": grafanaHostname,
			},
		})
	}

	comp = &models.YataiComponent{
		Type:    opt.Type,
		Release: release_,
	}

	return
}

func (s *yataiComponentService) List(ctx context.Context, clusterId uint) (comps []*models.YataiComponent, err error) {
	cluster, err := ClusterService.Get(ctx, clusterId)
	if err != nil {
		return
	}

	clientGetter := helmchart.NewRESTClientGetter(consts.KubeNamespaceYataiOperators, cluster.KubeConfig)
	actionConfig := new(action.Configuration)
	err = actionConfig.Init(clientGetter, consts.KubeNamespaceYataiOperators, "", func(format string, v ...interface{}) {
		logrus.Infof(format, v...)
	})
	if err != nil {
		err = errors.Wrapf(err, "new action config")
		return
	}

	list := action.NewList(actionConfig)
	list.All = true
	list.AllNamespaces = true
	releases, err := list.Run()
	if err != nil {
		err = errors.Wrapf(err, "list releases")
		return
	}

	for _, release_ := range releases {
		if release_.Namespace != consts.KubeNamespaceYataiOperators {
			continue
		}
		comps = append(comps, &models.YataiComponent{
			Type:    modelschemas.YataiComponentType(release_.Name),
			Release: release_,
		})
	}

	return
}

func (s *yataiComponentService) Delete(ctx context.Context, opt DeleteYataiComponentReleaseOption) (comp *models.YataiComponent, err error) {
	cluster, err := ClusterService.Get(ctx, opt.ClusterId)
	if err != nil {
		return
	}

	clientGetter := helmchart.NewRESTClientGetter(consts.KubeNamespaceYataiOperators, cluster.KubeConfig)
	actionConfig := new(action.Configuration)
	err = actionConfig.Init(clientGetter, consts.KubeNamespaceYataiOperators, "", func(format string, v ...interface{}) {
		logrus.Infof(format, v...)
	})
	if err != nil {
		err = errors.Wrapf(err, "new action config")
		return
	}

	operatorReleaseName := string(opt.Type)

	get := action.NewUninstall(actionConfig)
	release_, err := get.Run(operatorReleaseName)
	operatorReleaseNotFound := errors.Is(err, driver.ErrReleaseNotFound)
	if err != nil && !operatorReleaseNotFound {
		return
	}

	if release_ == nil {
		return nil, nil
	}

	comp = &models.YataiComponent{
		Type:    opt.Type,
		Release: release_.Release,
	}

	return
}
