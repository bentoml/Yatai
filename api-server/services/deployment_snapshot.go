package services

import (
	"context"
	"fmt"

	"github.com/bentoml/yatai/common/sync/errsgroup"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"k8s.io/utils/pointer"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/bentoml/yatai/schemas/modelschemas"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type deploymentSnapshotService struct{}

var DeploymentSnapshotService = deploymentSnapshotService{}

func (s *deploymentSnapshotService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.DeploymentSnapshot{})
}

type CreateDeploymentSnapshotOption struct {
	CreatorId      uint
	DeploymentId   uint
	BentoVersionId uint
	Type           modelschemas.DeploymentSnapshotType
	Status         modelschemas.DeploymentSnapshotStatus
	CanaryRules    *modelschemas.DeploymentSnapshotCanaryRules
	Config         *modelschemas.DeploymentSnapshotConfig
}

type UpdateDeploymentSnapshotOption struct {
	Status *modelschemas.DeploymentSnapshotStatus
}

type ListDeploymentSnapshotOption struct {
	BaseListOption
	DeploymentId uint
	Type         *modelschemas.DeploymentSnapshotType
	Status       *modelschemas.DeploymentSnapshotStatus
}

func (*deploymentSnapshotService) Create(ctx context.Context, opt CreateDeploymentSnapshotOption) (*models.DeploymentSnapshot, error) {
	if opt.Config == nil {
		opt.Config = &modelschemas.DeploymentSnapshotConfig{
			Resources: &modelschemas.DeploymentSnapshotResources{
				Requests: &modelschemas.DeploymentSnapshotResourceItem{
					CPU:    "500m",
					Memory: "1G",
				},
				Limits: &modelschemas.DeploymentSnapshotResourceItem{
					CPU:    "1000m",
					Memory: "2G",
				},
			},
			HPAConf: &modelschemas.DeploymentSnapshotHPAConf{
				CPU:         pointer.Int32(80),
				GPU:         pointer.Int32(80),
				MinReplicas: pointer.Int32(2),
				MaxReplicas: pointer.Int32(10),
			},
		}
	}
	snapshot := models.DeploymentSnapshot{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		DeploymentAssociate: models.DeploymentAssociate{
			DeploymentId: opt.DeploymentId,
		},
		BentoVersionAssociate: models.BentoVersionAssociate{
			BentoVersionId: opt.BentoVersionId,
		},
		Type:        opt.Type,
		Status:      opt.Status,
		CanaryRules: opt.CanaryRules,
		Config:      opt.Config,
	}
	err := mustGetSession(ctx).Create(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, err
}

func (s *deploymentSnapshotService) Update(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot, opt UpdateDeploymentSnapshotOption) (*models.DeploymentSnapshot, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Status != nil {
		updaters["status"] = *opt.Status
		defer func() {
			if err == nil {
				deploymentSnapshot.Status = *opt.Status
			}
		}()
	}

	if len(updaters) == 0 {
		return deploymentSnapshot, nil
	}

	err = s.getBaseDB(ctx).Where("id = ?", deploymentSnapshot.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}

	return deploymentSnapshot, err
}

func (s *deploymentSnapshotService) Get(ctx context.Context, id uint) (*models.DeploymentSnapshot, error) {
	var snapshot models.DeploymentSnapshot
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	if snapshot.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &snapshot, nil
}

func (s *deploymentSnapshotService) GetByUid(ctx context.Context, uid string) (*models.DeploymentSnapshot, error) {
	var snapshot models.DeploymentSnapshot
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	if snapshot.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &snapshot, nil
}

func (s *deploymentSnapshotService) List(ctx context.Context, opt ListDeploymentSnapshotOption) ([]*models.DeploymentSnapshot, uint, error) {
	query := getBaseQuery(ctx, s).Where("deployment_id = ?", opt.DeploymentId)
	if opt.Type != nil {
		query = query.Where("type = ?", *opt.Type)
	}
	if opt.Status != nil {
		query = query.Where("status = ?", *opt.Status)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	deployments := make([]*models.DeploymentSnapshot, 0)
	query = opt.BindQuery(query)
	err = query.Order("id DESC").Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}
	return deployments, uint(total), err
}

func (s *deploymentSnapshotService) GetKubeName(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot) (string, error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentSnapshot)
	if err != nil {
		return "", err
	}
	kubeName := fmt.Sprintf("%s-%s", DeploymentService.GetKubeName(deployment), modelschemas.DeploymentSnapshotTypeAddrs[deploymentSnapshot.Type])
	if deploymentSnapshot.Type == modelschemas.DeploymentSnapshotTypeCanary {
		kubeName = fmt.Sprintf("%s-%d", kubeName, deploymentSnapshot.ID)
	}
	return kubeName, nil
}

func (s *deploymentSnapshotService) GetIngressHost(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot) (string, error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentSnapshot)
	if err != nil {
		return "", err
	}
	return DeploymentService.GetIngressHost(ctx, deployment)
}

