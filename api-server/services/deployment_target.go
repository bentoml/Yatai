package services

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/common/consts"
)

type deploymentTargetService struct{}

var DeploymentTargetService = deploymentTargetService{}

func (s *deploymentTargetService) getBaseDB(ctx context.Context) *gorm.DB {
	return mustGetSession(ctx).Model(&models.DeploymentTarget{})
}

type CreateDeploymentTargetOption struct {
	CreatorId            uint
	DeploymentId         uint
	DeploymentRevisionId uint
	BentoId              uint
	Type                 modelschemas.DeploymentTargetType
	CanaryRules          *modelschemas.DeploymentTargetCanaryRules
	Config               *modelschemas.DeploymentTargetConfig
}

type UpdateDeploymentTargetOption struct {
	Config **modelschemas.DeploymentTargetConfig
}

type ListDeploymentTargetOption struct {
	BaseListOption
	DeploymentRevisionStatus *modelschemas.DeploymentRevisionStatus
	DeploymentId             *uint
	DeploymentIds            *[]uint
	DeploymentRevisionId     *uint
	DeploymentRevisionIds    *[]uint
	Type                     *modelschemas.DeploymentTargetType
}

func (*deploymentTargetService) Create(ctx context.Context, opt CreateDeploymentTargetOption) (*models.DeploymentTarget, error) {
	if opt.Config == nil {
		opt.Config = &modelschemas.DeploymentTargetConfig{
			Resources: &modelschemas.DeploymentTargetResources{
				Requests: &modelschemas.DeploymentTargetResourceItem{
					CPU:    "500m",
					Memory: "1G",
				},
				Limits: &modelschemas.DeploymentTargetResourceItem{
					CPU:    "1000m",
					Memory: "2G",
				},
			},
			HPAConf: &modelschemas.DeploymentTargetHPAConf{
				CPU:         pointer.Int32Ptr(80),
				GPU:         pointer.Int32Ptr(80),
				MinReplicas: pointer.Int32Ptr(2),
				MaxReplicas: pointer.Int32Ptr(10),
			},
		}
	}
	deploymentTarget := models.DeploymentTarget{
		CreatorAssociate: models.CreatorAssociate{
			CreatorId: opt.CreatorId,
		},
		DeploymentAssociate: models.DeploymentAssociate{
			DeploymentId: opt.DeploymentId,
		},
		DeploymentRevisionAssociate: models.DeploymentRevisionAssociate{
			DeploymentRevisionId: opt.DeploymentRevisionId,
		},
		BentoAssociate: models.BentoAssociate{
			BentoId: opt.BentoId,
		},
		Type:        opt.Type,
		CanaryRules: opt.CanaryRules,
		Config:      opt.Config,
	}
	err := mustGetSession(ctx).Create(&deploymentTarget).Error
	if err != nil {
		return nil, err
	}
	return &deploymentTarget, err
}

func (s *deploymentTargetService) Get(ctx context.Context, id uint) (*models.DeploymentTarget, error) {
	var deploymentTarget models.DeploymentTarget
	err := getBaseQuery(ctx, s).Where("id = ?", id).First(&deploymentTarget).Error
	if err != nil {
		return nil, err
	}
	if deploymentTarget.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &deploymentTarget, nil
}

func (s *deploymentTargetService) GetByUid(ctx context.Context, uid string) (*models.DeploymentTarget, error) {
	var deploymentTarget models.DeploymentTarget
	err := getBaseQuery(ctx, s).Where("uid = ?", uid).First(&deploymentTarget).Error
	if err != nil {
		return nil, err
	}
	if deploymentTarget.ID == 0 {
		return nil, consts.ErrNotFound
	}
	return &deploymentTarget, nil
}

func (s *deploymentTargetService) List(ctx context.Context, opt ListDeploymentTargetOption) ([]*models.DeploymentTarget, uint, error) {
	query := getBaseQuery(ctx, s)
	if opt.DeploymentRevisionStatus != nil {
		query = query.Joins("INNER JOIN deployment_revision ON deployment_revision.id = deployment_target.deployment_revision_id and deployment_revision.status = ?", *opt.DeploymentRevisionStatus)
	}
	if opt.DeploymentId != nil {
		query = query.Where("deployment_target.deployment_id = ?", *opt.DeploymentId)
	}
	if opt.DeploymentRevisionId != nil {
		query = query.Where("deployment_target.deployment_revision_id = ?", *opt.DeploymentRevisionId)
	}
	if opt.DeploymentIds != nil {
		query = query.Where("deployment_target.deployment_id in (?)", *opt.DeploymentIds)
	}
	if opt.DeploymentRevisionIds != nil {
		query = query.Where("deployment_target.deployment_revision_id in (?)", *opt.DeploymentRevisionIds)
	}
	if opt.Type != nil {
		query = query.Where("deployment_target.type = ?", *opt.Type)
	}
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	deploymentTargets := make([]*models.DeploymentTarget, 0)
	query = opt.BindQueryWithLimit(query)
	err = query.Order("deployment_target.id ASC").Find(&deploymentTargets).Error
	if err != nil {
		return nil, 0, err
	}
	return deploymentTargets, uint(total), err
}

