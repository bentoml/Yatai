package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/typed/autoscaling/v2beta2"
	"k8s.io/client-go/rest"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	appstypev1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	batchtypev1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	batchtypev1beta "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
	apitypev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	exttypev1beta "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type deploymentService struct{}

var DeploymentService = deploymentService{}

func (s *deploymentService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Deployment{})
}

type CreateDeploymentOption struct {
	CreatorId   uint
	ClusterId   uint
	Name        string
	Description string
}

type UpdateDeploymentOption struct {
	Description *string
}

type ListDeploymentOption struct {
	BaseListOption
}

func (*deploymentService) Create(ctx context.Context, opt CreateDeploymentOption) (*models.Deployment, error) {
	errs := validation.IsDNS1035Label(opt.Name)
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ";"))
	}

	deployment := models.Deployment{
		ResourceMixin: models.ResourceMixin{
			Name: opt.Name,
		},
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		ClusterAssociate: models.ClusterAssociate{
			ClusterId: opt.ClusterId,
		},
		Description: opt.Description,
		Status:      models.DeploymentStatusNonDeployed,
	}
	err := mustGetSession(ctx).Create(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, err
}

func (s *deploymentService) Update(ctx context.Context, b *models.Deployment, opt UpdateDeploymentOption) (*models.Deployment, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Description != nil {
		updaters["description"] = *opt.Description
		defer func() {
			if err == nil {
				b.Description = *opt.Description
			}
		}()
	}

	if len(updaters) == 0 {
		return b, nil
	}

	err = s.getBaseDB(ctx).Where("id = ?", b.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}

	return b, err
}

func (s *deploymentService) Get(ctx context.Context, id uint) (*models.Deployment, error) {
	var deployment models.Deployment
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&deployment).Error
	if err != nil {
		return nil, err
	}
	if deployment.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &deployment, nil
}

func (s *deploymentService) GetByName(ctx context.Context, clusterId uint, name string) (*models.Deployment, error) {
	var deployment models.Deployment
	err := getBaseQuery(ctx, s).Where("cluster_id = ?", clusterId).Where("name = ?", name).First(&deployment).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get deployment %s", name)
	}
	if deployment.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &deployment, nil
}

func (s *deploymentService) List(ctx context.Context, opt ListDeploymentOption) ([]*models.Deployment, uint, error) {
	query := getBaseQuery(ctx, s)
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	deployments := make([]*models.Deployment, 0)
	query = opt.BindQuery(query)
	err = query.Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}
	return deployments, uint(total), err
}

func (s *deploymentService) ListUnsynced(ctx context.Context) ([]*models.Deployment, error) {
	q := getBaseQuery(ctx, s)
	now := time.Now()
	t := now.Add(-time.Minute)
	q = q.Where("status_syncing_at is null or status_syncing_at < ? or status_updated_at is null or status_updated_at < ?", t, t)
	envs := make([]*models.Deployment, 0)
	err := q.Order("id DESC").Find(&envs).Error
	return envs, err
}

type UpdateDeploymentStatusOption struct {
	Status    *models.DeploymentStatus
	SyncingAt **time.Time
	UpdatedAt **time.Time
}

func (s *deploymentService) UpdateStatus(ctx context.Context, deployment *models.Deployment, opt UpdateDeploymentStatusOption) (*models.Deployment, error) {
	updater := map[string]interface{}{}
	if opt.Status != nil {
		deployment.Status = *opt.Status
		updater["status"] = *opt.Status
	}
	if opt.SyncingAt != nil {
		deployment.StatusSyncingAt = *opt.SyncingAt
		updater["status_syncing_at"] = *opt.SyncingAt
	}
	if opt.UpdatedAt != nil {
		deployment.StatusUpdatedAt = *opt.UpdatedAt
		updater["status_updated_at"] = *opt.UpdatedAt
	}
	err := s.getBaseDB(ctx).Where("id = ?", deployment.ID).Updates(updater).Error
	return deployment, err
}