func (s *deploymentSnapshotService) GetKubeLabels(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot) (map[string]string, error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentSnapshot)
	if err != nil {
		return nil, err
	}

	labels := map[string]string{
		consts.KubeLabelYataiDeploymentId:           fmt.Sprintf("%d", deployment.ID),
		consts.KubeLabelYataiDeployment:             deployment.Name,
		consts.KubeLabelCreator:                     consts.KubeCreator,
		consts.KubeLabelYataiDeploymentSnapshotType: string(deploymentSnapshot.Type),
		consts.KubeLabelYataiDeployToken:            deployment.KubeDeployToken,
	}
	return labels, nil
}

func (s *deploymentSnapshotService) GetKubeAnnotations(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot) (map[string]string, error) {
	bentoVersion, err := BentoVersionService.GetAssociatedBentoVersion(ctx, deploymentSnapshot)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		consts.KubeAnnotationBentoVersion: bentoVersion.Version,
	}, nil
}

func (s *deploymentSnapshotService) GetKubeOwnerReferenceName(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot) (string, error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentSnapshot)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("yatai-%s-owner-ref-%d", deployment.Name, deploymentSnapshot.ID), nil
}

func (s *deploymentSnapshotService) MakeSureKubeOwnerReferences(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot) ([]metav1.OwnerReference, error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentSnapshot)
	if err != nil {
		return nil, err
	}
	kubeCli, _, err := s.GetKubeCliSet(ctx, deploymentSnapshot)
	if err != nil {
		return nil, err
	}
	cmCli := kubeCli.CoreV1().ConfigMaps(consts.KubeNamespaceYataiDeployment)
	name, err := s.GetKubeOwnerReferenceName(ctx, deploymentSnapshot)
	if err != nil {
		return nil, errors.Wrap(err, "get kube owner reference name")
	}
	cm, err := cmCli.Get(ctx, name, metav1.GetOptions{})
	isNotFound := errors2.IsNotFound(err)
	if err != nil && !isNotFound {
		return nil, errors.Wrapf(err, "get configmap %s", name)
	}
	if isNotFound {
		labels := map[string]string{
			consts.KubeLabelYataiDeployment:     deployment.Name,
			consts.KubeLabelYataiDeploymentId:   fmt.Sprintf("%d", deployment.ID),
			consts.KubeLabelYataiOwnerReference: consts.KubeLabelTrue,
		}
		annotations, err := s.GetKubeAnnotations(ctx, deploymentSnapshot)
		if err != nil {
			return nil, errors.Wrap(err, "get kube annotations")
		}
		annotations[consts.KubeAnnotationYataiDeploymentId] = fmt.Sprintf("%d", deployment.ID)
		cm = &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Labels:      labels,
				Annotations: annotations,
			},
		}
		cm, err = cmCli.Create(ctx, cm, metav1.CreateOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "create configmap %s", name)
		}
	}
	return []metav1.OwnerReference{
		{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Name:       cm.Name,
			UID:        cm.UID,
		},
	}, nil
}

func (s *deploymentSnapshotService) DeleteKubeOwnerReferences(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot) error {
	name, err := s.GetKubeOwnerReferenceName(ctx, deploymentSnapshot)
	if err != nil {
		return errors.Wrap(err, "get kube owner reference name")
	}
	kubeCli, _, err := s.GetKubeCliSet(ctx, deploymentSnapshot)
	if err != nil {
		return err
	}
	cmCli := kubeCli.CoreV1().ConfigMaps(consts.KubeNamespaceYataiDeployment)
	_, err = cmCli.Get(ctx, name, metav1.GetOptions{})
	if errors2.IsNotFound(err) {
		return nil
	}
	err = cmCli.Delete(ctx, name, metav1.DeleteOptions{})
	if errors2.IsNotFound(err) {
		return nil
	}
	return errors.Wrapf(err, "delete configmap %s", name)
}

func (s *deploymentSnapshotService) GetDeployOption(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot, force bool) (*models.DeployOption, error) {
	ownerReferences, err := s.MakeSureKubeOwnerReferences(ctx, deploymentSnapshot)
	if err != nil {
		return nil, err
	}
	deployOption := &models.DeployOption{
		Force:           force,
		OwnerReferences: ownerReferences,
	}
	return deployOption, nil
}

