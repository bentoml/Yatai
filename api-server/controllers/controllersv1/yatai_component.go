package controllersv1

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai-schemas/modelschemas"
	"github.com/bentoml/yatai-schemas/schemasv1"
	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
)

type yataiComponentController struct {
	baseController
}

var YataiComponentController = yataiComponentController{}

type GetYataiComponentSchema struct {
	GetClusterSchema
	YataiComponentName string `path:"yataiComponentName"`
}

func (s *GetYataiComponentSchema) GetYataiComponent(ctx context.Context) (*models.YataiComponent, error) {
	cluster, err := s.GetCluster(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster")
	}
	yataiComponent, err := services.YataiComponentService.GetByName(ctx, cluster.ID, s.YataiComponentName)
	if err != nil {
		return nil, errors.Wrapf(err, "get yataiComponent %s", s.YataiComponentName)
	}
	return yataiComponent, nil
}

func (c *yataiComponentController) canView(ctx context.Context, yataiComponent *models.YataiComponent) error {
	cluster, err := services.ClusterService.GetAssociatedCluster(ctx, yataiComponent)
	if err != nil {
		return errors.Wrap(err, "get associated cluster")
	}
	return ClusterController.canView(ctx, cluster)
}

func (c *yataiComponentController) canUpdate(ctx context.Context, yataiComponent *models.YataiComponent) error {
	cluster, err := services.ClusterService.GetAssociatedCluster(ctx, yataiComponent)
	if err != nil {
		return errors.Wrap(err, "get associated cluster")
	}
	return ClusterController.canUpdate(ctx, cluster)
}

func (c *yataiComponentController) canOperate(ctx context.Context, yataiComponent *models.YataiComponent) error {
	cluster, err := services.ClusterService.GetAssociatedCluster(ctx, yataiComponent)
	if err != nil {
		return errors.Wrap(err, "get associated cluster")
	}
	return ClusterController.canOperate(ctx, cluster)
}

type RegisterYataiComponentSchema struct {
	schemasv1.RegisterYataiComponentSchema
	GetClusterSchema
}

func (c *yataiComponentController) Register(ctx *gin.Context, schema *RegisterYataiComponentSchema) (*schemasv1.YataiComponentSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}
	if err = ClusterController.canUpdate(ctx, cluster); err != nil {
		return nil, err
	}

	kubeNamespace := strings.TrimSpace(schema.KubeNamespace)

	// nolint: ineffassign, staticcheck
	_, ctx_, df, err := services.StartTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { df(err) }()

	yataiComponent, err := services.YataiComponentService.GetByName(ctx_, cluster.ID, string(schema.Name))
	isNotFound := utils.IsNotFound(err)
	if err != nil && !isNotFound {
		return nil, errors.Wrap(err, "get yataiComponent")
	}

	if isNotFound {
		yataiComponent, err = services.YataiComponentService.Create(ctx_, services.CreateYataiComponentOption{
			CreatorId:      user.ID,
			OrganizationId: cluster.OrganizationId,
			ClusterId:      cluster.ID,
			Name:           string(schema.Name),
			KubeNamespace:  kubeNamespace,
			Version:        schema.Version,
			Manifest: &modelschemas.YataiComponentManifestSchema{
				SelectorLabels: schema.SelectorLabels,
			},
		})
	} else {
		manifest := &modelschemas.YataiComponentManifestSchema{
			SelectorLabels: schema.SelectorLabels,
		}
		now := time.Now()
		now_ := &now
		opt := services.UpdateYataiComponentOption{
			LatestHeartbeatAt: &now_,
		}
		if yataiComponent.Version != schema.Version {
			opt.Version = &schema.Version
			opt.LatestInstalledAt = &now_
			opt.Manifest = &manifest
		}
		yataiComponent, err = services.YataiComponentService.Update(ctx_, yataiComponent, opt)
	}

	if err != nil {
		return nil, errors.Wrap(err, "register yataiComponent")
	}

	yataiComponentSchema, err := transformersv1.ToYataiComponentSchema(ctx_, yataiComponent)
	return yataiComponentSchema, err
}

func (c *yataiComponentController) ListAll(ctx *gin.Context, schema *GetOrganizationSchema) ([]*schemasv1.YataiComponentSchema, error) {
	org, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	if err = OrganizationController.canView(ctx, org); err != nil {
		return nil, err
	}

	yataiComponents, err := services.YataiComponentService.List(ctx, services.ListYataiComponentOption{
		OrganizationId: &org.ID,
	})
	if err != nil {
		return nil, err
	}

	return transformersv1.ToYataiComponentSchemas(ctx, yataiComponents)
}

func (c *yataiComponentController) List(ctx *gin.Context, schema *GetClusterSchema) ([]*schemasv1.YataiComponentSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}
	if err = ClusterController.canView(ctx, cluster); err != nil {
		return nil, err
	}

	yataiComponents, err := services.YataiComponentService.List(ctx, services.ListYataiComponentOption{
		ClusterId: &cluster.ID,
	})
	if err != nil {
		return nil, err
	}

	return transformersv1.ToYataiComponentSchemas(ctx, yataiComponents)
}
