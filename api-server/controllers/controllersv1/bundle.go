package controllersv1

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bentoml/yatai/api-server/models"
	"github.com/bentoml/yatai/api-server/services"
	"github.com/bentoml/yatai/api-server/transformers/transformersv1"
	"github.com/bentoml/yatai/common/utils"
	"github.com/bentoml/yatai/schemas/schemasv1"
)

type bundleController struct {
	baseController
}

var BundleController = bundleController{}

type GetBundleSchema struct {
	GetOrganizationSchema
	BundleName string `path:"bundleName"`
}

func (s *GetBundleSchema) GetBundle(ctx context.Context) (*models.Bundle, error) {
	organization, err := s.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "get organization %s", organization.Name)
	}
	bundle, err := services.BundleService.GetByName(ctx, organization.ID, s.BundleName)
	if err != nil {
		return nil, errors.Wrapf(err, "get bundle %s", s.BundleName)
	}
	return bundle, nil
}

func (c *bundleController) canView(ctx context.Context, bundle *models.Bundle) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, bundle)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canView(ctx, organization)
}

func (c *bundleController) canUpdate(ctx context.Context, bundle *models.Bundle) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, bundle)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canUpdate(ctx, organization)
}

func (c *bundleController) canOperate(ctx context.Context, bundle *models.Bundle) error {
	organization, err := services.OrganizationService.GetAssociatedOrganization(ctx, bundle)
	if err != nil {
		return errors.Wrap(err, "get associated organization")
	}
	return OrganizationController.canOperate(ctx, organization)
}

type CreateBundleSchema struct {
	schemasv1.CreateBundleSchema
	GetOrganizationSchema
}

func (c *bundleController) Create(ctx *gin.Context, schema *CreateBundleSchema) (*schemasv1.BundleSchema, error) {
	user, err := services.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canUpdate(ctx, organization); err != nil {
		return nil, err
	}

	bundle, err := services.BundleService.Create(ctx, services.CreateBundleOption{
		CreatorId:      user.ID,
		OrganizationId: organization.ID,
		Name:           schema.Name,
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
	GetOrganizationSchema
}

func (c *bundleController) List(ctx *gin.Context, schema *ListBundleSchema) (*schemasv1.BundleListSchema, error) {
	organization, err := schema.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}

	if err = OrganizationController.canView(ctx, organization); err != nil {
		return nil, err
	}

	bundles, total, err := services.BundleService.List(ctx, services.ListBundleOption{
		BaseListOption: services.BaseListOption{
			Start:  utils.UintPtr(schema.Start),
			Count:  utils.UintPtr(schema.Count),
			Search: schema.Search,
		},
		OrganizationId: utils.UintPtr(organization.ID),
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