func (s *deploymentTargetService) GetKubeName(ctx context.Context, deploymentTarget *models.DeploymentTarget) (string, error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return "", err
	}
	kubeName := fmt.Sprintf("%s-%s", DeploymentService.GetKubeName(deployment), modelschemas.DeploymentTargetTypeAddrs[deploymentTarget.Type])
	if deploymentTarget.Type == modelschemas.DeploymentTargetTypeCanary {
		kubeName = fmt.Sprintf("%s-%d", kubeName, deploymentTarget.ID)
	}
	return kubeName, nil
}

func (s *deploymentTargetService) GenerateIngressHost(ctx context.Context, deploymentTarget *models.DeploymentTarget) (string, error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return "", err
	}
	return DeploymentService.GenerateDefaultHostname(ctx, deployment)
}

func (s *deploymentTargetService) GetKubeLabels(ctx context.Context, deploymentTarget *models.DeploymentTarget) (map[string]string, error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return nil, err
	}

	labels := map[string]string{
		consts.KubeLabelYataiDeploymentId:         fmt.Sprintf("%d", deployment.ID),
		consts.KubeLabelYataiDeployment:           deployment.Name,
		consts.KubeLabelCreator:                   consts.KubeCreator,
		consts.KubeLabelYataiDeploymentTargetType: string(deploymentTarget.Type),
		consts.KubeLabelYataiDeployToken:          deployment.KubeDeployToken,
	}
	return labels, nil
}

func (s *deploymentTargetService) GetKubeAnnotations(ctx context.Context, deploymentTarget *models.DeploymentTarget) (map[string]string, error) {
	bento, err := BentoService.GetAssociatedBento(ctx, deploymentTarget)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		consts.KubeAnnotationBento: bento.Version,
	}, nil
}

func (s *deploymentTargetService) Update(ctx context.Context, b *models.DeploymentTarget, opt UpdateDeploymentTargetOption) (*models.DeploymentTarget, error) {
	var err error
	updaters := make(map[string]interface{})

	if opt.Config != nil {
		updaters["config"] = *opt.Config
		defer func() {
			if err == nil {
				b.Config = *opt.Config
			}
		}()
	}

	if len(updaters) == 0 {
		return b, nil
	}

	err = s.getBaseDB(ctx).Where("id = ?", b.ID).Updates(updaters).Error

	return b, err
}

func (s *deploymentTargetService) Deploy(ctx context.Context, deploymentTarget *models.DeploymentTarget, deployOption *models.DeployOption) (deploymentTarget_ *models.DeploymentTarget, err error) {
	deploymentTarget_ = deploymentTarget

	kubeBentoDeployment, err := KubeBentoDeploymentService.Deploy(ctx, deploymentTarget, deployOption)
	if err != nil {
		err = errors.Wrap(err, "failed to deploy kube bento deployment")
		return
	}

	configUpdated := false
	config := deploymentTarget.Config
	if config == nil {
		configUpdated = true
		config = &modelschemas.DeploymentTargetConfig{
			KubeResourceUid: string(kubeBentoDeployment.UID),
		}
	} else if config.KubeResourceUid == "" {
		configUpdated = true
		config.KubeResourceUid = string(kubeBentoDeployment.UID)
	}

	if configUpdated {
		deploymentTarget_, err = s.Update(ctx, deploymentTarget, UpdateDeploymentTargetOption{
			Config: &config,
		})
		if err != nil {
			err = errors.Wrap(err, "failed to update deployment target config")
			return
		}
	}

	return
}

func (s *deploymentTargetService) GetKubeCliSet(ctx context.Context, deploymentTarget *models.DeploymentTarget) (kubeCli *kubernetes.Clientset, restConfig *rest.Config, err error) {
	deployment, err := DeploymentService.GetAssociatedDeployment(ctx, deploymentTarget)
	if err != nil {
		return
	}
	cluster, err := ClusterService.GetAssociatedCluster(ctx, deployment)
	if err != nil {
		return
	}
	return ClusterService.GetKubeCliSet(ctx, cluster)
}