func (s *deploymentSnapshotService) Deploy(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot, force bool) (err error) {
	deploymentSnapshotStatus := modelschemas.DeploymentSnapshotStatusActive
	deploymentSnapshotType := &deploymentSnapshot.Type
	if deploymentSnapshot.Type == modelschemas.DeploymentSnapshotTypeStable {
		deploymentSnapshotType = nil
	}
	oldDeploymentSnapshots, _, err := s.List(ctx, ListDeploymentSnapshotOption{
		BaseListOption: BaseListOption{},
		DeploymentId:   deploymentSnapshot.DeploymentId,
		Type:           deploymentSnapshotType,
		Status:         &deploymentSnapshotStatus,
	})
	if err != nil {
		return
	}

	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentSnapshot)
	if err != nil {
		return
	}

	for _, oldDeploymentSnapshot := range oldDeploymentSnapshots {
		if oldDeploymentSnapshot.ID == deploymentSnapshot.ID {
			continue
		}
		_, err = s.Update(ctx, oldDeploymentSnapshot, UpdateDeploymentSnapshotOption{
			Status: modelschemas.DeploymentSnapshotStatusPtr(modelschemas.DeploymentSnapshotStatusInactive),
		})
		if err != nil {
			err = errors.Wrapf(err, "inactive old %s deployment snapshot", deploymentSnapshot.Type)
			return
		}
	}

	defer func() {
		if err != nil {
			if deploymentSnapshot != nil && deploymentSnapshot.ID != 0 {
				mustGetSession(ctx).Unscoped().Delete(deploymentSnapshot)
			}
			for _, oldDeploymentSnapshot := range oldDeploymentSnapshots {
				if oldDeploymentSnapshot.ID == deploymentSnapshot.ID {
					continue
				}
				_, _ = s.Update(ctx, oldDeploymentSnapshot, UpdateDeploymentSnapshotOption{
					Status: modelschemas.DeploymentSnapshotStatusPtr(modelschemas.DeploymentSnapshotStatusActive),
				})
			}
		} else {
			for _, oldDeploymentSnapshot := range oldDeploymentSnapshots {
				if oldDeploymentSnapshot.ID == deploymentSnapshot.ID {
					continue
				}
				err_ := s.DeleteKubeOwnerReferences(ctx, oldDeploymentSnapshot)
				if err_ != nil {
					logrus.Errorf("deployment %s delete kube owner reference %d failed", deployment.Name, oldDeploymentSnapshot.ID)
				}
			}
		}
	}()

	defer func() {
		go func() {
			_, _ = DeploymentService.SyncStatus(ctx, deployment)
		}()
	}()

	if force {
		gid := xid.New()
		newDeployToken := gid.String()
		oldDeployToken := deployment.KubeDeployToken
		defer func() {
			if err != nil {
				_, _ = DeploymentService.UpdateKubeDeployToken(ctx, deployment, newDeployToken, oldDeployToken)
			}
		}()
		_, err = DeploymentService.UpdateKubeDeployToken(ctx, deployment, oldDeployToken, newDeployToken)
		if err != nil {
			return
		}
	}

	cluster, err := ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return
	}

	_, err = KubeNamespaceService.MakeSureNamespace(ctx, cluster, consts.KubeNamespaceYataiDeployment)
	if err != nil {
		return
	}

	deployOption, err := s.GetDeployOption(ctx, deploymentSnapshot, force)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			return
		}
		status := modelschemas.DeploymentStatusDeploying
		_, _ = DeploymentService.UpdateStatus(ctx, deployment, UpdateDeploymentStatusOption{
			Status: &status,
		})
		go func() {
			_, _ = DeploymentService.SyncStatus(ctx, deployment)
		}()
	}()

	var eg errsgroup.Group

	eg.Go(func() error {
		return KubeDeploymentService.DeployDeploymentSnapshotAsKubeDeployment(ctx, deploymentSnapshot, deployOption)
	})

	eg.Go(func() error {
		return KubeServiceService.DeployDeploymentSnapshotAsKubeService(ctx, deploymentSnapshot, deployOption)
	})

	eg.Go(func() error {
		return KubeIngressService.DeployDeploymentSnapshotAsKubeIngresses(ctx, deploymentSnapshot, deployOption)
	})

	err = eg.Wait()

	return err
}

func (s *deploymentSnapshotService) GetKubeCliSet(ctx context.Context, deploymentSnapshot *models.DeploymentSnapshot) (kubeCli *kubernetes.Clientset, restConfig *rest.Config, err error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentSnapshot)
	if err != nil {
		return
	}
	cluster, err := ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return
	}
	return ClusterService.GetKubeCliSet(ctx, cluster)
}

type IDeploymentSnapshotAssociate interface {
	GetAssociatedDeploymentSnapshotId() uint
	GetAssociatedDeploymentSnapshotCache() *models.DeploymentSnapshot
	SetAssociatedDeploymentSnapshotCache(deployment *models.DeploymentSnapshot)
}

func (s *deploymentSnapshotService) GetAssociatedDeploymentSnapshot(ctx context.Context, associate IDeploymentSnapshotAssociate) (*models.DeploymentSnapshot, error) {
	cache := associate.GetAssociatedDeploymentSnapshotCache()
	if cache != nil {
		return cache, nil
	}
	deployment, err := s.Get(ctx, associate.GetAssociatedDeploymentSnapshotId())
	associate.SetAssociatedDeploymentSnapshotCache(deployment)
	return deployment, err
}
