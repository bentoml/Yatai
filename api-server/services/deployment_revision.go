package services

import (
	"context"
	"fmt"

	"github.com/bentoml/yatai/common/utils"

	"github.com/bentoml/yatai/common/sync/errsgroup"

	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
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

type deploymentRevisionService struct{}

var DeploymentRevisionService = deploymentRevisionService{}

func (s *deploymentRevisionService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.DeploymentRevision{})
}

type CreateDeploymentRevisionOption struct {
	CreatorId      uint
	DeploymentId   uint
	BentoVersionId uint
	Status         modelschemas.DeploymentRevisionStatus
}

type UpdateDeploymentRevisionOption struct {
	Status *modelschemas.DeploymentRevisionStatus
}

type ListDeploymentRevisionOption struct {
	BaseListOption
	DeploymentId  *uint
	DeploymentIds *[]uint
	Status        *modelschemas.DeploymentRevisionStatus
}

func (*deploymentRevisionService) Create(ctx context.Context, opt CreateDeploymentRevisionOption) (*models.DeploymentRevision, error) {
	deploymentRevision := models.DeploymentRevision{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		DeploymentAssociate: models.DeploymentAssociate{
			DeploymentId: opt.DeploymentId,
		},
		Status: opt.Status,
	}
	err := mustGetSession(ctx).Create(&deploymentRevision).Error
	if err != nil {
		return nil, err
	}
	return &deploymentRevision, err
}

func (s *deploymentRevisionService) Update(ctx context.Context, deploymentRevision *models.DeploymentRevision, opt UpdateDeploymentRevisionOption) (*models.DeploymentRevision, error) {
	var err error
	updaters := make(map[string]interface{})
	if opt.Status != nil {
		updaters["status"] = *opt.Status
		defer func() {
			if err == nil {
				deploymentRevision.Status = *opt.Status
			}
		}()
	}

	if len(updaters) == 0 {
		return deploymentRevision, nil
	}

	err = s.getBaseDB(ctx).Where("id = ?", deploymentRevision.ID).Updates(updaters).Error
	if err != nil {
		return nil, err
	}

	return deploymentRevision, err
}

func (s *deploymentRevisionService) Get(ctx context.Context, id uint) (*models.DeploymentRevision, error) {
	var deploymentRevision models.DeploymentRevision
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&deploymentRevision).Error
	if err != nil {
		return nil, err
	}
	if deploymentRevision.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &deploymentRevision, nil
}

func (s *deploymentRevisionService) GetByUid(ctx context.Context, uid string) (*models.DeploymentRevision, error) {
	var deploymentRevision models.DeploymentRevision
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&deploymentRevision).Error
	if err != nil {
		return nil, err
	}
	if deploymentRevision.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &deploymentRevision, nil
}

func (s *deploymentRevisionService) List(ctx context.Context, opt ListDeploymentRevisionOption) ([]*models.DeploymentRevision, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.DeploymentId != nil {
		query = query.Where("deployment_id = ?", *opt.DeploymentId)
	}
	if opt.DeploymentIds != nil {
		query = query.Where("deployment_id in (?)", *opt.DeploymentIds)
	}
	if opt.Status != nil {
		query = query.Where("status = ?", *opt.Status)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	deployments := make([]*models.DeploymentRevision, 0)
	query = opt.BindQueryWithLimit(query)
	err = query.Order("id DESC").Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}
	return deployments, uint(total), err
}

func (s *deploymentRevisionService) GenerateIngressHost(ctx context.Context, deploymentRevision *models.DeploymentRevision) (string, error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentRevision)
	if err != nil {
		return "", err
	}
	return DeploymentService.GenerateDefaultHostname(ctx, deployment)
}

func (s *deploymentRevisionService) GetKubeOwnerReferenceName(ctx context.Context, deploymentRevision *models.DeploymentRevision) (string, error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentRevision)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("yatai-%s-owner-ref-%d", deployment.Name, deploymentRevision.ID), nil
}

func (s *deploymentRevisionService) MakeSureKubeOwnerReferences(ctx context.Context, deploymentRevision *models.DeploymentRevision) ([]metav1.OwnerReference, error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentRevision)
	if err != nil {
		return nil, err
	}
	kubeCli, _, err := s.GetKubeCliSet(ctx, deploymentRevision)
	if err != nil {
		return nil, err
	}
	kubeNs := DeploymentService.GetKubeNamespace(deployment)
	cmCli := kubeCli.CoreV1().ConfigMaps(kubeNs)
	name, err := s.GetKubeOwnerReferenceName(ctx, deploymentRevision)
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
		cm = &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
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

func (s *deploymentRevisionService) DeleteKubeOwnerReferences(ctx context.Context, deploymentRevision *models.DeploymentRevision) error {
	name, err := s.GetKubeOwnerReferenceName(ctx, deploymentRevision)
	if err != nil {
		return errors.Wrap(err, "get kube owner reference name")
	}
	kubeCli, _, err := s.GetKubeCliSet(ctx, deploymentRevision)
	if err != nil {
		return err
	}
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentRevision)
	if err != nil {
		return err
	}
	kubeNs := DeploymentService.GetKubeNamespace(deployment)
	cmCli := kubeCli.CoreV1().ConfigMaps(kubeNs)
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

