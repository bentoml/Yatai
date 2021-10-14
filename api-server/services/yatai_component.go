package services

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/bentoml/grafana-operator/api/integreatly/v1alpha1"
	"github.com/huandu/xstrings"
	"gopkg.in/yaml.v3"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"helm.sh/helm/v3/pkg/release"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

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

func (s *yataiComponentService) ListOperatorHelmCharts(ctx context.Context) (charts []*chart.Chart, err error) {
	dirPaths, err := filepath.Glob("scripts/helm-charts/yatai-*-comp-operator/chart")

	for _, dirPath := range dirPaths {
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

		charts = append(charts, chart_)
	}

	return
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

	var grafanaHostname string
	grafanaHostname, err = ClusterService.GenerateGrafanaHostname(ctx, cluster)
	if err != nil {
		return
	}

	var grafanaRootPath string
	grafanaRootPath, err = ClusterService.GetGrafanaRootPath(ctx, cluster)
	if err != nil {
		return
	}

	var grafana *v1alpha1.Grafana
	grafana, err = ClusterService.GetGrafana(ctx, cluster)
	if err != nil && !k8serrors.IsNotFound(err) {
		return
	}
	err = nil

	if grafana != nil {
		grafanaHostname = grafana.Spec.Ingress.Hostname
	}

	values := map[string]interface{}{
		string(opt.Type): map[string]interface{}{
			"grafana": map[string]interface{}{
				"hostname": grafanaHostname,
				"rootUrl":  fmt.Sprintf("%%(protocol)s://%%(domain)s:%%(http_port)s%s", grafanaRootPath),
			},
		},
	}

	if release_ == nil {
		install := action.NewInstall(actionConfig)

		install.Namespace = consts.KubeNamespaceYataiOperators
		install.ReleaseName = operatorReleaseName
		install.CreateNamespace = true
		install.Wait = true
		install.Timeout = time.Minute * 5

		release_, err = install.Run(chart_, values)
	} else if release_.Chart.Metadata.Version != chart_.Metadata.Version {
		upgrade := action.NewUpgrade(actionConfig)

		upgrade.Namespace = consts.KubeNamespaceYataiOperators
		upgrade.Wait = true
		upgrade.Timeout = time.Minute * 5

		release_, err = upgrade.Run(release_.Name, chart_, values)
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
		return nil, nil
	}

	meta := &struct {
		Kind       string `yaml:"kind"`
		APIVersion string `yaml:"apiVersion"`
	}{}
	resourceYamlFileName := fmt.Sprintf("%s.yaml", opt.Type)
	for _, f := range release_.Chart.Templates {
		if filepath.Base(f.Name) != resourceYamlFileName {
			continue
		}
		err = yaml.Unmarshal(f.Data, meta)
		if err != nil {
			return
		}
		break
	}

	if meta.Kind != "" {
		group, _, version := xstrings.LastPartition(meta.APIVersion, "/")
		var restConf *rest.Config
		_, restConf, err = ClusterService.GetKubeCliSet(ctx, cluster)
		if err != nil {
			return
		}
		var client dynamic.Interface
		client, err = dynamic.NewForConfig(restConf)
		if err != nil {
			return
		}
		utdsCli := client.Resource(schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: fmt.Sprintf("%ss", opt.Type),
		})
		_, err = utdsCli.Get(ctx, string(opt.Type), metav1.GetOptions{})
		isNotFound := k8serrors.IsNotFound(err)
		if err != nil && !isNotFound {
			return
		}
		if !isNotFound {
			err = utdsCli.Delete(ctx, string(opt.Type), metav1.DeleteOptions{})
			if err != nil {
				return
			}
			// FIXME: waiting yatai component operator to cleanup resources, I know it's stupid, but I haven't found a better way
			time.Sleep(7 * time.Second)
		}
	}

	uninstall := action.NewUninstall(actionConfig)
	_, err = uninstall.Run(operatorReleaseName)
	if err != nil {
		return
	}

	comp = &models.YataiComponent{
		Type:    opt.Type,
		Release: release_,
	}

	return
}
