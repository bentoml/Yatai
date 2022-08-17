package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/xid"
	"gorm.io/gorm"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	appstypev1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/kubernetes/typed/autoscaling/v2beta2"
	batchtypev1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	batchtypev1beta "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
	apitypev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	networkingtypev1 "k8s.io/client-go/kubernetes/typed/networking/v1"
	"k8s.io/client-go/rest"

	"github.com/bentoml/yatai-common/system"
	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
	"github.com/bentoml/yatai/common/utils"

	servingv1alpha2 "github.com/bentoml/yatai-deployment/generated/serving/clientset/versioned/typed/serving/v1alpha2"
)

type deploymentService struct{}

var DeploymentService = deploymentService{}

func (s *deploymentService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.Deployment{})
}

type CreateDeploymentOption struct {
	CreatorId     uint
	ClusterId     uint
	Name          string
	Description   string
	Labels        modelschemas.LabelItemsSchema
	KubeNamespace string
}

type UpdateDeploymentOption struct {
	Description *string
	Labels      *modelschemas.LabelItemsSchema
	Status      *modelschemas.DeploymentStatus
}

type UpdateDeploymentStatusOption struct {
	Status    *modelschemas.DeploymentStatus
	SyncingAt **time.Time
	UpdatedAt **time.Time
	Labels    *modelschemas.LabelItemsSchema
}

type ListDeploymentOption struct {
	BaseListOption
	BaseListByLabelsOption
	ClusterId       *uint
	CreatorId       *uint
	LastUpdaterId   *uint
	OrganizationId  *uint
	ClusterIds      *[]uint
	CreatorIds      *[]uint
	LastUpdaterIds  *[]uint
	OrganizationIds *[]uint
	Ids             *[]uint
	BentoIds        *[]uint
	Statuses        *[]modelschemas.DeploymentStatus
	Order           *string
}