func (s *deploymentRevisionService) GetDeployOption(ctx context.Context, deploymentRevision *models.DeploymentRevision, force bool) (*models.DeployOption, error) {
	ownerReferences, err := s.MakeSureKubeOwnerReferences(ctx, deploymentRevision)
	if err != nil {
		return nil, err
	}
	deployOption := &models.DeployOption{
		Force:           force,
		OwnerReferences: ownerReferences,
	}
	return deployOption, nil
}

func (s *deploymentRevisionService) Terminate(ctx context.Context, deploymentRevision *models.DeploymentRevision) (err error) {
	return s.DeleteKubeOwnerReferences(ctx, deploymentRevision)
}

func (s *deploymentRevisionService) Deploy(ctx context.Context, deploymentRevision *models.DeploymentRevision, deploymentTargets []*models.DeploymentTarget, force bool) (err error) {
	deploymentRevisionStatus := modelschemas.DeploymentRevisionStatusActive
	oldDeploymentRevisions, _, err := s.List(ctx, ListDeploymentRevisionOption{
		BaseListOption: BaseListOption{},
		DeploymentId:   utils.UintPtr(deploymentRevision.DeploymentId),
		Status:         &deploymentRevisionStatus,
	})
	if err != nil {
		return
	}

	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentRevision)
	if err != nil {
		return
	}

	for _, oldDeploymentRevision := range oldDeploymentRevisions {
		if oldDeploymentRevision.ID == deploymentRevision.ID {
			continue
		}
		_, err = s.Update(ctx, oldDeploymentRevision, UpdateDeploymentRevisionOption{
			Status: modelschemas.DeploymentRevisionStatusPtr(modelschemas.DeploymentRevisionStatusInactive),
		})
		if err != nil {
			return
		}
	}

	defer func() {
		if err != nil {
			if deploymentRevision != nil && deploymentRevision.ID != 0 {
				mustGetSession(ctx).Unscoped().Delete(deploymentRevision)
			}
			for _, oldDeploymentRevision := range oldDeploymentRevisions {
				if oldDeploymentRevision.ID == deploymentRevision.ID {
					continue
				}
				_, _ = s.Update(ctx, oldDeploymentRevision, UpdateDeploymentRevisionOption{
					Status: modelschemas.DeploymentRevisionStatusPtr(modelschemas.DeploymentRevisionStatusActive),
				})
			}
		} else {
			for _, oldDeploymentRevision := range oldDeploymentRevisions {
				if oldDeploymentRevision.ID == deploymentRevision.ID {
					continue
				}
				err_ := s.DeleteKubeOwnerReferences(ctx, oldDeploymentRevision)
				if err_ != nil {
					logrus.Errorf("deployment %s delete kube owner reference %d failed", deployment.Name, oldDeploymentRevision.ID)
				}
			}
		}
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

	kubeNs := DeploymentService.GetKubeNamespace(deployment)

	_, err = KubeNamespaceService.MakeSureNamespace(ctx, cluster, kubeNs)
	if err != nil {
		return
	}

	deployOption, err := s.GetDeployOption(ctx, deploymentRevision, force)
	if err != nil {
		return
	}

	if len(deploymentTargets) == 0 {
		deploymentTargets, _, err = DeploymentTargetService.List(ctx, ListDeploymentTargetOption{
			DeploymentRevisionId: utils.UintPtr(deploymentRevision.ID),
		})
		if err != nil {
			return
		}
	}

	defer func() {
		if err != nil {
			return
		}
		status := modelschemas.DeploymentStatusDeploying
		_, _ = DeploymentService.UpdateStatus(ctx, deployment, UpdateDeploymentStatusOption{
			Status: &status,
		})
		deployment.Status = status
		go func() {
			_, _ = DeploymentService.SyncStatus(ctx, deployment)
		}()
	}()

	var eg errsgroup.Group

	for _, deploymentTarget := range deploymentTargets {
		deploymentTarget := deploymentTarget
		eg.Go(func() error {
			return DeploymentTargetService.Deploy(ctx, deploymentTarget, deployOption)
		})
	}

	err = eg.Wait()

	return err
}

func (s *deploymentRevisionService) GetKubeCliSet(ctx context.Context, deploymentRevision *models.DeploymentRevision) (kubeCli *kubernetes.Clientset, restConfig *rest.Config, err error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentRevision)
	if err != nil {
		return
	}
	cluster, err := ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return
	}
	return ClusterService.GetKubeCliSet(ctx, cluster)
}

type IDeploymentRevisionAssociate interface {
	GetAssociatedDeploymentRevisionId() uint
	GetAssociatedDeploymentRevisionCache() *models.DeploymentRevision
	SetAssociatedDeploymentRevisionCache(deployment *models.DeploymentRevision)
}

func (s *deploymentRevisionService) GetAssociatedDeploymentRevision(ctx context.Context, associate IDeploymentRevisionAssociate) (*models.DeploymentRevision, error) {
	cache := associate.GetAssociatedDeploymentRevisionCache()
	if cache != nil {
		return cache, nil
	}
	deployment, err := s.Get(ctx, associate.GetAssociatedDeploymentRevisionId())
	associate.SetAssociatedDeploymentRevisionCache(deployment)
	return deployment, err
}