type IDeploymentAssociate interface {
	GetAssociatedDeploymentId() uint
	GetAssociatedDeploymentCache() *models.Deployment
	SetAssociatedDeploymentCache(deployment *models.Deployment)
}

func (s *deploymentService) GetAssociatedDeployment(ctx context.Context, associate IDeploymentAssociate) (*models.Deployment, error) {
	cache := associate.GetAssociatedDeploymentCache()
	if cache != nil {
		return cache, nil
	}
	deployment, err := s.Get(ctx, associate.GetAssociatedDeploymentId())
	associate.SetAssociatedDeploymentCache(deployment)
	return deployment, err
}

func (s *deploymentService) GetKubeNamespace(d *models.Deployment) string {
	return fmt.Sprintf("yatai-deploy-%s", d.Name)
}

func (s *deploymentService) GetKubeCliSet(ctx context.Context, d *models.Deployment) (*kubernetes.Clientset, *rest.Config, error) {
	cluster, err := ClusterService.GetAssociatedCluster(ctx, d)
	if err != nil {
		return nil, nil, errors.Wrap(err, "get associated cluster")
	}
	return ClusterService.GetKubeCliSet(ctx, cluster)
}

func (s *deploymentService) GetKubePodsCli(ctx context.Context, d *models.Deployment) (apitypev1.PodInterface, error) {
	cliset, _, err := s.GetKubeCliSet(ctx, d)
	if err != nil {
		return nil, errors.Wrap(err, "get k8s cliset")
	}
	ns := s.GetKubeNamespace(d)
	podsCli := cliset.CoreV1().Pods(ns)
	return podsCli, nil
}

func (s *deploymentService) GetKubeDeploymentsCli(ctx context.Context, d *models.Deployment) (appstypev1.DeploymentInterface, error) {
	cliset, _, err := s.GetKubeCliSet(ctx, d)
	if err != nil {
		return nil, errors.Wrap(err, "get k8s cliset")
	}
	ns := s.GetKubeNamespace(d)
	deploymentsCli := cliset.AppsV1().Deployments(ns)
	return deploymentsCli, nil
}

func (s *deploymentService) GetKubeHPAsCli(ctx context.Context, d *models.Deployment) (v2beta2.HorizontalPodAutoscalerInterface, error) {
	cliset, _, err := s.GetKubeCliSet(ctx, d)
	if err != nil {
		return nil, errors.Wrap(err, "get k8s cliset")
	}
	ns := s.GetKubeNamespace(d)
	hpaCli := cliset.AutoscalingV2beta2().HorizontalPodAutoscalers(ns)
	return hpaCli, nil
}

func (s *deploymentService) GetKubeServicesCli(ctx context.Context, d *models.Deployment) (apitypev1.ServiceInterface, error) {
	cliset, _, err := s.GetKubeCliSet(ctx, d)
	if err != nil {
		return nil, errors.Wrap(err, "get k8s cliset")
	}
	ns := s.GetKubeNamespace(d)
	servicesCli := cliset.CoreV1().Services(ns)
	return servicesCli, nil
}

func (s *deploymentService) GetKubeStatefulSetsCli(ctx context.Context, d *models.Deployment) (appstypev1.StatefulSetInterface, error) {
	cliset, _, err := s.GetKubeCliSet(ctx, d)
	if err != nil {
		return nil, errors.Wrap(err, "get k8s cliset")
	}
	ns := s.GetKubeNamespace(d)
	servicesCli := cliset.AppsV1().StatefulSets(ns)
	return servicesCli, nil
}

func (s *deploymentService) GetKubeIngressesCli(ctx context.Context, d *models.Deployment) (exttypev1beta.IngressInterface, error) {
	cliset, _, err := s.GetKubeCliSet(ctx, d)
	if err != nil {
		return nil, errors.Wrap(err, "get k8s cliset")
	}
	ns := s.GetKubeNamespace(d)
	ingressesCli := cliset.ExtensionsV1beta1().Ingresses(ns)
	return ingressesCli, nil
}