func (*deploymentService) Create(ctx context.Context, opt CreateDeploymentOption) (*models.Deployment, error) {
	errs := validation.IsDNS1035Label(opt.Name)
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ";"))
	}

	errs = validation.IsDNS1035Label(opt.KubeNamespace)
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ";"))
	}

	guid := xid.New()

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
		Description:     opt.Description,
		Status:          modelschemas.DeploymentStatusNonDeployed,
		KubeDeployToken: guid.String(),
		KubeNamespace:   opt.KubeNamespace,
	}
	err := mustGetSession(ctx).Create(&deployment).Error
	if err != nil {
		return nil, err
	}
	cluster, err := ClusterService.Get(ctx, opt.ClusterId)
	if err != nil {
		return nil, err
	}
	org, err := OrganizationService.GetAssociatedOrganization(ctx, cluster)
	if err != nil {
		return nil, err
	}
	err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, opt.Labels, opt.CreatorId, org.ID, &deployment)
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

	if opt.Status != nil {
		updaters["status"] = *opt.Status
		defer func() {
			if err == nil {
				b.Status = *opt.Status
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
	if opt.Labels != nil {
		cluster, err := ClusterService.GetAssociatedCluster(ctx, b)
		if err != nil {
			return nil, err
		}
		org, err := OrganizationService.GetAssociatedOrganization(ctx, cluster)
		if err != nil {
			return nil, err
		}
		user, err := GetCurrentUser(ctx)
		if err != nil {
			return nil, err
		}
		err = LabelService.CreateOrUpdateLabelsFromLabelItemsSchema(ctx, *opt.Labels, user.ID, org.ID, b)
		if err != nil {
			return nil, err
		}
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

func (s *deploymentService) GetByUid(ctx context.Context, uid string) (*models.Deployment, error) {
	var deployment models.Deployment
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&deployment).Error
	if err != nil {
		return nil, err
	}
	if deployment.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &deployment, nil
}

func (s *deploymentService) GetByName(ctx context.Context, clusterId uint, kubeNamespace, name string) (*models.Deployment, error) {
	var deployment models.Deployment
	err := getBaseQuery(ctx, s).Where("cluster_id = ?", clusterId).Where("kube_namespace = ?", kubeNamespace).Where("name = ?", name).First(&deployment).Error
	if err != nil {
		return nil, errors.Wrapf(err, "get deployment %s", name)
	}
	if deployment.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &deployment, nil
}

func (s *deploymentService) ListByUids(ctx context.Context, uids []string) ([]*models.Deployment, error) {
	deployments := make([]*models.Deployment, 0, len(uids))
	if len(uids) == 0 {
		return deployments, nil
	}
	err := getBaseQuery(ctx, s).Where("uid in (?)", uids).Find(&deployments).Error
	return deployments, err
}

func (s *deploymentService) List(ctx context.Context, opt ListDeploymentOption) ([]*models.Deployment, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.OrganizationId != nil {
		query = query.Joins("LEFT JOIN cluster ON cluster.id = deployment.cluster_id")
		query = query.Where("cluster.organization_id = ?", *opt.OrganizationId)
	}
	if opt.Ids != nil {
		query = query.Where("deployment.id in (?)", *opt.Ids)
	}
	if opt.OrganizationIds != nil {
		query = query.Joins("LEFT JOIN cluster ON cluster.id = deployment.cluster_id")
		query = query.Where("cluster.organization_id IN (?)", *opt.OrganizationIds)
	}
	query = query.Joins("LEFT JOIN deployment_revision ON deployment_revision.deployment_id = deployment.id AND deployment_revision.status = ?", modelschemas.DeploymentRevisionStatusActive)
	if opt.LastUpdaterId != nil {
		query = query.Where("deployment_revision.creator_id = ?", *opt.LastUpdaterId)
	}
	if opt.LastUpdaterIds != nil {
		query = query.Where("deployment_revision.creator_id IN (?)", *opt.LastUpdaterIds)
	}
	if opt.BentoIds != nil {
		query = query.Joins("LEFT JOIN deployment_target ON deployment_target.deployment_revision_id = deployment_revision.id").Where("deployment_target.bento_id IN (?)", *opt.BentoIds)
	}
	if opt.ClusterId != nil {
		query = query.Where("deployment.cluster_id = ?", *opt.ClusterId)
	}
	if opt.ClusterIds != nil {
		query = query.Where("deployment.cluster_id IN (?)", *opt.ClusterIds)
	}
	if opt.CreatorId != nil {
		query = query.Where("deployment.creator_id = ?", *opt.CreatorId)
	}
	if opt.CreatorIds != nil {
		query = query.Where("deployment.creator_id IN (?)", *opt.CreatorIds)
	}
	if opt.Statuses != nil {
		query = query.Where("deployment.status IN (?)", *opt.Statuses)
	}
	query = opt.BindQueryWithKeywords(query, "deployment")
	query = opt.BindQueryWithLabels(query, modelschemas.ResourceTypeDeployment)
	query = query.Select("deployment_revision.*, deployment.*")
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	query = opt.BindQueryWithLimit(query)
	if opt.Order != nil {
		query = query.Order(*opt.Order)
	} else {
		query.Order("deployment.id DESC")
	}
	deployments := make([]*models.Deployment, 0)
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

type INullableDeploymentAssociate interface {
	GetAssociatedDeploymentId() *uint
	GetAssociatedDeploymentCache() *models.Deployment
	SetAssociatedDeploymentCache(cluster *models.Deployment)
}

func (s *deploymentService) GetAssociatedNullableDeployment(ctx context.Context, associate INullableDeploymentAssociate) (*models.Deployment, error) {
	cache := associate.GetAssociatedDeploymentCache()
	if cache != nil {
		return cache, nil
	}
	deploymentId := associate.GetAssociatedDeploymentId()
	if deploymentId == nil {
		return nil, nil
	}
	deployment, err := s.Get(ctx, *deploymentId)
	associate.SetAssociatedDeploymentCache(deployment)
	return deployment, err
}

func (s *deploymentService) GetKubeNamespace(d *models.Deployment) string {
	return d.KubeNamespace
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

func (s *deploymentService) GetKubeIngressesCli(ctx context.Context, d *models.Deployment) (networkingtypev1.IngressInterface, error) {
	cliset, _, err := s.GetKubeCliSet(ctx, d)
	if err != nil {
		return nil, errors.Wrap(err, "get k8s cliset")
	}
	ns := s.GetKubeNamespace(d)
	ingressesCli := cliset.NetworkingV1().Ingresses(ns)
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

func (s *deploymentService) GetKubeBentoDeploymentCli(ctx context.Context, d *models.Deployment) (servingv1alpha2.BentoDeploymentInterface, error) {
	_, restConf, err := s.GetKubeCliSet(ctx, d)
	if err != nil {
		return nil, errors.Wrap(err, "get k8s cliset")
	}
	ns := s.GetKubeNamespace(d)
	cli, err := servingv1alpha2.NewForConfig(restConf)
	if err != nil {
		return nil, errors.Wrap(err, "get bento deployment cliset")
	}
	return cli.BentoDeployments(ns), nil
}

func (s *deploymentService) SyncStatus(ctx context.Context, d *models.Deployment) (modelschemas.DeploymentStatus, error) {
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

func (s *deploymentService) getStatusFromK8s(ctx context.Context, d *models.Deployment) (modelschemas.DeploymentStatus, error) {
	defaultStatus := modelschemas.DeploymentStatusUnknown

	cluster, err := ClusterService.GetAssociatedCluster(ctx, d)
	if err != nil {
		return defaultStatus, errors.Wrap(err, "get associated cluster")
	}

	namespace := DeploymentService.GetKubeNamespace(d)

	_, podLister, err := GetPodInformer(ctx, cluster, namespace)
	if err != nil {
		return defaultStatus, err
	}

	pods, err := KubePodService.ListPodsByDeployment(ctx, podLister, d)
	if err != nil {
		return defaultStatus, err
	}

	if len(pods) == 0 {
		if d.Status == modelschemas.DeploymentStatusTerminating || d.Status == modelschemas.DeploymentStatusTerminated {
			return modelschemas.DeploymentStatusTerminated, nil
		}
		if d.Status == modelschemas.DeploymentStatusDeploying {
			return modelschemas.DeploymentStatusDeploying, nil
		}
		return modelschemas.DeploymentStatusNonDeployed, nil
	}

	if d.Status == modelschemas.DeploymentStatusTerminated {
		return d.Status, nil
	}

	hasFailed := false
	hasRunning := false
	hasPending := false

	for _, p := range pods {
		podStatus := p.Status
		if podStatus.Status == modelschemas.KubePodActualStatusRunning {
			hasRunning = true
		}
		if podStatus.Status == modelschemas.KubePodActualStatusFailed {
			hasFailed = true
		}
		if podStatus.Status == modelschemas.KubePodActualStatusPending {
			hasPending = true
		}
	}

	if d.Status == modelschemas.DeploymentStatusTerminating {
		if !hasRunning {
			return modelschemas.DeploymentStatusTerminated, nil
		}
		return d.Status, nil
	}

	if hasFailed && hasRunning {
		if hasPending {
			return modelschemas.DeploymentStatusDeploying, nil
		}
		return modelschemas.DeploymentStatusUnhealthy, nil
	}

	if hasPending {
		return modelschemas.DeploymentStatusDeploying, nil
	}

	if hasRunning {
		return modelschemas.DeploymentStatusRunning, nil
	}

	return modelschemas.DeploymentStatusFailed, nil
}

func (s *deploymentService) UpdateKubeDeployToken(ctx context.Context, deployment *models.Deployment, oldToken, newToken string) (*models.Deployment, error) {
	db := mustGetSession(ctx)
	err := db.Model(&models.Deployment{}).Where("id = ?", deployment.ID).Where("kube_deploy_token = ?", oldToken).Updates(map[string]interface{}{
		"kube_deploy_token": newToken,
	}).Error
	if err != nil {
		return nil, err
	}
	deployment.KubeDeployToken = newToken
	return deployment, nil
}

func (s *deploymentService) Delete(ctx context.Context, deployment *models.Deployment) (*models.Deployment, error) {
	if deployment.Status != modelschemas.DeploymentStatusTerminated && deployment.Status != modelschemas.DeploymentStatusTerminating {
		return nil, errors.New("deployment is not terminated")
	}
	return deployment, s.getBaseDB(ctx).Unscoped().Delete(deployment).Error
}

func (s *deploymentService) Terminate(ctx context.Context, deployment *models.Deployment) (*models.Deployment, error) {
	deployment, err := s.UpdateStatus(ctx, deployment, UpdateDeploymentStatusOption{
		Status: modelschemas.DeploymentStatusTerminating.Ptr(),
	})
	if err != nil {
		return nil, err
	}
	deploymentRevisions, _, err := DeploymentRevisionService.List(ctx, ListDeploymentRevisionOption{
		BaseListOption: BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		DeploymentId: utils.UintPtr(deployment.ID),
		Status:       modelschemas.DeploymentRevisionStatusActive.Ptr(),
	})
	if err != nil {
		return nil, err
	}
	for _, deploymentRevision := range deploymentRevisions {
		err = DeploymentRevisionService.Terminate(ctx, deploymentRevision)
		if err != nil {
			return nil, err
		}
	}
	_, err = s.SyncStatus(ctx, deployment)
	return deployment, err
}

func (s *deploymentService) GetKubeName(deployment *models.Deployment) string {
	return fmt.Sprintf("yatai-%s", deployment.Name)
}

func (s *deploymentService) GenerateDefaultHostname(ctx context.Context, deployment *models.Deployment) (string, error) {
	cluster, err := ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return "", err
	}
	clientset, _, err := ClusterService.GetKubeCliSet(ctx, cluster)
	if err != nil {
		return "", err
	}
	domainSuffix, err := system.GetDomainSuffix(ctx, clientset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s.%s", deployment.Name, deployment.KubeNamespace, domainSuffix), nil
}

func (s *deploymentService) GetURLs(ctx context.Context, deployment *models.Deployment) ([]string, error) {
	status := modelschemas.DeploymentRevisionStatusActive
	deploymentRevisions, _, err := DeploymentRevisionService.List(ctx, ListDeploymentRevisionOption{
		BaseListOption: BaseListOption{
			Start: utils.UintPtr(0),
			Count: utils.UintPtr(1),
		},
		DeploymentId: utils.UintPtr(deployment.ID),
		Status:       &status,
	})
	if err != nil {
		return nil, err
	}
	if len(deploymentRevisions) == 0 {
		return []string{}, nil
	}
	urls := make([]string, 0)
	kubeName := deployment.Name
	ingCli, err := s.GetKubeIngressesCli(ctx, deployment)
	if err != nil {
		return nil, err
	}
	ing, err := ingCli.Get(ctx, kubeName, metav1.GetOptions{})
	ingIsNotFound := k8serrors.IsNotFound(err)
	if err != nil && !ingIsNotFound {
		return nil, err
	}
	if ingIsNotFound {
		return []string{}, nil
	}
	for _, rule := range ing.Spec.Rules {
		urls = append(urls, fmt.Sprintf("http://%s", rule.Host))
	}
	return urls, nil
}

func (s *deploymentService) GroupByBentoRepositoryIds(ctx context.Context, bentoRepositoryIds []uint, count uint) (map[uint][]*models.Deployment, error) {
	db := mustGetSession(ctx)

	query := db.Raw(`select deployment_target.deployment_id as deployment_id, deployment_target.deployment_revision_id as deployment_revision_id, bento.bento_repository_id as bento_repository_id from deployment_target join bento on deployment_target.bento_id = bento.id where bento.bento_repository_id in (?)`, bentoRepositoryIds)

	type Item struct {
		DeploymentId         uint
		DeploymentRevisionId uint
		BentoRepositoryId    uint
	}

	items := make([]*Item, 0, len(bentoRepositoryIds))
	err := query.Find(&items).Error
	if err != nil {
		return nil, err
	}

	deploymentId2BentoRepositoryId := make(map[uint]uint, len(items))
	for _, item := range items {
		deploymentId2BentoRepositoryId[item.DeploymentId] = item.BentoRepositoryId
	}

	deploymentIds := make([]uint, 0, len(items))
	deploymentIdsSeen := make(map[uint]struct{})

	for _, item := range items {
		if _, ok := deploymentIdsSeen[item.DeploymentId]; !ok {
			deploymentIds = append(deploymentIds, item.DeploymentId)
			deploymentIdsSeen[item.DeploymentId] = struct{}{}
		}
	}

	deploymentRevisions, _, err := DeploymentRevisionService.List(ctx, ListDeploymentRevisionOption{
		DeploymentIds: utils.UintSlicePtr(deploymentIds),
		Status:        modelschemas.DeploymentRevisionStatusPtr(modelschemas.DeploymentRevisionStatusActive),
	})
	if err != nil {
		return nil, err
	}

	activeDeploymentIds := make([]uint, 0, len(deploymentRevisions))
	activeDeploymentIdsSeen := make(map[uint]struct{})

	for _, deploymentRevision := range deploymentRevisions {
		if _, ok := activeDeploymentIdsSeen[deploymentRevision.DeploymentId]; !ok {
			activeDeploymentIds = append(activeDeploymentIds, deploymentRevision.DeploymentId)
			activeDeploymentIdsSeen[deploymentRevision.DeploymentId] = struct{}{}
		}
	}

	deployments, _, err := s.List(ctx, ListDeploymentOption{
		Ids: utils.UintSlicePtr(activeDeploymentIds),
	})
	if err != nil {
		return nil, err
	}

	res := make(map[uint][]*models.Deployment, len(deploymentId2BentoRepositoryId))
	for _, deployment := range deployments {
		bentoRepositoryId, ok := deploymentId2BentoRepositoryId[deployment.ID]
		if !ok {
			continue
		}
		deployments_ := res[bentoRepositoryId]
		if len(deployments_) < int(count) {
			res[bentoRepositoryId] = append(deployments_, deployment)
		}
	}

	return res, nil
}

func (s *deploymentService) CountByBentoRepositoryIds(ctx context.Context, bentoRepositoryIds []uint) (map[uint]uint, error) {
	db := mustGetSession(ctx)

	query := db.Raw(`select deployment_target.deployment_id as deployment_id, deployment_target.deployment_revision_id as deployment_revision_id, bento.bento_repository_id as bento_repository_id from deployment_target join bento on deployment_target.bento_id = bento.id where bento.bento_repository_id in (?)`, bentoRepositoryIds)

	type Item struct {
		DeploymentId         uint
		DeploymentRevisionId uint
		BentoRepositoryId    uint
	}

	items := make([]*Item, 0, len(bentoRepositoryIds))
	err := query.Find(&items).Error
	if err != nil {
		return nil, err
	}

	deploymentRevisionId2BentoRepositoryId := make(map[uint]uint, len(items))
	for _, item := range items {
		deploymentRevisionId2BentoRepositoryId[item.DeploymentRevisionId] = item.BentoRepositoryId
	}

	deploymentIds := make([]uint, 0, len(items))
	deploymentIdsSeen := make(map[uint]struct{})

	for _, item := range items {
		if _, ok := deploymentIdsSeen[item.DeploymentId]; !ok {
			deploymentIds = append(deploymentIds, item.DeploymentId)
			deploymentIdsSeen[item.DeploymentId] = struct{}{}
		}
	}

	deploymentRevisions, _, err := DeploymentRevisionService.List(ctx, ListDeploymentRevisionOption{
		DeploymentIds: utils.UintSlicePtr(deploymentIds),
		Status:        modelschemas.DeploymentRevisionStatusPtr(modelschemas.DeploymentRevisionStatusActive),
	})
	if err != nil {
		return nil, err
	}

	res := make(map[uint]uint, len(deploymentRevisions))
	for _, deploymentRevision := range deploymentRevisions {
		count := res[deploymentRevision.DeploymentId]
		if _, ok := deploymentRevisionId2BentoRepositoryId[deploymentRevision.ID]; ok {
			res[deploymentRevision.DeploymentId] = count + 1
		}
	}

	return res, nil
}
