package controllersv1

import (
	"context"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type bundleController struct {
	baseController
}

var BundleController = bundleController{}

type GetBundleSchema struct {
	GetClusterSchema
	BundleName string `path:"bundleName"`
}

func (s *GetBundleSchema) GetBundle(ctx context.Context) (*models.Bundle, error) {
	cluster, err := s.GetCluster(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "get cluster %s", cluster.Name)
	}
	bundle, err := services.BundleService.GetByName(ctx, cluster.ID, s.BundleName)
	if err != nil {
		return nil, errors.Wrapf(err, "get bundle %s", s.BundleName)
	}
	return bundle, nil
}

func (c *bundleController) canView(ctx context.Context, bundle *models.Bundle) error {
	cluster, err := services.ClusterService.GetAssociatedCluster(ctx, bundle)
	if err != nil {
		return errors.Wrap(err, "get associated cluster")
	}
	return ClusterController.canView(ctx, cluster)
}

func (c *bundleController) canUpdate(ctx context.Context, bundle *models.Bundle) error {
	cluster, err := services.ClusterService.GetAssociatedCluster(ctx, bundle)
	if err != nil {
		return errors.Wrap(err, "get associated cluster")
	}
	return ClusterController.canUpdate(ctx, cluster)
}

func (c *bundleController) canOperate(ctx context.Context, bundle *models.Bundle) error {
	cluster, err := services.ClusterService.GetAssociatedCluster(ctx, bundle)
	if err != nil {
		return errors.Wrap(err, "get associated cluster")
	}
	return ClusterController.canOperate(ctx, cluster)
}

type CreateBundleSchema struct {
	schemasv1.CreateBundleSchema
	GetClusterSchema
}

func (c *bundleController) Create(ctx *gin.Context, schema *CreateBundleSchema) (*schemasv1.BundleSchema, error) {
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

	bundle, err := services.BundleService.Create(ctx, services.CreateBundleOption{
		CreatorId: user.ID,
		ClusterId: cluster.ID,
		Name:      schema.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create bundle")
	}
	return transformersv1.ToBundleSchema(ctx, bundle)
}

type UpdateBundleSchema struct {
	schemasv1.UpdateBundleSchema
	GetBundleSchema
}

func (c *bundleController) Update(ctx *gin.Context, schema *UpdateBundleSchema) (*schemasv1.BundleSchema, error) {
	bundle, err := schema.GetBundle(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canUpdate(ctx, bundle); err != nil {
		return nil, err
	}
	bundle, err = services.BundleService.Update(ctx, bundle, services.UpdateBundleOption{
		Description: schema.Description,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update bundle")
	}
	return transformersv1.ToBundleSchema(ctx, bundle)
}

func (c *bundleController) Get(ctx *gin.Context, schema *GetBundleSchema) (*schemasv1.BundleSchema, error) {
	bundle, err := schema.GetBundle(ctx)
	if err != nil {
		return nil, err
	}
	if err = c.canView(ctx, bundle); err != nil {
		return nil, err
	}
	return transformersv1.ToBundleSchema(ctx, bundle)
}

type ListBundleSchema struct {
	schemasv1.ListQuerySchema
	GetClusterSchema
}

func (c *bundleController) List(ctx *gin.Context, schema *ListBundleSchema) (*schemasv1.BundleListSchema, error) {
	cluster, err := schema.GetCluster(ctx)
	if err != nil {
		return nil, err
	}

	if err = ClusterController.canView(ctx, cluster); err != nil {
		return nil, err
	}

	bundles, total, err := services.BundleService.List(ctx, services.ListBundleOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		ClusterId: utils.UintPtr(cluster.ID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "list bundles")
	}

	bundleSchemas, err := transformersv1.ToBundleSchemas(ctx, bundles)
	return &schemasv1.BundleListSchema{
		BaseListSchema: schemasv1.BaseListSchema{
			Total: total,
			Start: schema.Start,
			Count: schema.Count,
		},
		Items: bundleSchemas,
	}, err
}
