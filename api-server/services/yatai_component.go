package services

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	listerAppsV1 "k8s.io/client-go/listers/apps/v1"

	"github.com/bentoml/grafana-operator/api/integreatly/v1alpha1"
	"github.com/huandu/xstrings"
	"gopkg.in/yaml.v3"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlDecoder "k8s.io/apimachinery/pkg/util/yaml"
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
	"github.com/bentoml/yatai/schemas/modelschemas"
)

type yataiComponentService struct{}

var YataiComponentService = yataiComponentService{}

var helmReleaseMagicGzip = []byte{0x1f, 0x8b, 0x08}

type CreateYataiComponentReleaseOption struct {
	ClusterId uint
	Type      modelschemas.YataiComponentType
}

type GetYataiComponentReleaseOption struct {
	ClusterId uint
	Type      modelschemas.YataiComponentType
}

type DeleteYataiComponentReleaseOption struct {
	ClusterId uint
	Type      modelschemas.YataiComponentType
}

type ListYataiComponentHelmChartReleaseResourcesOption struct {
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

	clientGetter, err := ClusterService.GetRESTClientGetter(ctx, cluster, consts.KubeNamespaceYataiOperators)
	if err != nil {
		return
	}
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

	var values map[string]interface{}

	if opt.Type == modelschemas.YataiComponentTypeDeployment {
		values = map[string]interface{}{
			string(opt.Type): map[string]interface{}{
				"minio":          map[string]interface{}{},
				"dockerRegistry": map[string]interface{}{},
			},
		}
	} else {
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

		values = map[string]interface{}{
			string(opt.Type): map[string]interface{}{
				"grafana": map[string]interface{}{
					"hostname":         grafanaHostname,
					"rootUrl":          fmt.Sprintf("%%(protocol)s://%%(domain)s:%%(http_port)s%s", grafanaRootPath),
					"ingressClassName": consts.KubeIngressClassName,
				},
			},
		}
	}

	if release_ == nil {
		install := action.NewInstall(actionConfig)

		install.Namespace = consts.KubeNamespaceYataiOperators
		install.ReleaseName = operatorReleaseName
		install.CreateNamespace = true

		release_, err = install.Run(chart_, values)
	} else if release_.Chart.Metadata.Version != chart_.Metadata.Version {
		upgrade := action.NewUpgrade(actionConfig)

		upgrade.Namespace = consts.KubeNamespaceYataiOperators

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

	clientGetter, err := ClusterService.GetRESTClientGetter(ctx, cluster, consts.KubeNamespaceYataiOperators)
	if err != nil {
		return
	}
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

func (s *yataiComponentService) getHelmReleaseNames(type_ modelschemas.YataiComponentType) (releaseNames []string) {
	switch type_ {
	case modelschemas.YataiComponentTypeDeployment:
		return []string{
			"yatai-ingress-controller",
			"yatai-minio",
			"yatai-docker-registry",
		}
	case modelschemas.YataiComponentTypeLogging:
		return []string{
			"yatai-grafana",
			"yatai-loki",
			"yatai-minio",
			"yatai-promtail",
		}
	case modelschemas.YataiComponentTypeMonitoring:
		return []string{
			"yatai-grafana",
			"yatai-prometheus",
		}
	}
	return
}

func (s *yataiComponentService) ListHelmChartReleaseResources(ctx context.Context, opt ListYataiComponentHelmChartReleaseResourcesOption) (resources []*models.KubeResource, err error) {
	cluster, err := ClusterService.Get(ctx, opt.ClusterId)
	if err != nil {
		return
	}
	_, secretLister, err := GetSecretInformer(ctx, cluster, consts.KubeNamespaceYataiComponents)
	if err != nil {
		return
	}
	selector, err := labels.Parse("owner=helm")
	if err != nil {
		return
	}
	secrets, err := secretLister.List(selector)
	if err != nil {
		return
	}
	releaseNames := s.getHelmReleaseNames(opt.Type)
	for _, releaseName := range releaseNames {
		for _, secret := range secrets {
			if !strings.HasPrefix(secret.Name, fmt.Sprintf("sh.helm.release.v1.%s.", releaseName)) {
				continue
			}
			var release_ *release.Release
			release_, err = s.decodeHelm3(string(secret.Data["release"]))
			if err != nil {
				return
			}
			var resources_ []*models.KubeResource
			resources_, err = s.resourcesFromManifest(release_.Namespace, release_.Manifest)
			if err != nil {
				return
			}
			for _, resource := range resources_ {
				switch resource.Kind {
				case "StatefulSet":
					var stsLister listerAppsV1.StatefulSetNamespaceLister
					_, stsLister, err = GetStatefulSetInformer(ctx, cluster, resource.Namespace)
					logrus.Info("[done] get statefulSet informer")
					if err != nil {
						return nil, errors.Wrap(err, "get statefulSet informer")
					}
					var sts *v1.StatefulSet
					sts, err = stsLister.Get(resource.Name)
					if err != nil {
						return nil, errors.Wrapf(err, "get release %s statefulSet %s", release_.Name, resource.Name)
					}
					resource.MatchLabels = sts.Spec.Selector.MatchLabels
				case "Deployment":
					var deploymentNamespaceLister listerAppsV1.DeploymentNamespaceLister
					_, deploymentNamespaceLister, err = GetDeploymentInformer(ctx, cluster, resource.Namespace)
					logrus.Info("[done] get deployment informer")
					if err != nil {
						return nil, errors.Wrap(err, "get deployment informer")
					}
					var deployment *v1.Deployment
					deployment, err = deploymentNamespaceLister.Get(resource.Name)
					if err != nil {
						return nil, errors.Wrapf(err, "get release %s deployment %s", release_.Name, resource.Name)
					}
					resource.MatchLabels = deployment.Spec.Selector.MatchLabels
				case "DaemonSet":
					var daemonSetNamespaceLister listerAppsV1.DaemonSetNamespaceLister
					_, daemonSetNamespaceLister, err = GetDaemonSetInformer(ctx, cluster, resource.Namespace)
					logrus.Info("[done] get daemonSet informer")
					if err != nil {
						return nil, errors.Wrap(err, "get daemonSet informer")
					}
					var daemonSet *v1.DaemonSet
					daemonSet, err = daemonSetNamespaceLister.Get(resource.Name)
					if err != nil {
						return nil, errors.Wrapf(err, "get release %s damonseSet %s", release_.Name, resource.Name)
					}
					resource.MatchLabels = daemonSet.Spec.Selector.MatchLabels
				default:
					continue
				}
				resources = append(resources, resource)
			}
			break
		}
	}
	return
}

func (s *yataiComponentService) Get(ctx context.Context, opt GetYataiComponentReleaseOption) (comp *models.YataiComponent, err error) {
	cluster, err := ClusterService.Get(ctx, opt.ClusterId)
	if err != nil {
		return
	}

	clientGetter, err := ClusterService.GetRESTClientGetter(ctx, cluster, consts.KubeNamespaceYataiOperators)
	if err != nil {
		return
	}
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

	meta_ := &struct {
		Kind       string `yaml:"kind"`
		APIVersion string `yaml:"apiVersion"`
	}{}
	resourceYamlFileName := fmt.Sprintf("%s.yaml", opt.Type)
	for _, f := range release_.Chart.Templates {
		if filepath.Base(f.Name) != resourceYamlFileName {
			continue
		}
		data := strings.Join(strings.Split(string(f.Data), "\n")[:2], "\n")
		err = yaml.Unmarshal([]byte(data), meta_)
		if err != nil {
			return
		}
		break
	}

	if meta_.Kind != "" {
		group, _, version := xstrings.LastPartition(meta_.APIVersion, "/")
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
	}

	comp = &models.YataiComponent{
		Type:    opt.Type,
		Release: release_,
	}

	return
}

func (s *yataiComponentService) Delete(ctx context.Context, opt DeleteYataiComponentReleaseOption) (comp *models.YataiComponent, err error) {
	if opt.Type == modelschemas.YataiComponentTypeDeployment {
		err = errors.New("not support delete yatai deployment component")
		return
	}

	cluster, err := ClusterService.Get(ctx, opt.ClusterId)
	if err != nil {
		return
	}

	clientGetter, err := ClusterService.GetRESTClientGetter(ctx, cluster, consts.KubeNamespaceYataiOperators)
	if err != nil {
		return
	}
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

	meta_ := &struct {
		Kind       string `yaml:"kind"`
		APIVersion string `yaml:"apiVersion"`
	}{}
	resourceYamlFileName := fmt.Sprintf("%s.yaml", opt.Type)
	for _, f := range release_.Chart.Templates {
		if filepath.Base(f.Name) != resourceYamlFileName {
			continue
		}
		data := strings.Join(strings.Split(string(f.Data), "\n")[:2], "\n")
		err = yaml.Unmarshal([]byte(data), meta_)
		if err != nil {
			return
		}
		break
	}

	if meta_.Kind != "" {
		group, _, version := xstrings.LastPartition(meta_.APIVersion, "/")
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

func (s *yataiComponentService) ToObjects(in io.Reader) ([]runtime.Object, error) {
	var result []runtime.Object
	reader := yamlDecoder.NewYAMLReader(bufio.NewReaderSize(in, 4096))
	for {
		raw, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		obj, err := s.toObjects(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, obj...)
	}

	return result, nil
}

func (s *yataiComponentService) toObjects(bytes_ []byte) ([]runtime.Object, error) {
	bytes_, err := yamlDecoder.ToJSON(bytes_)
	if err != nil {
		return nil, err
	}

	check := map[string]interface{}{}
	if err := json.Unmarshal(bytes_, &check); err != nil || len(check) == 0 {
		return nil, err
	}

	obj, _, err := unstructured.UnstructuredJSONScheme.Decode(bytes_, nil, nil)
	if err != nil {
		return nil, err
	}

	if l, ok := obj.(*unstructured.UnstructuredList); ok {
		var result []runtime.Object
		for _, obj := range l.Items {
			copy_ := obj
			result = append(result, &copy_)
		}
		return result, nil
	}

	return []runtime.Object{obj}, nil
}

func (s *yataiComponentService) decodeHelm3(data string) (*release.Release, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	// For backwards compatibility with releases that were stored before
	// compression was introduced we skip decompression if the
	// gzip magic header is not found
	if len(b) <= 3 {
		return nil, errors.New(fmt.Sprintf("content: %s not valid", data))
	}
	if bytes.Equal(b[0:3], helmReleaseMagicGzip) {
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		b2, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		b = b2
	}

	var rls release.Release
	// unmarshal release object bytes
	if err := json.Unmarshal(b, &rls); err != nil {
		return nil, err
	}
	return &rls, nil
}

func (s *yataiComponentService) resourcesFromManifest(namespace, manifest string) (result []*models.KubeResource, err error) {
	objs, err := s.ToObjects(bytes.NewReader([]byte(manifest)))
	if err != nil {
		return nil, err
	}

	for _, obj := range objs {
		accessor, err := meta.Accessor(obj)
		if err != nil {
			return nil, err
		}
		r := models.KubeResource{
			Name:      accessor.GetName(),
			Namespace: accessor.GetNamespace(),
		}
		gvk := obj.GetObjectKind().GroupVersionKind()
		if r.Namespace == "" {
			r.Namespace = namespace
		}
		r.APIVersion, r.Kind = gvk.ToAPIVersionAndKind()

		result = append(result, &r)
	}

	return result, nil
}