func (s *deploymentService) GetKubeCronJobsCli(ctx context.Context, d *models.Deployment) (batchtypev1beta.CronJobInterface, error) {
	cliset, _, err := s.GetKubeCliSet(ctx, d)
	if err != nil {
		return nil, errors.Wrap(err, "get k8s cliset")
	}
	ns := s.GetKubeNamespace(d)
	cronJobsCli := cliset.BatchV1beta1().CronJobs(ns)
	return cronJobsCli, nil
}

func (s *deploymentService) GetKubeJobsCli(ctx context.Context, d *models.Deployment) (batchtypev1.JobInterface, error) {
	cliset, _, err := s.GetKubeCliSet(ctx, d)
	if err != nil {
		return nil, errors.Wrap(err, "get k8s cliset")
	}
	ns := s.GetKubeNamespace(d)
	jobsCli := cliset.BatchV1().Jobs(ns)
	return jobsCli, nil
}

func (s *deploymentService) SyncStatus(ctx context.Context, d *models.Deployment) (models.DeploymentStatus, error) {
	now := time.Now()
	nowPtr := &now
	_, err := s.UpdateStatus(ctx, d, UpdateDeploymentStatusOption{
		SyncingAt: &nowPtr,
	})
	if err != nil {
		return d.Status, err
	}
	currentStatus, err := s.getStatusFromK8s(ctx, d)
	if err != nil {
		return d.Status, err
	}
	now = time.Now()
	nowPtr = &now
	_, err = s.UpdateStatus(ctx, d, UpdateDeploymentStatusOption{
		Status:    &currentStatus,
		UpdatedAt: &nowPtr,
	})
	if err != nil {
		return currentStatus, err
	}
	return currentStatus, nil
}

func (s *deploymentService) getStatusFromK8s(ctx context.Context, d *models.Deployment) (models.DeploymentStatus, error) {
	defaultStatus := models.DeploymentStatusUnknown

	cluster, err := ClusterService.GetAssociatedCluster(ctx, d)
	if err != nil {
		return defaultStatus, errors.Wrap(err, "get associated cluster")
	}

	namespace := DeploymentService.GetKubeNamespace(d)

	_, podLister, err := GetPodInformer(ctx, cluster, namespace)
	if err != nil {
		return defaultStatus, err
	}

	pods, err := KubePodService.ListPodsByDeployment(podLister, d)
	if err != nil {
		return defaultStatus, err
	}

	if len(pods) == 0 {
		return models.DeploymentStatusNonDeployed, nil
	}

	events, err := KubeEventService.ListAllKubeEvents(ctx, d)
	if err != nil {
		return defaultStatus, err
	}
	warningsMapping := KubeEventService.GetKubePodsWarningEventsMapping(events, pods)

	hasFailed := false
	hasRunning := false
	hasPending := false

	for _, p := range pods {
		warnings, ok := warningsMapping[p.UID]
		if !ok {
			warnings = make([]apiv1.Event, 0)
		}
		podStatus := KubePodService.GetKubePodStatus(p, warnings)
		if podStatus.Status == "Running" {
			hasRunning = true
		}
		if podStatus.Status == "Failed" {
			hasFailed = true
		}
		if podStatus.Status == "Pending" {
			hasPending = true
		}
	}

	if hasFailed && hasRunning {
		if hasPending {
			return models.DeploymentStatusDeploying, nil
		}
		return models.DeploymentStatusUnhealthy, nil
	}

	if hasPending {
		return models.DeploymentStatusDeploying, nil
	}

	if hasRunning {
		return models.DeploymentStatusRunning, nil
	}

	return models.DeploymentStatusFailed, nil
}
